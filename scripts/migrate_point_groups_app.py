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

# Command dispatch, help, and completion surface.
replace_once("internal/cli/app.go", '\tcase "points":\n\t\terr = a.points(args[1:])\n\tcase "read":\n', '\tcase "points":\n\t\terr = a.points(args[1:])\n\tcase "groups":\n\t\terr = a.groups(args[1:])\n\tcase "read":\n')
replace_once("internal/cli/app.go", '\tcase "read-point":\n\t\terr = a.readPoint(args[1:])\n\tcase "write":\n', '\tcase "read-point":\n\t\terr = a.readPoint(args[1:])\n\tcase "read-group":\n\t\terr = a.readGroup(args[1:])\n\tcase "write":\n')
replace_once("internal/cli/app.go", '\tcase "write-point":\n\t\terr = a.writePoint(args[1:])\n\tcase "watch":\n', '\tcase "write-point":\n\t\terr = a.writePoint(args[1:])\n\tcase "write-group":\n\t\terr = a.writeGroup(args[1:])\n\tcase "watch":\n')
replace_once("internal/cli/app.go", '\tcase "watch-point":\n\t\terr = a.watchPoint(args[1:])\n\tcase "completions":\n', '\tcase "watch-point":\n\t\terr = a.watchPoint(args[1:])\n\tcase "watch-group":\n\t\terr = a.watchGroup(args[1:])\n\tcase "completions":\n')
replace_once("internal/cli/app.go", '  logix-cli points --format json\n  logix-cli read Motor.Speed --type real\n', '  logix-cli points --format json\n  logix-cli groups --format json\n  logix-cli read Motor.Speed --type real\n')
replace_once("internal/cli/app.go", '  logix-cli read-point motor_speed\n  logix-cli write Motor.Enable --type bool --value true --yes\n', '  logix-cli read-point motor_speed\n  logix-cli read-group motor --format json\n  logix-cli write Motor.Enable --type bool --value true --yes\n')
replace_once("internal/cli/app.go", '  logix-cli write-point motor_enabled --value true --yes\n  logix-cli watch Motor.Speed --type real --interval 1s --format jsonl\n', '  logix-cli write-point motor_enabled --value true --yes\n  logix-cli write-group motor --set motor_enabled=true --yes\n  logix-cli watch Motor.Speed --type real --interval 1s --duration 30s --format jsonl\n')
replace_once("internal/cli/app.go", '  logix-cli watch-point motor_speed --format jsonl\n  logix-cli completions zsh\n', '  logix-cli watch-point motor_speed --duration 30s --format jsonl\n  logix-cli watch-group motor --duration 30s --format jsonl\n  logix-cli completions zsh\n')
replace_once("internal/cli/app.go", '  points           List configured named points\n  read             Read one tag, with optional type detection\n', '  points           List configured named points\n  groups           List configured point groups\n  read             Read one tag, with optional type detection\n')
replace_once("internal/cli/app.go", '  read-point       Read a configured named point\n  write            Write one tag; dry-run unless --yes is supplied\n', '  read-point       Read a configured named point\n  read-group       Read all points in a configured group\n  write            Write one tag; dry-run unless --yes is supplied\n')
replace_once("internal/cli/app.go", '  write-point      Write a configured named point; dry-run unless --yes is supplied\n  watch            Poll one tag repeatedly\n', '  write-point      Write a configured named point; dry-run unless --yes is supplied\n  write-group      Write selected writable points in a configured group\n  watch            Poll one tag repeatedly\n')
replace_once("internal/cli/app.go", '  watch-point      Poll a configured named point repeatedly\n  completions      Generate Bash or Zsh completion scripts\n', '  watch-point      Poll a configured named point repeatedly\n  watch-group      Poll all points in a configured group\n  completions      Generate Bash or Zsh completion scripts\n')
replace_once("internal/cli/app.go", '\tcase "read", "write", "watch", "read-point", "write-point", "watch-point":\n', '\tcase "read", "write", "watch", "read-point", "write-point", "watch-point", "read-group", "write-group", "watch-group":\n')
replace_once("internal/cli/app.go", '\tcase "validate-config", "test-connection", "status", "identify", "programs", "tags", "points", "read", "read-multi", "read-point", "write", "write-multi", "write-point", "watch", "watch-multi", "watch-point":\n', '\tcase "validate-config", "test-connection", "status", "identify", "programs", "tags", "points", "groups", "read", "read-multi", "read-point", "read-group", "write", "write-multi", "write-point", "write-group", "watch", "watch-multi", "watch-point", "watch-group":\n')

