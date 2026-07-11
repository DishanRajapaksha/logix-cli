#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def replace_once(path, old, new):
    file_path = ROOT / path
    data = file_path.read_text()
    if new in data:
        return
    count = data.count(old)
    if count != 1:
        raise RuntimeError(f"{path}: expected one replacement, found {count}")
    file_path.write_text(data.replace(old, new, 1))

# Raw tag write/watch controls.
replace_once("internal/cli/client_commands.go", '\tyes := fs.Bool("yes", false, "transmit the write")\n\ttag, err := parseTagArgs(fs, args, "write")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif *valueType == "" {\n', '\tyes := fs.Bool("yes", false, "transmit the write")\n\tdryRun := fs.Bool("dry-run", false, "explicitly keep the write as a dry run")\n\ttag, err := parseTagArgs(fs, args, "write")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif err := validateWriteMode(*yes, *dryRun); err != nil {\n\t\treturn err\n\t}\n\tif *valueType == "" {\n')
replace_once("internal/cli/client_commands.go", '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of samples; 0 means until interrupted")\n\ttag, err := parseTagArgs(fs, args, "watch")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif *elements == 0 || *elements > 65535 {\n\t\treturn fmt.Errorf("elements must be between 1 and 65535")\n\t}\n\tif *interval <= 0 {\n\t\treturn fmt.Errorf("interval must be positive")\n\t}\n\tif *count < 0 {\n\t\treturn fmt.Errorf("count must be non-negative")\n\t}\n', '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of samples; 0 means unlimited")\n\tduration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")\n\ttag, err := parseTagArgs(fs, args, "watch")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif *elements == 0 || *elements > 65535 {\n\t\treturn fmt.Errorf("elements must be between 1 and 65535")\n\t}\n\tif err := validateWatchOptions(*interval, *count, *duration); err != nil {\n\t\treturn err\n\t}\n')
replace_once("internal/cli/client_commands.go", '\twritten := 0\n\tfor {\n\t\tvalue, actualType, err := client.Read(tag, *valueType, uint16(*elements))\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tresult := sample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Tag: tag, Type: actualType, Value: value}\n\t\trow := []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)}\n\t\tif err := stream.Write(row, result); err != nil {\n\t\t\treturn err\n\t\t}\n\t\twritten++\n\t\tif *count > 0 && written >= *count {\n\t\t\treturn nil\n\t\t}\n\t\ttime.Sleep(*interval)\n\t}\n', '\tstarted := time.Now()\n\twritten := 0\n\tfor {\n\t\tvalue, actualType, err := client.Read(tag, *valueType, uint16(*elements))\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tresult := sample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Tag: tag, Type: actualType, Value: value}\n\t\trow := []string{result.Timestamp, result.Tag, result.Type, fmt.Sprint(result.Value)}\n\t\tif err := stream.Write(row, result); err != nil {\n\t\t\treturn err\n\t\t}\n\t\twritten++\n\t\tif watchShouldStop(started, written, *count, *duration) {\n\t\t\treturn nil\n\t\t}\n\t\tif !waitForNextWatch(started, *duration, *interval) {\n\t\t\treturn nil\n\t\t}\n\t}\n')
