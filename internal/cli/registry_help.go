package cli

import sharedhelp "github.com/DishanRajapaksha/industrial-cli-kit/help"

func (a *App) writeRegistryUsage() {
	_ = sharedhelp.Write(a.out, cliRegistry, sharedhelp.Options{
		Description: "logix-cli is a script-friendly Rockwell Logix tag client over EtherNet/IP.",
		Usage:       []string{"logix-cli [global flags] <command> [flags]"},
		Examples: []string{
			"logix-cli init-config",
			"logix-cli test-connection --address 192.168.1.10",
			"logix-cli status --format json",
			"logix-cli programs --format json",
			"logix-cli tags --filter Motor --limit 50",
			"logix-cli read Motor.Speed --type real",
			"logix-cli read-multi --item Motor.Speed=real --item Counter=dint",
			"logix-cli write Motor.Enable --type bool --value true --yes",
			"logix-cli write-multi --set Motor.Enable=bool:true --set Recipe=dint:12 --yes",
			"logix-cli watch Motor.Speed --type real --interval 1s --duration 30s --format jsonl",
			"logix-cli watch-group motor --duration 30s --format jsonl",
			"logix-cli completions zsh",
		},
	})
}
