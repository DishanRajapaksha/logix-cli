package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var ErrConfig = errors.New("configuration error")

const DefaultPath = "config.yaml"

type Config struct {
	DefaultProfile string             `yaml:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Points         []Point            `yaml:"points,omitempty"`
}

type Profile struct {
	Address string
	Port    uint
	Path    string
	Timeout time.Duration
}

type Point struct {
	Name        string `yaml:"name" json:"name"`
	Tag         string `yaml:"tag" json:"tag"`
	Type        string `yaml:"type,omitempty" json:"type"`
	Elements    uint16 `yaml:"elements,omitempty" json:"elements"`
	Unit        string `yaml:"unit,omitempty" json:"unit,omitempty"`
	Writable    bool   `yaml:"writable,omitempty" json:"writable"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func (p *Profile) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("profile must be a mapping")
	}
	allowed := map[string]struct{}{
		"address": {}, "port": {}, "path": {}, "timeout": {},
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unknown profile field %q", key)
		}
	}
	var raw struct {
		Address string  `yaml:"address"`
		Port    *uint   `yaml:"port"`
		Path    *string `yaml:"path"`
		Timeout string  `yaml:"timeout"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	p.Address = raw.Address
	p.Port = 44818
	p.Path = "1,0"
	p.Timeout = 5 * time.Second
	if raw.Port != nil {
		p.Port = *raw.Port
	}
	if raw.Path != nil {
		p.Path = *raw.Path
	}
	if raw.Timeout != "" {
		parsed, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", raw.Timeout, err)
		}
		p.Timeout = parsed
	}
	return nil
}

func (p Profile) MarshalYAML() (any, error) {
	return struct {
		Address string `yaml:"address"`
		Port    uint   `yaml:"port"`
		Path    string `yaml:"path"`
		Timeout string `yaml:"timeout"`
	}{
		Address: p.Address,
		Port:    p.Port,
		Path:    p.Path,
		Timeout: p.Timeout.String(),
	}, nil
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
		Points: []Point{
			{Name: "motor_speed", Tag: "Motor.Speed", Type: "real", Elements: 1, Unit: "rpm"},
			{Name: "motor_enabled", Tag: "Motor.Enable", Type: "bool", Elements: 1, Writable: true},
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
	seen := map[string]struct{}{}
	for _, declared := range c.Points {
		point := normalisePoint(declared)
		if point.Name == "" {
			return fmt.Errorf("%w: point name is required", ErrConfig)
		}
		key := strings.ToLower(point.Name)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: duplicate point name %q", ErrConfig, point.Name)
		}
		seen[key] = struct{}{}
		if point.Tag == "" {
			return fmt.Errorf("%w: point %q tag is required", ErrConfig, point.Name)
		}
		if !validPointType(point.Type) {
			return fmt.Errorf("%w: point %q has unsupported type %q", ErrConfig, point.Name, point.Type)
		}
		if point.Elements == 0 {
			return fmt.Errorf("%w: point %q elements must be between 1 and 65535", ErrConfig, point.Name)
		}
		if point.Writable && point.Type == "auto" {
			return fmt.Errorf("%w: writable point %q requires an explicit type", ErrConfig, point.Name)
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

func (c Config) Point(name string) (Point, error) {
	name = strings.TrimSpace(name)
	for _, declared := range c.Points {
		point := normalisePoint(declared)
		if strings.EqualFold(point.Name, name) {
			return point, nil
		}
	}
	return Point{}, fmt.Errorf("%w: point %q does not exist", ErrConfig, name)
}

func (c Config) NormalisedPoints() []Point {
	points := make([]Point, len(c.Points))
	for i, point := range c.Points {
		points[i] = normalisePoint(point)
	}
	return points
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
	var cfg Config
	decoder := yaml.NewDecoder(strings.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("%w: parse YAML: %v", ErrConfig, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("%w: multiple YAML documents are not supported", ErrConfig)
		}
		return Config{}, fmt.Errorf("%w: parse trailing YAML: %v", ErrConfig, err)
	}
	return cfg, nil
}

func Marshal(c Config) string {
	data, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("marshal configuration: %v", err))
	}
	return string(data)
}

func normalisePoint(point Point) Point {
	point.Name = strings.TrimSpace(point.Name)
	point.Tag = strings.TrimSpace(point.Tag)
	point.Type = strings.ToLower(strings.TrimSpace(point.Type))
	if point.Type == "" {
		point.Type = "auto"
	}
	if point.Elements == 0 {
		point.Elements = 1
	}
	point.Unit = strings.TrimSpace(point.Unit)
	point.Description = strings.TrimSpace(point.Description)
	return point
}

func validPointType(value string) bool {
	switch value {
	case "auto", "bool", "sint", "int", "dint", "lint", "usint", "uint", "udint", "ulint", "real", "lreal", "string":
		return true
	default:
		return false
	}
}
