package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DishanRajapaksha/logix-cli/internal/config"
	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
	"github.com/DishanRajapaksha/logix-cli/internal/output"
)

const (
	appName           = "logix-cli"
	exitSuccess       = 0
	exitGeneralError  = 1
	exitConfigError   = 2
	exitConnection    = 3
	exitRequestError  = 4
	exitWriteRejected = 7
	exitTimeout       = 8
	exitOutputError   = 9
)

var errWriteRejected = errors.New("write rejected")

type App struct {
	out     io.Writer
	err     io.Writer
	factory logixclient.Factory
}

func NewApp(out, err io.Writer) *App {
	return &App{out: out, err: err, factory: logixclient.GoLogixFactory{}}
}

func NewAppWithFactory(out, err io.Writer, factory logixclient.Factory) *App {
	return &App{out: out, err: err, factory: factory}
}

func Main() {
	code := NewApp(os.Stdout, os.Stderr).Run(os.Args[1:])
	if code != 0 {
		os.Exit(code)
	}
}

func (a *App) Run(args []string) int {
	normalised, err := normaliseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(a.err, err)
		return exitConfigError
	}
	args = normalised
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		a.printUsage()
		return exitSuccess
	}
	if args[0] == "version" || args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintf(a.out, "%s development\n", appName)
		return exitSuccess
	}
	switch args[0] {
	case "init-config":
		err = a.initConfig(args[1:])
	case "validate-config":
		err = a.validateConfig(args[1:])
	case "test-connection":
		err = a.testConnection(args[1:])
	case "status":
		err = a.status(args[1:])
	case "identify":
		err = a.identify(args[1:])
	case "programs":
		err = a.programs(args[1:])
	case "tags":
		err = a.tags(args[1:])
	case "points":
		err = a.points(args[1:])
	case "read":
		err = a.read(args[1:])
	case "read-multi":
		err = a.readMulti(args[1:])
	case "read-point":
		err = a.readPoint(args[1:])
	case "write":
		err = a.write(args[1:])
	case "write-multi":
		err = a.writeMulti(args[1:])
	case "write-point":
		err = a.writePoint(args[1:])
	case "watch":
		err = a.watch(args[1:])
	case "watch-multi":
		err = a.watchMulti(args[1:])
	case "watch-point":
		err = a.watchPoint(args[1:])
	case "completions":
		err = a.completions(args[1:])
	default:
		a.printUsage()
		fmt.Fprintf(a.err, "unknown command %q\n", args[0])
		return exitGeneralError
	}
	if err == nil {
		return exitSuccess
	}
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	fmt.Fprintln(a.err, err)
	return mapExitCode(err)
}

func mapExitCode(err error) int {
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return exitSuccess
	case errors.Is(err, errWriteRejected):
		return exitWriteRejected
	case errors.Is(err, config.ErrConfig), errors.Is(err, logixclient.ErrValidation):
		return exitConfigError
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(strings.ToLower(err.Error()), "timeout"):
		return exitTimeout
	case errors.Is(err, output.ErrOutput):
		return exitOutputError
	case errors.Is(err, logixclient.ErrConnection):
		return exitConnection
	case errors.Is(err, logixclient.ErrRequest):
		return exitRequestError
	case strings.Contains(err.Error(), "flag provided but not defined"), strings.Contains(err.Error(), "invalid value"), strings.Contains(err.Error(), "requires"):
		return exitConfigError
	default:
		return exitGeneralError
	}
}

func (a *App) newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.err)
	return fs
}

