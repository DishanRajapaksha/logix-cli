package cli

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
	"github.com/DishanRajapaksha/logix-cli/internal/output"
)

type sample struct {
	Timestamp string `json:"timestamp"`
	Tag       string `json:"tag"`
	Type      string `json:"type"`
	Value     any    `json:"value"`
}

func (a *App) testConnection(args []string) error {
	fs := a.newFlagSet("test-connection")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("test-connection takes no positional arguments")
	}
	client, options, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	fmt.Fprintf(a.out, "connected to %s:%d via path %q\n", options.Address, options.Port, options.Path)
	return nil
}

func (a *App) programs(args []string) error {
	fs := a.newFlagSet("programs")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("programs takes no positional arguments")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	programs, err := client.Programs()
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(programs))
	for _, p := range programs {
		rows = append(rows, []string{p.Name, uint32String(p.ID)})
	}
	return output.Snapshot(a.out, flags.format, []string{"NAME", "ID"}, rows, programs)
}

func (a *App) tags(args []string) error {
	fs := a.newFlagSet("tags")
	flags := addCommonFlags(fs, true)
	filter := fs.String("filter", "", "case-insensitive name filter")
	program := fs.String("program", "", "program scope filter")
	limit := fs.Int("limit", 0, "maximum rows; 0 means unlimited")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("tags takes no positional arguments")
	}
	if *limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	tags, err := client.Tags()
	if err != nil {
		return err
	}
	filtered := make([]logixclient.Tag, 0, len(tags))
	for _, tag := range tags {
		if *filter != "" && !strings.Contains(strings.ToLower(tag.Name), strings.ToLower(*filter)) {
			continue
		}
		if *program != "" && !strings.EqualFold(tag.Program, *program) {
			continue
		}
		filtered = append(filtered, tag)
		if *limit > 0 && len(filtered) >= *limit {
			break
		}
	}
	rows := make([][]string, 0, len(filtered))
	for _, tag := range filtered {
		rows = append(rows, []string{tag.Name, tag.Type, uint32String(tag.Instance), tag.Program, formatDimensions(tag.Dimensions)})
	}
	return output.Snapshot(a.out, flags.format, []string{"NAME", "TYPE", "INSTANCE", "PROGRAM", "DIMENSIONS"}, rows, filtered)
}

func (a *App) read(args []string) error {
	fs := a.newFlagSet("read")
	flags := addCommonFlags(fs, true)
	valueType := fs.String("type", "auto", "auto, bool, sint, int, dint, lint, usint, uint, udint, ulint, real, lreal, or string")
	elements := fs.Uint("elements", 1, "number of elements")
	tag, err := parseTagArgs(fs, args, "read")
	if err != nil {
		return err
	}
	if *elements == 0 || *elements > 65535 {
		return fmt.Errorf("elements must be between 1 and 65535")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	value, actualType, err := client.Read(tag, *valueType, uint16(*elements))
	if err != nil {
		return err
	}
	result := sample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Tag: tag, Type: actualType, Value: value}
	row := []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)}
	return output.Snapshot(a.out, flags.format, []string{"TIMESTAMP", "TAG", "TYPE", "VALUE"}, [][]string{row}, result)
}

func (a *App) write(args []string) error {
	fs := a.newFlagSet("write")
	flags := addCommonFlags(fs, true)
	valueType := fs.String("type", "", "BOOL, SINT, INT, DINT, LINT, USINT, UINT, UDINT, ULINT, REAL, LREAL, or STRING")
	var rawValue string
	valueSet := false
	fs.Func("value", "value to write", func(value string) error {
		rawValue = value
		valueSet = true
		return nil
	})
	yes := fs.Bool("yes", false, "transmit the write")
	dryRun := fs.Bool("dry-run", false, "explicitly keep the write as a dry run")
	tag, err := parseTagArgs(fs, args, "write")
	if err != nil {
		return err
	}
	if err := validateWriteMode(*yes, *dryRun); err != nil {
		return err
	}
	if *valueType == "" {
		return fmt.Errorf("write requires --type")
	}
	if !valueSet {
		return fmt.Errorf("write requires --value")
	}
	value, err := logixclient.ParseValue(*valueType, rawValue)
	if err != nil {
		return err
	}
	options, _, err := a.options(flags)
	if err != nil {
		return err
	}
	if !*yes {
		fmt.Fprintf(a.out, "dry-run: would write controller=%s:%d tag=%s type=%s value=%v; add --yes to transmit\n", options.Address, options.Port, tag, strings.ToLower(*valueType), value)
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
	if err := client.Write(tag, value); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "wrote tag=%s type=%s value=%v\n", tag, strings.ToLower(*valueType), value)
	return nil
}

func (a *App) watch(args []string) error {
	fs := a.newFlagSet("watch")
	flags := addCommonFlags(fs, false)
	valueType := fs.String("type", "auto", "tag type or auto")
	elements := fs.Uint("elements", 1, "number of elements")
	interval := fs.Duration("interval", time.Second, "poll interval")
	count := fs.Int("count", 0, "number of samples; 0 means unlimited")
	duration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")
	tag, err := parseTagArgs(fs, args, "watch")
	if err != nil {
		return err
	}
	if *elements == 0 || *elements > 65535 {
		return fmt.Errorf("elements must be between 1 and 65535")
	}
	if err := validateWatchOptions(*interval, *count, *duration); err != nil {
		return err
	}
	stream, err := output.NewStream(a.out, flags.format, []string{"timestamp", "tag", "type", "value"})
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
		value, actualType, err := client.Read(tag, *valueType, uint16(*elements))
		if err != nil {
			return err
		}
		result := sample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Tag: tag, Type: actualType, Value: value}
		row := []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)}
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

func formatDimensions(values []int) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, "x")
}

func parseTagArgs(fs *flag.FlagSet, args []string, command string) (string, error) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		tag := args[0]
		if err := fs.Parse(args[1:]); err != nil {
			return "", err
		}
		if fs.NArg() != 0 {
			return "", fmt.Errorf("%s requires exactly one tag", command)
		}
		return tag, nil
	}
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 1 {
		return "", fmt.Errorf("%s requires exactly one tag", command)
	}
	return fs.Arg(0), nil
}
