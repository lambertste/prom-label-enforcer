package enforcer

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
)

// Enforcer validates that a given set of labels satisfies all required label
// rules defined in the configuration.
type Enforcer struct {
	cfg *Config
}

// New creates a new Enforcer from the provided Config.
func New(cfg *Config) *Enforcer {
	return &Enforcer{cfg: cfg}
}

// Validate checks whether the provided label set satisfies every rule in the
// enforcer's configuration. It returns a list of violations (one per failing
// rule). An empty slice means the label set is compliant.
func (e *Enforcer) Validate(lset labels.Labels) []string {
	var violations []string

	for _, rule := range e.cfg.Rules {
		val := lset.Get(rule.LabelName)

		// Check presence requirement.
		if rule.Required && val == "" {
			violations = append(violations,
				fmt.Sprintf("required label %q is missing", rule.LabelName))
			continue
		}

		// If the label is absent and not required, skip value checks.
		if val == "" {
			continue
		}

		// Check allowed values if specified.
		if len(rule.AllowedValues) > 0 && !contains(rule.AllowedValues, val) {
			violations = append(violations,
				fmt.Sprintf("label %q has disallowed value %q (allowed: %s)",
					rule.LabelName, val, strings.Join(rule.AllowedValues, ", ")))
		}
	}

	return violations
}

// IsCompliant returns true when the label set passes all rules.
func (e *Enforcer) IsCompliant(lset labels.Labels) bool {
	return len(e.Validate(lset)) == 0
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
