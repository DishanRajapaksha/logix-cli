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

# Multi-tag write/watch controls.
replace_once("internal/cli/advanced_commands.go", '\tyes := fs.Bool("yes", false, "transmit all writes")\n\tif err := fs.Parse(args); err != nil {\n\t\treturn err\n\t}\n', '\tyes := fs.Bool("yes", false, "transmit all writes")\n\tdryRun := fs.Bool("dry-run", false, "explicitly keep all writes as a dry run")\n\tif err := fs.Parse(args); err != nil {\n\t\treturn err\n\t}\n\tif err := validateWriteMode(*yes, *dryRun); err != nil {\n\t\treturn err\n\t}\n')
replace_once("internal/cli/advanced_commands.go", '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of poll cycles; 0 means until interrupted")\n\tif err := fs.Parse(args); err != nil {\n\t\treturn err\n\t}\n', '\tinterval := fs.Duration("interval", time.Second, "poll interval")\n\tcount := fs.Int("count", 0, "number of poll cycles; 0 means unlimited")\n\tduration := fs.Duration("duration", 0, "maximum watch duration; 0 means unlimited")\n\tif err := fs.Parse(args); err != nil {\n\t\treturn err\n\t}\n')
replace_once("internal/cli/advanced_commands.go", '\tif *interval <= 0 {\n\t\treturn fmt.Errorf("interval must be positive")\n\t}\n\tif *count < 0 {\n\t\treturn fmt.Errorf("count must be non-negative")\n\t}\n', '\tif err := validateWatchOptions(*interval, *count, *duration); err != nil {\n\t\treturn err\n\t}\n')
replace_once("internal/cli/advanced_commands.go", '\tcycles := 0\n\tfor {\n\t\ttimestamp := time.Now().UTC().Format(time.RFC3339Nano)\n', '\tstarted := time.Now()\n\tcycles := 0\n\tfor {\n\t\ttimestamp := time.Now().UTC().Format(time.RFC3339Nano)\n')
replace_once("internal/cli/advanced_commands.go", '\t\tcycles++\n\t\tif *count > 0 && cycles >= *count {\n\t\t\treturn nil\n\t\t}\n\t\ttime.Sleep(*interval)\n\t}\n', '\t\tcycles++\n\t\tif watchShouldStop(started, cycles, *count, *duration) {\n\t\t\treturn nil\n\t\t}\n\t\tif !waitForNextWatch(started, *duration, *interval) {\n\t\t\treturn nil\n\t\t}\n\t}\n')
