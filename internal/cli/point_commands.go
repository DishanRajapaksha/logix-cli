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

type pointSample struct {
	Timestamp string `json:"timestamp"`
	Group     string `json:"group,omitempty"`
	Point     string `json:"point"`
	Tag       string `json:"tag"`
	Type      string `json:"type"`
	Value     any    `json:"value"`
	Unit      string `json:"unit,omitempty"`
}

func (a *App) points(args []string) error {
	fs := a.newFlagSet("points")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("points takes no positional arguments")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	cfg, err := loadPointConfig(flags)
	if err != nil {
		return err
	}
	points := cfg.NormalisedPoints()
	rows := make([][]string, 0, len(points))
	for _, point := range points {
		rows = append(rows, []string{point.Name, point.Tag, point.Type, strconv.FormatUint(uint64(point.Elements), 10), point.Unit, strconv.FormatBool(point.Writable), point.Description})
	}
	return output.Snapshot(a.out, flags.format, []string{"NAME", "TAG", "TYPE", "ELEMENTS", "UNIT", "WRITABLE", "DESCRIPTION"}, rows, points)
}

func (a *App) readPoint(args []string) error {
	fs := a.newFlagSet("read-point")
	flags := addCommonFlags(fs, true)
	name, err := parseTagArgs(fs, args, "read-point")
	if err != nil {
		return err
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	point, err := configuredPoint(flags, name)
	if err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	value, actualType, err := client.Read(point.Tag, point.Type, point.Elements)
	if err != nil {
		return err
	}
	result := pointSample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}
	row := []string{result.Timestamp, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit}
	return output.Snapshot(a.out, flags.format, []string{"TIMESTAMP", "POINT", "TAG", "TYPE", "VALUE", "UNIT"}, [][]string{row}, result)
}

func (a *App) writePoint(args []string) error {
	fs := a.newFlagSet("write-point")
	flags := addCommonFlags(fs, true)
	var rawValue string
	valueSet := false
	fs.Func("value", "value to write", func(value string) error {
		rawValue = value
		valueSet = true
		return nil
	})
	yes := fs.Bool("yes", false, "transmit the write")
	dryRun := fs.Bool("dry-run", false, "explicitly keep the write as a dry run")
	name, err := parseTagArgs(fs, args, "write-point")
	if err != nil {
		return err
	}
	if err := validateWriteMode(*yes, *dryRun); err != nil {
		return err
	}
	if !valueSet {
		return fmt.Errorf("write-point requires --value")
	}
	point, err := configuredPoint(flags, name)
	if err != nil {
		return err
	}
	if !point.Writable {
		return fmt.Errorf("%w: point %q is not writable", errWriteRejected, point.Name)
	}
	value, err := logixclient.ParseValue(point.Type, rawValue)
	if err != nil {
		return err
	}
	options, _, err := a.options(flags)
	if err != nil {
		return err
	}
	if !*yes {
		fmt.Fprintf(a.out, "dry-run: would write controller=%s:%d point=%s tag=%s type=%s value=%v; add --yes to transmit\n", options.Address, options.Port, point.Name, point.Tag, point.Type, value)
		return nil
	}
	client, err := a.factory.New(options)
	if err != nil {
		return err
	}
	if err := client.Connect(); err != nil {
		return err
	}
	defer closeClient(client)
	if err := client.Write(point.Tag, value); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "wrote point=%s tag=%s type=%s value=%v\n", point.Name, point.Tag, point.Type, value)
	return nil
}

func (a *App) watchPoint(args []string) error {
	fs := a.newFlagSet("watch-point")
	flags := addCommonFlags(fs, false)
	interval := fs.Duration("interval", time.Second, "poll interval")
	count := fs.Int("count", 0, "number of samples; 0 means unlimited")
	duration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")
	name, err := parseTagArgs(fs, args, "watch-point")
	if err != nil {
		return err
	}
	if err := validateWatchOptions(*interval, *count, *duration); err != nil {
		return err
	}
	point, err := configuredPoint(flags, name)
	if err != nil {
		return err
	}
	stream, err := output.NewStream(a.out, flags.format, []string{"timestamp", "point", "tag", "type", "value", "unit"})
	if err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	started := time.Now()
	written := 0
	for {
		value, actualType, err := client.Read(point.Tag, point.Type, point.Elements)
		if err != nil {
			return err
		}
		result := pointSample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}
		row := []string{result.Timestamp, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit}
		if err := stream.Write(row, result); err != nil {
			return err
		}
		written++
		if watchShouldStop(started, written, *count, *duration) {
			return nil
		}
		if !waitForNextWatch(started, *duration, *interval) {
			return nil
		}
	}
}

func loadPointConfig(flags *commonFlags) (config.Config, error) {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return config.Config{}, err
	}
	if _, _, err := cfg.Profile(flags.profile); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func configuredPoint(flags *commonFlags, name string) (config.Point, error) {
	cfg, err := loadPointConfig(flags)
	if err != nil {
		return config.Point{}, err
	}
	point, err := cfg.Point(strings.TrimSpace(name))
	if err != nil {
		return config.Point{}, err
	}
	return point, nil
}
