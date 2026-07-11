package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var ErrConfig = errors.New("configuration error")

const DefaultPath = "config.yaml"

type Config struct {
	DefaultProfile string
	Profiles       map[string]Profile
}

type Profile struct {
	Address string
	Port    uint
	Path    string
	Timeout time.Duration
}

func Starter() Config {
	return Config{
		DefaultProfile: "local",
		Profiles: map[string]Profile{
			"local": {
				Address: "192.168.1.10",
				Port:    44818,
				Path:    "1,0",
				Timeout: 5 * time.Second,
			},
		},
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.DefaultProfile) == "" {
		return fmt.Errorf("%w: default_profile is required", ErrConfig)
	}
	if len(c.Profiles) == 0 {
		return fmt.Errorf("%w: at least one profile is required", ErrConfig)
	}
	if _, ok := c.Profiles[c.DefaultProfile]; !ok {
		return fmt.Errorf("%w: default profile %q does not exist", ErrConfig, c.DefaultProfile)
	}
	for name, profile := range c.Profiles {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("%w: profile name cannot be empty", ErrConfig)
		}
		if strings.TrimSpace(profile.Address) == "" {
			return fmt.Errorf("%w: profile %q address is required", ErrConfig, name)
		}
		if profile.Port == 0 || profile.Port > 65535 {
			return fmt.Errorf("%w: profile %q port must be between 1 and 65535", ErrConfig, name)
		}
		if profile.Timeout <= 0 {
			return fmt.Errorf("%w: profile %q timeout must be positive", ErrConfig, name)
		}
	}
	return nil
}

func (c Config) Profile(name string) (Profile, string, error) {
	if strings.TrimSpace(name) == "" {
		name = c.DefaultProfile
	}
	profile, ok := c.Profiles[name]
	if !ok {
		return Profile{}, "", fmt.Errorf("%w: profile %q does not exist", ErrConfig, name)
	}
	return profile, name, nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("%w: read %s: %v", ErrConfig, path, err)
	}
	cfg, err := Parse(string(data))
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Parse(data string) (Config, error) {
	cfg := Config{Profiles: map[string]Profile{}}
	var current string
	for lineNumber, raw := range strings.Split(data, "\n") {
		line := strings.TrimRight(raw, " \t\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return Config{}, fmt.Errorf("%w: line %d must contain ':'", ErrConfig, lineNumber+1)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"")

		switch {
		case indent == 0 && key == "default_profile":
			cfg.DefaultProfile = value
		case indent == 0 && key == "profiles" && value == "":
			continue
		case indent == 2 && value == "":
			current = key
			cfg.Profiles[current] = Profile{Port: 44818, Path: "1,0", Timeout: 5 * time.Second}
		case indent == 4 && current != "":
			profile := cfg.Profiles[current]
			switch key {
			case "address":
				profile.Address = value
			case "port":
				parsed, err := strconv.ParseUint(value, 10, 16)
				if err != nil {
					return Config{}, fmt.Errorf("%w: line %d invalid port %q", ErrConfig, lineNumber+1, value)
				}
				profile.Port = uint(parsed)
			case "path":
				profile.Path = value
			case "timeout":
				parsed, err := time.ParseDuration(value)
				if err != nil {
					return Config{}, fmt.Errorf("%w: line %d invalid timeout %q", ErrConfig, lineNumber+1, value)
				}
				profile.Timeout = parsed
			default:
				return Config{}, fmt.Errorf("%w: line %d unknown profile field %q", ErrConfig, lineNumber+1, key)
			}
			cfg.Profiles[current] = profile
		default:
			return Config{}, fmt.Errorf("%w: line %d has unsupported indentation or field", ErrConfig, lineNumber+1)
		}
	}
	return cfg, nil
}

func Marshal(c Config) string {
	var b strings.Builder
	fmt.Fprintf(&b, "default_profile: %s\n", c.DefaultProfile)
	b.WriteString("profiles:\n")
	order := make([]string, 0, len(c.Profiles))
	if _, ok := c.Profiles[c.DefaultProfile]; ok {
		order = append(order, c.DefaultProfile)
	}
	for name := range c.Profiles {
		if name != c.DefaultProfile {
			order = append(order, name)
		}
	}
	for _, name := range order {
		p := c.Profiles[name]
		fmt.Fprintf(&b, "  %s:\n", name)
		fmt.Fprintf(&b, "    address: %s\n", p.Address)
		fmt.Fprintf(&b, "    port: %d\n", p.Port)
		fmt.Fprintf(&b, "    path: \"%s\"\n", p.Path)
		fmt.Fprintf(&b, "    timeout: %s\n", p.Timeout)
	}
	return b.String()
}
