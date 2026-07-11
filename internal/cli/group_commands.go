package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DishanRajapaksha/logix-cli/internal/config"
	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
	"github.com/DishanRajapaksha/logix-cli/internal/output"
)

type groupSummary struct {
	Name        string   `json:"name"`
	Points      []string `json:"points"`
	Description string   `json:"description,omitempty"`
}

type pointWriteResult struct {
	Group  string `json:"group"`
	Point  string `json:"point"`
	Tag    string `json:"tag"`
	Type   string `json:"type"`
	Value  any    `json:"value"`
	Status string `json:"status"`
}

type pointValueSpec struct {
	Point config.Point
	Value any
}

func (a *App) groups(args []string) error {
	fs := a.newFlagSet("groups")
	flags := addCommonFlags(fs, true)
	filter := fs.String("filter", "", "case-insensitive group name filter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("groups takes no positional arguments")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	cfg, err := loadPointConfig(flags)
	if err != nil {
		return err
	}
	declared := cfg.NormalisedGroups()
	results := make([]groupSummary, 0, len(declared))
	rows := make([][]string, 0, len(declared))
	for _, group := range declared {
		if *filter != "" && !strings.Contains(strings.ToLower(group.Name), strings.ToLower(*filter)) {
			continue
		}
		result := groupSummary{Name: group.Name, Points: group.Points, Description: group.Description}
		results = append(results, result)
		rows = append(rows, []string{result.Name, strconv.Itoa(len(result.Points)), strings.Join(result.Points, ","), result.Description})
	}
	return output.Snapshot(a.out, flags.format, []string{"NAME", "POINT_COUNT", "POINTS", "DESCRIPTION"}, rows, results)
}

func (a *App) readGroup(args []string) error {
	fs := a.newFlagSet("read-group")
	flags := addCommonFlags(fs, true)
	name, err := parseTagArgs(fs, args, "read-group")
	if err != nil {
		return err
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	group, points, err := configuredGroup(flags, name)
	if err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	results := make([]pointSample, 0, len(points))
	rows := make([][]string, 0, len(points))
	for _, point := range points {
		value, actualType, err := client.Read(point.Tag, point.Type, point.Elements)
		if err != nil {
			return err
		}
		result := pointSample{Timestamp: timestamp, Group: group.Name, Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}
		results = append(results, result)
		rows = append(rows, []string{result.Timestamp, result.Group, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit})
	}
	return output.Snapshot(a.out, flags.format, []string{"TIMESTAMP", "GROUP", "POINT", "TAG", "TYPE", "VALUE", "UNIT"}, rows, results)
}

func (a *App) writeGroup(args []string) error {
	fs := a.newFlagSet("write-group")
	flags := addCommonFlags(fs, true)
	var rawSets repeatedValue
	fs.Var(&rawSets, "set", "repeat POINT=VALUE")
	yes := fs.Bool("yes", false, "transmit all writes")
	dryRun := fs.Bool("dry-run", false, "explicitly keep all writes as a dry run")
	name, err := parseTagArgs(fs, args, "write-group")
	if err != nil {
		return err
	}
	if len(rawSets) == 0 {
		return fmt.Errorf("write-group requires at least one --set")
	}
	if err := validateWriteMode(*yes, *dryRun); err != nil {
		return err
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	group, points, err := configuredGroup(flags, name)
	if err != nil {
		return err
	}
	specs, err := parsePointValueSpecs(group, points, rawSets)
	if err != nil {
		return err
	}
	options, _, err := a.options(flags)
	if err != nil {
		return err
	}
	status := "dry-run"
	if *yes {
		client, err := a.factory.New(options)
		if err != nil {
			return err
		}
		if err := client.Connect(); err != nil {
			return err
		}
		defer closeClient(client)
		for i, spec := range specs {
			if err := client.Write(spec.Point.Tag, spec.Value); err != nil {
				return fmt.Errorf("write-group stopped after %d successful writes: %w", i, err)
			}
		}
		status = "written"
	}
	results := make([]pointWriteResult, 0, len(specs))
	rows := make([][]string, 0, len(specs))
	for _, spec := range specs {
		result := pointWriteResult{Group: group.Name, Point: spec.Point.Name, Tag: spec.Point.Tag, Type: spec.Point.Type, Value: spec.Value, Status: status}
		results = append(results, result)
		rows = append(rows, []string{result.Group, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Status})
	}
	return output.Snapshot(a.out, flags.format, []string{"GROUP", "POINT", "TAG", "TYPE", "VALUE", "STATUS"}, rows, results)
}

func (a *App) watchGroup(args []string) error {
	fs := a.newFlagSet("watch-group")
	flags := addCommonFlags(fs, false)
	interval := fs.Duration("interval", time.Second, "poll interval")
	count := fs.Int("count", 0, "number of poll cycles; 0 means unlimited")
	duration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")
	name, err := parseTagArgs(fs, args, "watch-group")
	if err != nil {
		return err
	}
	if err := validateWatchOptions(*interval, *count, *duration); err != nil {
		return err
	}
	group, points, err := configuredGroup(flags, name)
	if err != nil {
		return err
	}
	stream, err := output.NewStream(a.out, flags.format, []string{"timestamp", "group", "point", "tag", "type", "value", "unit"})
	if err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	started := time.Now()
	cycles := 0
	for {
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		for _, point := range points {
			value, actualType, err := client.Read(point.Tag, point.Type, point.Elements)
			if err != nil {
				return err
			}
			result := pointSample{Timestamp: timestamp, Group: group.Name, Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}
			row := []string{result.Timestamp, result.Group, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit}
			if err := stream.Write(row, result); err != nil {
				return err
			}
		}
		cycles++
		if watchShouldStop(started, cycles, *count, *duration) {
			return nil
		}
		if !waitForNextWatch(started, *duration, *interval) {
			return nil
		}
	}
}

func configuredGroup(flags *commonFlags, name string) (config.PointGroup, []config.Point, error) {
	cfg, err := loadPointConfig(flags)
	if err != nil {
		return config.PointGroup{}, nil, err
	}
	group, points, err := cfg.PointsForGroup(strings.TrimSpace(name))
	if err != nil {
		return config.PointGroup{}, nil, err
	}
	return group, points, nil
}

func parsePointValueSpecs(group config.PointGroup, points []config.Point, values []string) ([]pointValueSpec, error) {
	allowed := make(map[string]config.Point, len(points))
	for _, point := range points {
		allowed[strings.ToLower(point.Name)] = point
	}
	seen := make(map[string]struct{}, len(values))
	specs := make([]pointValueSpec, 0, len(values))
	for _, raw := range values {
		equals := strings.Index(raw, "=")
		if equals <= 0 {
			return nil, fmt.Errorf("%w: group set %q must be POINT=VALUE", logixclient.ErrValidation, raw)
		}
		name := strings.TrimSpace(raw[:equals])
		point, ok := allowed[strings.ToLower(name)]
		if !ok {
			return nil, fmt.Errorf("%w: point %q is not in group %q", logixclient.ErrValidation, name, group.Name)
		}
		key := strings.ToLower(point.Name)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("%w: point %q is set more than once", logixclient.ErrValidation, point.Name)
		}
		seen[key] = struct{}{}
		if !point.Writable {
			return nil, fmt.Errorf("%w: point %q is not writable", errWriteRejected, point.Name)
		}
		value, err := logixclient.ParseValue(point.Type, raw[equals+1:])
		if err != nil {
			return nil, err
		}
		specs = append(specs, pointValueSpec{Point: point, Value: value})
	}
	return specs, nil
}
