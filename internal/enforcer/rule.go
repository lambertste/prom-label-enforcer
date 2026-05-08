package enforcer

import "regexp"

// Rule represents a single label enforcement rule.
type Rule struct {
	Label    string   `yaml:"label"`
	Required bool     `yaml:"required"`
	Allowed  []string `yaml:"allowed,omitempty"`
	Pattern  string   `yaml:"pattern,omitempty"`

	compiledPattern *regexp.Regexp
}

// Compile pre-compiles the pattern regex if one is set.
func (r *Rule) Compile() error {
	if r.Pattern == "" {
		return nil
	}
	cp, err := regexp.Compile(r.Pattern)
	if err != nil {
		return err
	}
	r.compiledPattern = cp
	return nil
}

// MatchesValue returns true when the given value satisfies the rule constraints.
// An empty Allowed list and no Pattern means any non-empty value is accepted.
func (r *Rule) MatchesValue(value string) bool {
	if len(r.Allowed) > 0 {
		for _, a := range r.Allowed {
			if a == value {
				return true
			}
		}
		return false
	}
	if r.compiledPattern != nil {
		return r.compiledPattern.MatchString(value)
	}
	return value != ""
}
