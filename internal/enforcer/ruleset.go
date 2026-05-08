package enforcer

import "fmt"

// RuleSet holds a named collection of Rules and provides batch validation.
type RuleSet struct {
	Name  string  `yaml:"name"`
	Rules []*Rule `yaml:"rules"`
}

// Compile compiles all patterns in the rule set.
func (rs *RuleSet) Compile() error {
	for _, r := range rs.Rules {
		if err := r.Compile(); err != nil {
			return fmt.Errorf("rule %q: %w", r.Label, err)
		}
	}
	return nil
}

// Validate checks a label map against the rule set.
// It returns a slice of human-readable violation strings (empty means compliant).
func (rs *RuleSet) Validate(labels map[string]string) []string {
	var violations []string
	for _, r := range rs.Rules {
		value, present := labels[r.Label]
		if r.Required && !present {
			violations = append(violations, fmt.Sprintf("missing required label %q", r.Label))
			continue
		}
		if present && !r.MatchesValue(value) {
			violations = append(violations, fmt.Sprintf("label %q has disallowed value %q", r.Label, value))
		}
	}
	return violations
}
