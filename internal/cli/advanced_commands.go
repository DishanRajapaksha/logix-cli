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

type repeatedValue []string

func (v *repeatedValue) String() string { return strings.Join(*v, ",") }
func (v *repeatedValue) Set(value string) error {
	*v = append(*v, value)
	return nil
}

type readSpec struct {
	Tag      string
	Type     string
	Elements uint16
}

type writeSpec struct {
	Tag   string
	Type  string
	Raw   string
	Value any
}

type statusResult struct {
	Address     string `json:"address"`
	ProductName string `json:"product_name"`
	Revision    string `json:"revision"`
	Status      uint16 `json:"status"`
	StatusHex   string `json:"status_hex"`
}

type writeResult struct {
	Tag    string `json:"tag"`
	Type   string `json:"type"`
	Value  any    `json:"value"`
	Status string `json:"status"`
}

func (a *App) identify(args []string) error {
	fs := a.newFlagSet("identify")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("identify takes no positional arguments")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	identity, err := client.Identity()
	if err != nil {
		return err
	}
	row := []string{
		strconv.FormatUint(uint64(identity.VendorID), 10),
		strconv.FormatUint(uint64(identity.DeviceType), 10),
		strconv.FormatUint(uint64(identity.ProductCode), 10),
		identity.Revision,
		fmt.Sprintf("0x%04X", identity.Status),
		fmt.Sprintf("0x%08X", identity.SerialNumber),
		identity.ProductName,
	}
	return output.Snapshot(a.out, flags.format,
		[]string{"VENDOR_ID", "DEVICE_TYPE", "PRODUCT_CODE", "REVISION", "STATUS", "SERIAL_NUMBER", "PRODUCT_NAME"},
		[][]string{row}, identity)
}

func (a *App) status(args []string) error {
	fs := a.newFlagSet("status")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("status takes no positional arguments")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	client, options, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)
	identity, err := client.Identity()
	if err != nil {
		return err
	}
	result := statusResult{
		Address:     fmt.Sprintf("%s:%d", options.Address, options.Port),
		ProductName: identity.ProductName,
		Revision:    identity.Revision,
		Status:      identity.Status,
		StatusHex:   fmt.Sprintf("0x%04X", identity.Status),
	}
	row := []string{result.Address, result.ProductName, result.Revision, result.StatusHex}
	return output.Snapshot(a.out, flags.format, []string{"ADDRESS", "PRODUCT_NAME", "REVISION", "STATUS"}, [][]string{row}, result)
}

func (a *App) readMulti(args []string) error {
	fs := a.newFlagSet("read-multi")
	flags := addCommonFlags(fs, true)
	var rawItems repeatedValue
	fs.Var(&rawItems, "item", "repeat TAG[=TYPE[:ELEMENTS]]")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("read-multi takes no positional arguments; repeat --item")
	}
	if len(rawItems) == 0 {
		return fmt.Errorf("read-multi requires at least one --item")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	specs, err := parseReadSpecs(rawItems)
	if err != nil {
		return err
	}
	client, _, err := a.connect(flags)
	if err != nil {
		return err
	}
	defer closeClient(client)

	results := make([]sample, 0, len(specs))
	rows := make([][]string, 0, len(specs))
	for _, spec := range specs {
		value, actualType, err := client.Read(spec.Tag, spec.Type, spec.Elements)
		if err != nil {
			return err
		}
		result := sample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Tag: spec.Tag, Type: actualType, Value: value}
		results = append(results, result)
		rows = append(rows, []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)})
	}
	return output.Snapshot(a.out, flags.format, []string{"TIMESTAMP", "TAG", "TYPE", "VALUE"}, rows, results)
}

func (a *App) writeMulti(args []string) error {
	fs := a.newFlagSet("write-multi")
	flags := addCommonFlags(fs, true)
	var rawSets repeatedValue
	fs.Var(&rawSets, "set", "repeat TAG=TYPE:VALUE")
	yes := fs.Bool("yes", false, "transmit all writes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("write-multi takes no positional arguments; repeat --set")
	}
	if len(rawSets) == 0 {
		return fmt.Errorf("write-multi requires at least one --set")
	}
	if err := output.ValidateSnapshotFormat(flags.format); err != nil {
		return err
	}
	specs, err := parseWriteSpecs(rawSets)
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
			if err := client.Write(spec.Tag, spec.Value); err != nil {
				return fmt.Errorf("write-multi stopped after %d successful writes: %w", i, err)
			}
		}
		status = "written"
	}

	results := make([]writeResult, 0, len(specs))
	rows := make([][]string, 0, len(specs))
	for _, spec := range specs {
		result := writeResult{Tag: spec.Tag, Type: spec.Type, Value: spec.Value, Status: status}
		results = append(results, result)
		rows = append(rows, []string{result.Tag, result.Type, fmt.Sprint(result.Value), result.Status})
	}
	return output.Snapshot(a.out, flags.format, []string{"TAG", "TYPE", "VALUE", "STATUS"}, rows, results)
}

