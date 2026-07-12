package cli

import (
	"fmt"
	"strings"

	"github.com/DishanRajapaksha/industrial-cli-kit/completion"
)

func (a *App) completions(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("completions requires bash or zsh")
	}
	return completion.Write(a.out, strings.ToLower(strings.TrimSpace(args[0])), cliRegistry)
}
