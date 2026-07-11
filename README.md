# logix-cli

`logix-cli` is a script-friendly command-line client for reading and writing tags on Rockwell Automation ControlLogix, CompactLogix and compatible CIP-based controllers over EtherNet/IP.

It follows the command and output conventions used by `opc-xml-da-cli`, `opc-ua-cli`, `iec-104-cli` and `modbus-cli`.

The implementation uses [`github.com/danomagnum/gologix`](https://github.com/danomagnum/gologix). PLC-5, SLC and MicroLogix controllers that require PCCC are outside the project scope.

## Features

- YAML profiles with command-line overrides
- connection diagnostics
- CIP Identity and controller-status inspection
- controller program discovery
- controller and program-scoped tag discovery
- typed scalar and array reads
- heterogeneous multi-tag reads over one connection
- safety-gated single and multi-tag writes, with dry-run as the default
- single and multi-tag polling with text, JSON Lines or CSV output
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
```

The usual ControlLogix/CompactLogix route is `1,0`, meaning backplane, slot 0. Use an empty path for devices such as some Micro800 controllers:

```yaml
  micro800:
    address: 192.168.1.20
    port: 44818
    path: ""
    timeout: 5s
```

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

`--count` counts poll cycles. Each cycle reads every configured item and gives the rows a common timestamp.

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
