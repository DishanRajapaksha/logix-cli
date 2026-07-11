package cli

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/DishanRajapaksha/logix-cli/internal/config"
	"github.com/DishanRajapaksha/logix-cli/internal/logixclient"
)

type commonFlags struct {
	configPath string
	profile    string
	address    string
	port       uint
	path       string
	pathSet    bool
	timeout    time.Duration
	format     string
	verbose    bool
	debug      bool
}

func addCommonFlags(fs *flag.FlagSet, snapshot bool) *commonFlags {
	flags := &commonFlags{}
	fs.StringVar(&flags.configPath, "config", config.DefaultPath, "YAML config file")
	fs.StringVar(&flags.profile, "profile", "", "config profile name")
	fs.StringVar(&flags.address, "address", "", "controller IPv4 address")
	fs.UintVar(&flags.port, "port", 0, "EtherNet/IP port")
	fs.Func("path", "CIP route path", func(value string) error {
		flags.path = value
		flags.pathSet = true
		return nil
	})
	fs.DurationVar(&flags.timeout, "timeout", 0, "socket timeout")
	if snapshot {
		fs.StringVar(&flags.format, "format", "table", "output format")
	} else {
		fs.StringVar(&flags.format, "format", "text", "stream output format")
	}
	fs.BoolVar(&flags.verbose, "verbose", false, "print connection decisions")
	fs.BoolVar(&flags.debug, "debug", false, "enable client diagnostics")
	return flags
}

func (a *App) options(flags *commonFlags) (logixclient.Options, string, error) {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		if flags.address == "" {
			return logixclient.Options{}, "", err
		}
		cfg = config.Starter()
	}
	profile, profileName, err := cfg.Profile(flags.profile)
	if err != nil {
		return logixclient.Options{}, "", err
	}
	if flags.address != "" {
		profile.Address = flags.address
	}
	if flags.port != 0 {
		profile.Port = flags.port
	}
	if flags.pathSet {
		profile.Path = flags.path
	}
	if flags.timeout != 0 {
		profile.Timeout = flags.timeout
	}
	options := logixclient.Options{Address: profile.Address, Port: profile.Port, Path: profile.Path, Timeout: profile.Timeout, Debug: flags.debug}
	if err := logixclient.ValidateOptions(options); err != nil {
		return logixclient.Options{}, "", err
	}
	if flags.verbose {
		fmt.Fprintf(a.err, "profile=%s address=%s port=%d path=%q timeout=%s\n", profileName, options.Address, options.Port, options.Path, options.Timeout)
	}
	return options, profileName, nil
}

func (a *App) connect(flags *commonFlags) (logixclient.Client, logixclient.Options, error) {
	options, _, err := a.options(flags)
	if err != nil {
		return nil, logixclient.Options{}, err
	}
	client, err := a.factory.New(options)
	if err != nil {
		return nil, logixclient.Options{}, err
	}
	if err := client.Connect(); err != nil {
		return nil, logixclient.Options{}, err
	}
	return client, options, nil
}

func closeClient(client logixclient.Client) {
	_ = client.Disconnect()
}

func uint32String(v uint32) string { return strconv.FormatUint(uint64(v), 10) }
