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


# README additions.
replace_once("README.md", '- configured named points with type, element count, engineering unit and write policy\n', '- configured named points with type, element count, engineering unit and write policy\n- configured point groups for reusable read, write and watch workflows\n')
replace_once("README.md", '- safety-gated single, multi-tag and named-point writes, with dry-run as the default\n- single, multi-tag and named-point polling with text, JSON Lines or CSV output\n', '- safety-gated single, multi-tag, named-point and group writes, with dry-run as the default\n- single, multi-tag, named-point and group polling with count or duration bounds\n')
replace_once("README.md", '    description: Motor enable command\n```\n', '    description: Motor enable command\ngroups:\n  - name: motor\n    points:\n      - motor_speed\n      - motor_enabled\n    description: Motor operating values\n```\n')
replace_once("README.md", 'Configuration is parsed as strict YAML. Unknown fields, duplicate point names, invalid types and malformed durations fail validation instead of being quietly ignored.\n', 'Point groups are shared across profiles and reference named points. Group and point names are matched case-insensitively, while output preserves the canonical configured names. Unknown points, duplicate group names and duplicate entries within a group fail validation.\n\nConfiguration is parsed as strict YAML. Unknown fields, duplicate point names, invalid types and malformed durations fail validation instead of being quietly ignored.\n')
replace_once("README.md", '### Read a tag\n', '### Use configured point groups\n\nList groups without connecting:\n\n```bash\nlogix-cli groups\nlogix-cli groups --filter motor --format json\n```\n\nRead every point in a group over one controller connection:\n\n```bash\nlogix-cli read-group motor\nlogix-cli read-group motor --format json\n```\n\nWrite selected points by configured point name. Every target must belong to the group and be declared writable:\n\n```bash\nlogix-cli write-group motor --set motor_enabled=true\nlogix-cli write-group motor --set motor_enabled=true --yes\n```\n\nGroup writes validate the complete set before connecting, are dry-run by default, and are sequential rather than transactional.\n\nPoll the group with one timestamp shared by every point in a cycle:\n\n```bash\nlogix-cli watch-group motor --interval 1s --duration 30s --format jsonl\n```\n\n### Read a tag\n')
replace_once("README.md", 'logix-cli write-point motor_enabled --value true\nlogix-cli write-point motor_enabled --value true --yes\n', 'logix-cli write-point motor_enabled --value true\nlogix-cli write-point motor_enabled --value true --dry-run\nlogix-cli write-point motor_enabled --value true --yes\n')
replace_once("README.md", 'logix-cli write Motor.Enable --type bool --value true\n', 'logix-cli write Motor.Enable --type bool --value true\nlogix-cli write Motor.Enable --type bool --value true --dry-run\n')
replace_once("README.md", 'logix-cli watch Motor.Speed --type real --interval 1s\n', 'logix-cli watch Motor.Speed --type real --interval 1s\nlogix-cli watch Motor.Speed --type real --duration 30s --format jsonl\n')
replace_once("README.md", '`--count` counts poll cycles. Each cycle reads every configured item and gives the rows a common timestamp.\n', '`--count` counts poll cycles. Each cycle reads every configured item and gives the rows a common timestamp. All watch commands also accept `--duration`; whichever bound is reached first stops the stream. A zero count or duration means unlimited.\n')
replace_once("README.md", '## Output contract\n', '`--dry-run` may be supplied explicitly to any write command. It cannot be combined with `--yes`.\n\n## Output contract\n')