func (a *App) printUsage() {
	fmt.Fprintln(a.out, `logix-cli is a script-friendly Rockwell Logix tag client over EtherNet/IP.

Usage:
  logix-cli [global flags] <command> [flags]
  logix-cli init-config
  logix-cli validate-config --profile local
  logix-cli test-connection --address 192.168.1.10
  logix-cli status --format json
  logix-cli identify
  logix-cli programs --format json
  logix-cli tags --filter Motor --limit 50
  logix-cli points --format json
  logix-cli read Motor.Speed --type real
  logix-cli read-multi --item Motor.Speed=real --item Counter=dint
  logix-cli read-point motor_speed
  logix-cli write Motor.Enable --type bool --value true --yes
  logix-cli write-multi --set Motor.Enable=bool:true --set Recipe=dint:12 --yes
  logix-cli write-point motor_enabled --value true --yes
  logix-cli watch Motor.Speed --type real --interval 1s --format jsonl
  logix-cli watch-multi --item Motor.Speed=real --item Counter=dint --format jsonl
  logix-cli watch-point motor_speed --format jsonl
  logix-cli completions zsh
  logix-cli version

Commands:
  init-config       Write a starter YAML config file
  validate-config  Validate local config without connecting
  test-connection  Open and close a CIP connection
  status           Read controller product, revision, and Identity status
  identify         Read the complete CIP Identity object
  programs         List controller programs
  tags             List controller and program-scoped tags
  points           List configured named points
  read             Read one tag, with optional type detection
  read-multi       Read several typed tags over one connection
  read-point       Read a configured named point
  write            Write one tag; dry-run unless --yes is supplied
  write-multi      Write several typed tags; dry-run unless --yes is supplied
  write-point      Write a configured named point; dry-run unless --yes is supplied
  watch            Poll one tag repeatedly
  watch-multi      Poll several typed tags repeatedly
  watch-point      Poll a configured named point repeatedly
  completions      Generate Bash or Zsh completion scripts
  version          Print version information

Common flags:
  --config    YAML config file, defaults to config.yaml
  --profile   Config profile name
  --address   Controller IPv4 address
  --port      EtherNet/IP port, defaults to 44818
  --path      CIP route path, such as 1,0; empty for Micro800 devices
  --timeout   Socket timeout
  --format    snapshots: table, text, json, csv; streams: text, jsonl, csv
  --verbose   Print high-level connection decisions
  --debug     Enable lower-level client diagnostics`)
}

func normaliseGlobalFlags(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}
	var globals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 >= len(args) {
				return nil, errors.New("command is required after --")
			}
			return appendCommandGlobals(args[i+1:], globals), nil
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			return appendCommandGlobals(args[i:], globals), nil
		}
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			return args[i:], nil
		}
		name, inline, hasInline := strings.Cut(arg, "=")
		switch name {
		case "--verbose", "--debug":
			if hasInline {
				return nil, fmt.Errorf("%s does not take a value", name)
			}
			globals = append(globals, name)
		case "--config", "--profile", "--address", "--port", "--path", "--timeout", "--format":
			value := inline
			if !hasInline {
				i++
				if i >= len(args) || strings.HasPrefix(args[i], "-") {
					return nil, fmt.Errorf("%s requires a value", name)
				}
				value = args[i]
			}
			if value == "" && name != "--path" {
				return nil, fmt.Errorf("%s requires a value", name)
			}
			globals = append(globals, name, value)
		default:
			return nil, fmt.Errorf("unknown global flag %q", name)
		}
	}
	return nil, errors.New("command is required")
}

func appendCommandGlobals(args, globals []string) []string {
	if len(args) == 0 || len(globals) == 0 || !commandSupportsGlobals(args[0]) {
		return args
	}
	out := make([]string, 0, len(args)+len(globals))
	out = append(out, args[0])
	if commandTakesTag(args[0]) && len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		out = append(out, args[1])
		out = append(out, globals...)
		out = append(out, args[2:]...)
		return out
	}
	out = append(out, globals...)
	out = append(out, args[1:]...)
	return out
}

func commandTakesTag(command string) bool {
	switch command {
	case "read", "write", "watch", "read-point", "write-point", "watch-point":
		return true
	default:
		return false
	}
}

func commandSupportsGlobals(command string) bool {
	switch command {
	case "validate-config", "test-connection", "status", "identify", "programs", "tags", "points", "read", "read-multi", "read-point", "write", "write-multi", "write-point", "watch", "watch-multi", "watch-point":
		return true
	default:
		return false
	}
}