func (a *App) watchMulti(args []string) error {
	fs := a.newFlagSet("watch-multi")
	flags := addCommonFlags(fs, false)
	var rawItems repeatedValue
	fs.Var(&rawItems, "item", "repeat TAG[=TYPE[:ELEMENTS]]")
	interval := fs.Duration("interval", time.Second, "poll interval")
	count := fs.Int("count", 0, "number of poll cycles; 0 means until interrupted")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("watch-multi takes no positional arguments; repeat --item")
	}
	if len(rawItems) == 0 {
		return fmt.Errorf("watch-multi requires at least one --item")
	}
	if *interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if *count < 0 {
		return fmt.Errorf("count must be non-negative")
	}
	specs, err := parseReadSpecs(rawItems)
	if err != nil {
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

	cycles := 0
	for {
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)
		for _, spec := range specs {
			value, actualType, err := client.Read(spec.Tag, spec.Type, spec.Elements)
			if err != nil {
				return err
			}
			result := sample{Timestamp: timestamp, Tag: spec.Tag, Type: actualType, Value: value}
			row := []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)}
			if err := stream.Write(row, result); err != nil {
				return err
			}
		}
		cycles++
		if *count > 0 && cycles >= *count {
			return nil
		}
		time.Sleep(*interval)
	}
}

func parseReadSpecs(values []string) ([]readSpec, error) {
	specs := make([]readSpec, 0, len(values))
	for _, raw := range values {
		spec, err := parseReadSpec(raw)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func parseReadSpec(raw string) (readSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return readSpec{}, fmt.Errorf("%w: read item cannot be empty", logixclient.ErrValidation)
	}
	spec := readSpec{Tag: raw, Type: "auto", Elements: 1}
	if index := strings.LastIndex(raw, "="); index >= 0 {
		spec.Tag = strings.TrimSpace(raw[:index])
		typeAndElements := strings.TrimSpace(raw[index+1:])
		if typeAndElements == "" {
			return readSpec{}, fmt.Errorf("%w: read item %q requires a type after '='", logixclient.ErrValidation, raw)
		}
		parts := strings.SplitN(typeAndElements, ":", 2)
		spec.Type = parts[0]
		if len(parts) == 2 {
			parsed, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 16)
			if err != nil || parsed == 0 {
				return readSpec{}, fmt.Errorf("%w: read item %q has invalid element count", logixclient.ErrValidation, raw)
			}
			spec.Elements = uint16(parsed)
		}
	}
	if spec.Tag == "" {
		return readSpec{}, fmt.Errorf("%w: read item %q has an empty tag", logixclient.ErrValidation, raw)
	}
	normalised, err := logixclient.NormaliseType(spec.Type)
	if err != nil {
		return readSpec{}, err
	}
	spec.Type = normalised
	return spec, nil
}

func parseWriteSpecs(values []string) ([]writeSpec, error) {
	specs := make([]writeSpec, 0, len(values))
	for _, raw := range values {
		spec, err := parseWriteSpec(raw)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func parseWriteSpec(raw string) (writeSpec, error) {
	equals := strings.Index(raw, "=")
	if equals <= 0 {
		return writeSpec{}, fmt.Errorf("%w: write set %q must be TAG=TYPE:VALUE", logixclient.ErrValidation, raw)
	}
	tag := strings.TrimSpace(raw[:equals])
	if tag == "" {
		return writeSpec{}, fmt.Errorf("%w: write set %q has an empty tag", logixclient.ErrValidation, raw)
	}
	typeAndValue := raw[equals+1:]
	colon := strings.Index(typeAndValue, ":")
	if colon <= 0 {
		return writeSpec{}, fmt.Errorf("%w: write set %q must be TAG=TYPE:VALUE", logixclient.ErrValidation, raw)
	}
	valueType, err := logixclient.NormaliseType(typeAndValue[:colon])
	if err != nil {
		return writeSpec{}, err
	}
	if valueType == "auto" {
		return writeSpec{}, fmt.Errorf("%w: write set %q requires an explicit type", logixclient.ErrValidation, raw)
	}
	rawValue := typeAndValue[colon+1:]
	value, err := logixclient.ParseValue(valueType, rawValue)
	if err != nil {
		return writeSpec{}, err
	}
	return writeSpec{Tag: tag, Type: valueType, Raw: rawValue, Value: value}, nil
}

var _ flag.Value = (*repeatedValue)(nil)
