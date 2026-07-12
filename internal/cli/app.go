package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
	"github.com/DishanRajapaksha/industrial-cli-kit/exitcode"
	"github.com/DishanRajapaksha/logix-cli/internal/config"
	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
	"github.com/DishanRajapaksha/logix-cli/internal/output"
)

const (
	appName           = "logix-cli"
	exitSuccess       = int(exitcode.Success)
	exitGeneralError  = int(exitcode.General)
	exitConfigError   = int(exitcode.Config)
	exitConnection    = int(exitcode.Connection)
	exitRequestError  = int(exitcode.Request)
	exitWriteRejected = int(exitcode.Rejected)
	exitTimeout       = int(exitcode.Timeout)
	exitOutputError   = int(exitcode.Output)
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
	case "groups":
		err = a.groups(args[1:])
	case "read":
		err = a.read(args[1:])
	case "read-multi":
		err = a.readMulti(args[1:])
	case "read-point":
		err = a.readPoint(args[1:])
	case "read-group":
		err = a.readGroup(args[1:])
	case "write":
		err = a.write(args[1:])
	case "write-multi":
		err = a.writeMulti(args[1:])
	case "write-point":
		err = a.writePoint(args[1:])
	case "write-group":
		err = a.writeGroup(args[1:])
	case "watch":
		err = a.watch(args[1:])
	case "watch-multi":
		err = a.watchMulti(args[1:])
	case "watch-point":
		err = a.watchPoint(args[1:])
	case "watch-group":
		err = a.watchGroup(args[1:])
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
	a.writeRegistryUsage()
}

func normaliseGlobalFlags(args []string) ([]string, error) {
	return command.NormalizeGlobalFlagsForRegistry(args, cliRegistry)
}
