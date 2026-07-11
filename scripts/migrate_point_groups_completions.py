#!/usr/bin/env python3
from pathlib import Path

path = Path(__file__).resolve().parents[1] / "internal/cli/completions.go"
data = path.read_text()
old = "init-config validate-config test-connection status identify programs tags points read read-multi read-point write write-multi write-point watch watch-multi watch-point completions version"
new = "init-config validate-config test-connection status identify programs tags points groups read read-multi read-point read-group write write-multi write-point write-group watch watch-multi watch-point watch-group completions version"
if new not in data:
    count = data.count(old)
    if count != 2:
        raise RuntimeError(f"internal/cli/completions.go: expected two command lists, found {count}")
    path.write_text(data.replace(old, new))
