package enforcer

import (
	"testing"
)

func TestRule_Compile_Valid(t *testing.T) {
	r := &Rule{Pattern: `^prod-.*`}
	if err := r.Compile(); err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
}

func TestRule_Compile_Invalid(t *testing.T) {
	r := &Rule{Pattern: `[invalid`}
	if err := r.Compile(); err == nil {
		t.Fatal("expected compile error, got nil")
	}
}

func TestRule_MatchesValue_AllowedList(t *testing.T) {
	r := &Rule{Allowed: []string{"us-east-1", "eu-west-1"}}
	if !r.MatchesValue("us-east-1") {
		t.Error("expected us-east-1 to match")
	}
	if r.MatchesValue("ap-south-1") {
		t.Error("expected ap-south-1 not to match")
	}
}

func TestRule_MatchesValue_Pattern(t *testing.T) {
	r := &Rule{Pattern: `^prod-`}
	_ = r.Compile()
	if !r.MatchesValue("prod-api") {
		t.Error("expected prod-api to match")
	}
	if r.MatchesValue("staging-api") {
		t.Error("expected staging-api not to match")
	}
}

func TestRule_MatchesValue_AnyNonEmpty(t *testing.T) {
	r := &Rule{}
	if !r.MatchesValue("anything") {
		t.Error("expected non-empty value to match")
	}
	if r.MatchesValue("") {
		t.Error("expected empty value not to match")
	}
}
