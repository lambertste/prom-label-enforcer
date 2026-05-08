package enforcer

import "testing"

func buildRuleSet() *RuleSet {
	return &RuleSet{
		Name: "test-set",
		Rules: []*Rule{
			{Label: "env", Required: true, Allowed: []string{"prod", "staging"}},
			{Label: "region", Required: true, Pattern: `^[a-z]+-[a-z]+-[0-9]+$`},
			{Label: "owner", Required: false},
		},
	}
}

func TestRuleSet_Compile(t *testing.T) {
	rs := buildRuleSet()
	if err := rs.Compile(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRuleSet_Compile_BadPattern(t *testing.T) {
	rs := &RuleSet{Rules: []*Rule{{Label: "x", Pattern: `[bad`}}}
	if err := rs.Compile(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRuleSet_Validate_Compliant(t *testing.T) {
	rs := buildRuleSet()
	_ = rs.Compile()
	violations := rs.Validate(map[string]string{
		"env":    "prod",
		"region": "us-east-1",
	})
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got: %v", violations)
	}
}

func TestRuleSet_Validate_MissingRequired(t *testing.T) {
	rs := buildRuleSet()
	_ = rs.Compile()
	violations := rs.Validate(map[string]string{"env": "prod"})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
}

func TestRuleSet_Validate_DisallowedValue(t *testing.T) {
	rs := buildRuleSet()
	_ = rs.Compile()
	violations := rs.Validate(map[string]string{
		"env":    "dev",
		"region": "us-east-1",
	})
	if len(violations) == 0 {
		t.Fatal("expected at least one violation")
	}
}

func TestRuleSet_Validate_OptionalAbsent(t *testing.T) {
	rs := buildRuleSet()
	_ = rs.Compile()
	violations := rs.Validate(map[string]string{
		"env":    "staging",
		"region": "eu-west-1",
	})
	if len(violations) != 0 {
		t.Fatalf("optional absent label should not violate: %v", violations)
	}
}
