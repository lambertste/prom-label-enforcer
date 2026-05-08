package enforcer

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// LabelRule defines a required label and its optional allowed values.
type LabelRule struct {
	Name          string   `yaml:"name"`
	Required      bool     `yaml:"required"`
	AllowedValues []string `yaml:"allowed_values,omitempty"`
}

// Config holds the enforcer configuration loaded from a YAML file.
type Config struct {
	LabelRules []LabelRule `yaml:"label_rules"`
}

// LoadConfig reads and parses the YAML configuration file at the given path.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path must not be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks that each label rule has a non-empty name.
func (c *Config) validate() error {
	seen := make(map[string]struct{}, len(c.LabelRules))
	for i, rule := range c.LabelRules {
		if rule.Name == "" {
			return fmt.Errorf("label_rules[%d]: name must not be empty", i)
		}
		if _, dup := seen[rule.Name]; dup {
			return fmt.Errorf("label_rules[%d]: duplicate label name %q", i, rule.Name)
		}
		seen[rule.Name] = struct{}{}
	}
	return nil
}
