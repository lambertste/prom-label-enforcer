package enforcer

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuleSetFile is the top-level structure for a rule-set YAML file.
type RuleSetFile struct {
	RuleSets []*RuleSet `yaml:"rule_sets"`
}

// LoadRuleSets reads and compiles rule sets from a YAML file at path.
func LoadRuleSets(path string) ([]*RuleSet, error) {
	if path == "" {
		return nil, fmt.Errorf("rule set path must not be empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rule sets file: %w", err)
	}
	var rsf RuleSetFile
	if err := yaml.Unmarshal(data, &rsf); err != nil {
		return nil, fmt.Errorf("parsing rule sets file: %w", err)
	}
	for _, rs := range rsf.RuleSets {
		if rs.Name == "" {
			return nil, fmt.Errorf("every rule_set must have a name")
		}
		if err := rs.Compile(); err != nil {
			return nil, fmt.Errorf("compiling rule set %q: %w", rs.Name, err)
		}
	}
	return rsf.RuleSets, nil
}
