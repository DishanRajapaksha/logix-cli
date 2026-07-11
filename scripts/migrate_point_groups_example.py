#!/usr/bin/env python3
from pathlib import Path

path = Path(__file__).resolve().parents[1] / "config.example.yaml"
data = path.read_text()
block = """groups:
  - name: motor
    points:
      - motor_speed
      - motor_enabled
    description: Motor operating values
"""
if "\ngroups:\n" not in data:
    marker = "    description: Motor enable command\n"
    if data.count(marker) != 1:
        raise RuntimeError("config.example.yaml: motor enable description marker not found exactly once")
    data = data.replace(marker, marker + block, 1)
    path.write_text(data)
