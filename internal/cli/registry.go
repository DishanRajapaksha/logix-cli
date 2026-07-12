package cli

import "github.com/DishanRajapaksha/industrial-cli-kit/command"

var cliRegistry = command.Registry{
	Binary: appName,
	GlobalFlags: []command.Flag{
		{Name: "config", TakesValue: true, Summary: "YAML config file"},
		{Name: "profile", TakesValue: true, Summary: "config profile name"},
		{Name: "address", TakesValue: true, Summary: "controller IPv4 address"},
		{Name: "port", TakesValue: true, Summary: "EtherNet/IP port"},
		{Name: "path", TakesValue: true, AllowEmpty: true, Summary: "CIP route path"},
		{Name: "timeout", TakesValue: true, Summary: "socket timeout"},
		{Name: "format", TakesValue: true, Summary: "output format"},
		{Name: "verbose", Summary: "print connection decisions"},
		{Name: "debug", Summary: "enable client diagnostics"},
	},
	Commands: []command.Command{
		{Name: "init-config", Summary: "Write a starter YAML config file", Flags: registryFlags("output", "force")},
		{Name: "validate-config", Summary: "Validate local config without connecting"},
		{Name: "test-connection", Summary: "Open and close a CIP connection"},
		{Name: "status", Summary: "Read controller identity status"},
		{Name: "identify", Summary: "Read the complete CIP Identity object"},
		{Name: "programs", Summary: "List controller programs"},
		{Name: "tags", Summary: "List controller and program-scoped tags", Flags: registryFlags("filter", "limit", "program")},
		{Name: "points", Summary: "List configured named points"},
		{Name: "groups", Summary: "List configured point groups"},
		{Name: "read", Summary: "Read one tag", LeadingArgs: 1, Flags: registryFlags("type")},
		{Name: "read-multi", Summary: "Read several typed tags", Flags: registryFlags("item")},
		{Name: "read-point", Summary: "Read a configured named point", LeadingArgs: 1},
		{Name: "read-group", Summary: "Read a configured point group", LeadingArgs: 1},
		{Name: "write", Summary: "Write one tag", LeadingArgs: 1, Flags: registryFlags("type", "value", "yes", "dry-run")},
		{Name: "write-multi", Summary: "Write several typed tags", Flags: registryFlags("set", "yes", "dry-run")},
		{Name: "write-point", Summary: "Write a configured named point", LeadingArgs: 1, Flags: registryFlags("value", "yes", "dry-run")},
		{Name: "write-group", Summary: "Write selected points in a group", LeadingArgs: 1, Flags: registryFlags("set", "yes", "dry-run")},
		{Name: "watch", Summary: "Poll one tag", LeadingArgs: 1, Flags: registryFlags("type", "interval", "count", "duration")},
		{Name: "watch-multi", Summary: "Poll several typed tags", Flags: registryFlags("item", "interval", "count", "duration")},
		{Name: "watch-point", Summary: "Poll a configured named point", LeadingArgs: 1, Flags: registryFlags("interval", "count", "duration")},
		{Name: "watch-group", Summary: "Poll a configured point group", LeadingArgs: 1, Flags: registryFlags("interval", "count", "duration")},
		{Name: "completions", Summary: "Generate Bash or Zsh completion scripts", LeadingArgs: 1},
		{Name: "help", Summary: "Print help"},
		{Name: "version", Summary: "Print version information"},
	},
}

func registryFlags(names ...string) []command.Flag {
	flags := make([]command.Flag, 0, len(names))
	for _, name := range names {
		takesValue := name != "force" && name != "yes" && name != "dry-run"
		flags = append(flags, command.Flag{Name: name, TakesValue: takesValue})
	}
	return flags
}
