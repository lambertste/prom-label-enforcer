package enforcer

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadConfig_Valid(t *testing.T) {
	content := `
label_rules:
  - name: env
    required: true
    allowed_values: ["prod", "staging", "dev"]
  - name: team
    required: true
`
	path := writeTemp(t, content)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.LabelRules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.LabelRules))
	}
	if cfg.LabelRules[0].Name != "env" {
		t.Errorf("expected first rule name 'env', got %q", cfg.LabelRules[0].Name)
	}
	if len(cfg.LabelRules[0].AllowedValues) != 3 {
		t.Errorf("expected 3 allowed values, got %d", len(cfg.LabelRules[0].AllowedValues))
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	_, err := LoadConfig("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_DuplicateLabel(t *testing.T) {
	content := `
label_rules:
  - name: env
    required: true
  - name: env
    required: false
`
	path := writeTemp(t, content)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for duplicate label name, got nil")
	}
}

func TestLoadConfig_EmptyLabelName(t *testing.T) {
	content := `
label_rules:
  - name: ""
    required: true
`
	path := writeTemp(t, content)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for empty label name, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `
label_rules:
  - name: env
    required: [this is not valid yaml
`
	path := writeTemp(t, content)
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
