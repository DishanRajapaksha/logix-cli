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

# Shared point samples carry an optional group.
replace_once("internal/cli/point_commands.go", 'type pointSample struct {\n\tTimestamp string `json:"timestamp"`\n\tPoint     string `json:"point"`\n\tTag       string `json:"tag"`\n\tType      string `json:"type"`\n\tValue     any    `json:"value"`\n\tUnit      string `json:"unit,omitempty"`\n}\n', 'type pointSample struct {\n\tTimestamp string `json:"timestamp"`\n\tGroup     string `json:"group,omitempty"`\n\tPoint     string `json:"point"`\n\tTag       string `json:"tag"`\n\tType      string `json:"type"`\n\tValue     any    `json:"value"`\n\tUnit      string `json:"unit,omitempty"`\n}\n')
replace_once("internal/cli/point_commands.go", '\tyes := fs.Bool("yes", false, "transmit the write")\n\tname, err := parseTagArgs(fs, args, "write-point")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif !valueSet {\n', '\tyes := fs.Bool("yes", false, "transmit the write")\n\tdryRun := fs.Bool("dry-run", false, "explicitly keep the write as a dry run")\n\tname, err := parseTagArgs(fs, args, "write-point")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif err := validateWriteMode(*yes, *dryRun); err != nil {\n\t\treturn err\n\t}\n\tif !valueSet {\n')
replace_once("internal/cli/point_commands.go", '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of samples; 0 means until interrupted")\n\tname, err := parseTagArgs(fs, args, "watch-point")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif *interval <= 0 {\n\t\treturn fmt.Errorf("interval must be positive")\n\t}\n\tif *count < 0 {\n\t\treturn fmt.Errorf("count must be non-negative")\n\t}\n', '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of samples; 0 means unlimited")\n\tduration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")\n\tname, err := parseTagArgs(fs, args, "watch-point")\n\tif err != nil {\n\t\treturn err\n\t}\n\tif err := validateWatchOptions(*interval, *count, *duration); err != nil {\n\t\treturn err\n\t}\n')
replace_once("internal/cli/point_commands.go", '\twritten := 0\n\tfor {\n\t\tvalue, actualType, err := client.Read(point.Tag, point.Type, point.Elements)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tresult := pointSample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}\n\t\trow := []string{result.Timestamp, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit}\n\t\tif err := stream.Write(row, result); err != nil {\n\t\t\treturn err\n\t\t}\n\t\twritten++\n\t\tif *count > 0 && written >= *count {\n\t\t\treturn nil\n\t\t}\n\t\ttime.Sleep(*interval)\n\t}\n', '\tstarted := time.Now()\n\twritten := 0\n\tfor {\n\t\tvalue, actualType, err := client.Read(point.Tag, point.Type, point.Elements)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tresult := pointSample{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Point: point.Name, Tag: point.Tag, Type: actualType, Value: value, Unit: point.Unit}\n\t\trow := []string{result.Timestamp, result.Point, result.Tag, result.Type, fmt.Sprint(result.Value), result.Unit}\n\t\tif err := stream.Write(row, result); err != nil {\n\t\t\treturn err\n\t\t}\n\t\twritten++\n\t\tif watchShouldStop(started, written, *count, *duration) {\n\t\t\treturn nil\n\t\t}\n\t\tif !waitForNextWatch(started, *duration, *interval) {\n\t\t\treturn nil\n\t\t}\n\t}\n')
