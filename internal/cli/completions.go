package cli

import "fmt"

func (a *App) completions(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("completions requires bash or zsh")
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(a.out, `# bash completion for logix-cli
_logix_cli() {
  local commands="init-config validate-config test-connection status identify programs tags points groups read read-multi read-point read-group write write-multi write-point write-group watch watch-multi watch-point watch-group completions version"
  if [[ ${COMP_CWORD} -eq 1 ]]; then COMPREPLY=( $(compgen -W "$commands" -- "${COMP_WORDS[1]}") ); fi
}
complete -F _logix_cli logix-cli
`)
	case "zsh":
		fmt.Fprint(a.out, `#compdef logix-cli
_arguments '1:command:(init-config validate-config test-connection status identify programs tags points groups read read-multi read-point read-group write write-multi write-point write-group watch watch-multi watch-point watch-group completions version)' '*::arg:->args'
`)
	default:
		return fmt.Errorf("unsupported shell %q; use bash or zsh", args[0])
	}
	return nil
}
