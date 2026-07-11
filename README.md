# logix-cli

`logix-cli` is a script-friendly command-line client for reading and writing tags on Rockwell Automation ControlLogix, CompactLogix and compatible CIP-based controllers over EtherNet/IP.

It follows the command and output conventions used by `opc-xml-da-cli`, `opc-ua-cli`, `iec-104-cli` and `modbus-cli`.

The implementation uses [`github.com/danomagnum/gologix`](https://github.com/danomagnum/gologix). PLC-5, SLC and MicroLogix controllers that require PCCC are outside the project scope.

## Features

- YAML profiles with command-line overrides
- configured named points with type, element count, engineering unit and write policy
- configured point groups for reusable read, write and watch workflows
- connection diagnostics
- CIP Identity and controller-status inspection
- controller program discovery
- controller and program-scoped tag discovery
- typed scalar and array reads
- heterogeneous multi-tag reads over one connection
- safety-gated single, multi-tag, named-point and group writes, with dry-run as the default
- single, multi-tag, named-point and group polling with count or duration bounds
- table, text, JSON and CSV snapshot output
- Bash and Zsh completions
- script-friendly exit codes

## Install

```bash
go install github.com/DishanRajapaksha/logix-cli@latest
```

Or build locally:

```bash
git clone https://github.com/DishanRajapaksha/logix-cli.git
cd logix-cli
go build -o logix-cli .
```

## Configuration

Create a starter file:

```bash
logix-cli init-config
```

```yaml
default_profile: local
profiles:
  local:
    address: 192.168.1.10
    port: 44818
    path: "1,0"
    timeout: 5s
points:
  - name: motor_speed
    tag: Motor.Speed
    type: real
    elements: 1
    unit: rpm
    description: Motor shaft speed
  - name: motor_enabled
    tag: Motor.Enable
    type: bool
    elements: 1
    writable: true
    description: Motor enable command
groups:
  - name: motor
    points:
      - motor_speed
      - motor_enabled
    description: Motor operating values
```

The usual ControlLogix/CompactLogix route is `1,0`, meaning backplane, slot 0. Use an empty path for devices such as some Micro800 controllers:

```yaml
  micro800:
    address: 192.168.1.20
    port: 44818
    path: ""
    timeout: 5s
```

Named points are shared by all connection profiles. Their fields are:

| Field | Required | Meaning |
|---|---|---|
| `name` | yes | Stable CLI-facing name, matched case-insensitively |
| `tag` | yes | Controller or program-scoped Logix tag |
| `type` | no | Logix type; defaults to `auto` for reads |
| `elements` | no | Array element count; defaults to `1` |
| `unit` | no | Engineering unit carried into output |
| `writable` | no | Allows `write-point`; defaults to `false` |
| `description` | no | Human-readable explanation |

Writable points require an explicit type. A configuration that marks an auto-detected point writable is rejected rather than being permitted to improvise against a PLC.

Point groups are shared across profiles and reference named points. Group and point names are matched case-insensitively, while output preserves the canonical configured names. Unknown points, duplicate group names and duplicate entries within a group fail validation.

Configuration is parsed as strict YAML. Unknown fields, duplicate point names, invalid types and malformed durations fail validation instead of being quietly ignored.

Validate without connecting:

```bash
logix-cli validate-config --profile local
```

Global flags may appear before or after the command:

```bash
logix-cli --profile local --format json programs
logix-cli programs --profile local --format json
```

## Commands

### Test the connection

`test-connection` only verifies that the EtherNet/IP and CIP connection can be opened and closed.

```bash
logix-cli test-connection
logix-cli test-connection --address 192.168.1.10 --path 1,0
```

### Inspect status and identity

`status` reads the controller product name, revision and raw status word from the CIP Identity object:

```bash
logix-cli status
logix-cli status --format json
```

`identify` returns the complete identity summary exposed by the CLI:

```bash
logix-cli identify
logix-cli identify --format json
```

The fields are vendor ID, device type, product code, revision, status, serial number and product name. Numeric IDs remain numeric because pretending every vendor-specific value has a trustworthy friendly name would be theatre.

### List programs

```bash
logix-cli programs
logix-cli programs --format json
```

### List tags

```bash
logix-cli tags
logix-cli tags --filter Motor --limit 50
logix-cli tags --program MainProgram --format csv
```

Tag discovery can be expensive on large controllers. Use `--filter` and `--limit` to keep the output civilised, although the controller still returns its tag catalogue before local filtering.

### Use configured named points

List the point catalogue without connecting:

```bash
logix-cli points
logix-cli points --format json
```

Read a point using its configured tag, type and element count:

```bash
logix-cli read-point motor_speed
logix-cli read-point motor_speed --format json
```

Writes are dry-run by default and only work for points declared with `writable: true`:

```bash
logix-cli write-point motor_enabled --value true
logix-cli write-point motor_enabled --value true --dry-run
logix-cli write-point motor_enabled --value true --yes
```

Attempting to write a read-only point returns exit code `7` before a controller connection is opened.

Poll a point while preserving its point name, underlying tag and engineering unit in every row:

```bash
logix-cli watch-point motor_speed --interval 1s
logix-cli watch-point motor_speed --count 10 --format jsonl
```

### Use configured point groups

List groups without connecting:

```bash
logix-cli groups
logix-cli groups --filter motor --format json
```

Read every point in a group over one controller connection:

```bash
logix-cli read-group motor
logix-cli read-group motor --format json
```

Write selected points by configured point name. Every target must belong to the group and be declared writable:

```bash
logix-cli write-group motor --set motor_enabled=true
logix-cli write-group motor --set motor_enabled=true --yes
```

Group writes validate the complete set before connecting, are dry-run by default, and are sequential rather than transactional.

Poll the group with one timestamp shared by every point in a cycle:

```bash
logix-cli watch-group motor --interval 1s --duration 30s --format jsonl
```

### Read a tag

Auto-detect the type:

```bash
logix-cli read Motor.Speed
```

Specify a type and element count:

```bash
logix-cli read Motor.Speed --type real
logix-cli read ProductionCounts[0] --type dint --elements 10 --format json
```

Supported types:

```text
auto, bool, sint, int, dint, lint,
usint, uint, udint, ulint, real, lreal, string
```

### Read several tags

Repeat `--item` using this grammar:

```text
TAG[=TYPE[:ELEMENTS]]
```

Examples:

```bash
logix-cli read-multi \
  --item Motor.Speed=real \
  --item Counter=dint \
  --item ProductionCounts[0]=dint:10

logix-cli read-multi \
  --item Program:MainProgram.State=dint \
  --item Program:MainProgram.Message=string \
  --format json
```

Omitting the type uses automatic type detection. The command reuses one controller connection but currently performs one tag request per item. It is deliberately honest about this rather than calling a loop a magical atomic batch.

### Write a tag

Writes are dry-run by default:

```bash
logix-cli write Motor.Enable --type bool --value true
logix-cli write Motor.Enable --type bool --value true --dry-run
```

Transmit only with `--yes`:

```bash
logix-cli write Motor.Enable --type bool --value true --yes
logix-cli write RecipeNumber --type dint --value 12 --yes
```

The CLI deliberately requires an explicit type for writes. Guessing types while altering a PLC is an inventive way to ruin an afternoon.

### Write several tags

Repeat `--set` using this grammar:

```text
TAG=TYPE:VALUE
```

Dry-run the complete set:

```bash
logix-cli write-multi \
  --set Motor.Enable=bool:true \
  --set RecipeNumber=dint:12
```

Transmit the writes:

```bash
logix-cli write-multi \
  --set Motor.Enable=bool:true \
  --set RecipeNumber=dint:12 \
  --yes
```

Multi-tag writes are sequential and **not transactional**. If a later write fails, earlier successful writes remain applied. The error reports how many writes succeeded before the failure. Treat this command as a controlled sequence, not a database transaction wearing a hard hat.

### Watch one tag

```bash
logix-cli watch Motor.Speed --type real --interval 1s
logix-cli watch Motor.Speed --type real --duration 30s --format jsonl
logix-cli watch Motor.Speed --type real --count 10 --format jsonl
logix-cli watch ProductionCounts[0] --type dint --elements 5 --format csv
```

### Watch several tags

`watch-multi` uses the same `--item` grammar as `read-multi`:

```bash
logix-cli watch-multi \
  --item Motor.Speed=real \
  --item Counter=dint \
  --interval 1s \
  --format jsonl
```

`--count` counts poll cycles. Each cycle reads every configured item and gives the rows a common timestamp. All watch commands also accept `--duration`; whichever bound is reached first stops the stream. A zero count or duration means unlimited.

`--dry-run` may be supplied explicitly to any write command. It cannot be combined with `--yes`.

## Output contract

Snapshot commands accept:

```text
table, text, json, csv
```

Streaming commands accept:

```text
text, jsonl, csv
```

JSON is rejected for unbounded streams. JSON Lines exists because concatenating standalone JSON documents is not a format, merely optimism with braces.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | success |
| 1 | general error |
| 2 | usage or configuration error |
| 3 | transport or connection error |
| 4 | CIP request error |
| 7 | write or control rejected |
| 8 | timeout |
| 9 | output or formatting error |

## Safety and scope

`logix-cli` is an engineering and diagnostic tool. Confirm the controller, tag and expected data type before transmitting writes. The implementation intentionally excludes arbitrary CIP messages, server mode, UDT write construction and I/O adapter behaviour.
