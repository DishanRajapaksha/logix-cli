package cli

import (
	"fmt"
	"os"

	"github.com/DishanRajapaksha/logix-cli/internal/config"
)

func (a *App) initConfig(args []string) error {
	fs := a.newFlagSet("init-config")
	outputPath := fs.String("output", config.DefaultPath, "output config path")
	force := fs.Bool("force", false, "overwrite an existing file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("init-config takes no positional arguments")
	}
	if !*force {
		if _, err := os.Stat(*outputPath); err == nil {
			return fmt.Errorf("%w: %s already exists; use --force to overwrite", config.ErrConfig, *outputPath)
		}
	}
	if err := os.WriteFile(*outputPath, []byte(config.Marshal(config.Starter())), 0o600); err != nil {
		return fmt.Errorf("%w: write %s: %v", config.ErrConfig, *outputPath, err)
	}
	fmt.Fprintf(a.out, "wrote %s\n", *outputPath)
	return nil
}

func (a *App) validateConfig(args []string) error {
	fs := a.newFlagSet("validate-config")
	flags := addCommonFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("validate-config takes no positional arguments")
	}
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return err
	}
	_, profileName, err := cfg.Profile(flags.profile)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.out, "valid config: %s (profile %s)\n", flags.configPath, profileName)
	return nil
}
