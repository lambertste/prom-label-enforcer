package enforcer

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
)

func baseConfig() *Config {
	return &Config{
		Rules: []Rule{
			{LabelName: "env", Required: true, AllowedValues: []string{"prod", "staging", "dev"}},
			{LabelName: "team", Required: true, AllowedValues: nil},
			{LabelName: "region", Required: false, AllowedValues: []string{"us-east-1", "eu-west-1"}},
		},
	}
}

func TestEnforcer_Compliant(t *testing.T) {
	e := New(baseConfig())
	lset := labels.FromStrings("env", "prod", "team", "platform", "region", "us-east-1")
	if !e.IsCompliant(lset) {
		t.Errorf("expected compliant label set, got violations: %v", e.Validate(lset))
	}
}

func TestEnforcer_MissingRequiredLabel(t *testing.T) {
	e := New(baseConfig())
	lset := labels.FromStrings("team", "platform")
	violations := e.Validate(lset)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
}

func TestEnforcer_DisallowedValue(t *testing.T) {
	e := New(baseConfig())
	lset := labels.FromStrings("env", "canary", "team", "platform")
	violations := e.Validate(lset)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
}

func TestEnforcer_OptionalLabelAbsent(t *testing.T) {
	e := New(baseConfig())
	// region is optional — omitting it should not cause a violation.
	lset := labels.FromStrings("env", "dev", "team", "infra")
	if !e.IsCompliant(lset) {
		t.Errorf("expected compliant label set, got violations: %v", e.Validate(lset))
	}
}

func TestEnforcer_OptionalLabelDisallowedValue(t *testing.T) {
	e := New(baseConfig())
	lset := labels.FromStrings("env", "staging", "team", "infra", "region", "ap-southeast-1")
	violations := e.Validate(lset)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for disallowed region, got %d: %v", len(violations), violations)
	}
}

func TestEnforcer_MultipleViolations(t *testing.T) {
	e := New(baseConfig())
	// Missing both required labels.
	lset := labels.Labels{}
	violations := e.Validate(lset)
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(violations), violations)
	}
}
