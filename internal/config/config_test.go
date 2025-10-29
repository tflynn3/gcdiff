package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	if len(cfg.IgnoreFields) == 0 {
		t.Error("Default config should have ignore fields")
	}

	// Check that common fields are ignored
	expectedFields := []string{"id", "selfLink", "creationTimestamp", "fingerprint"}
	for _, field := range expectedFields {
		found := false
		for _, ignored := range cfg.IgnoreFields {
			if ignored == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field %q to be in default ignore list", field)
		}
	}
}

func TestLoad_NonExistent(t *testing.T) {
	cfg, err := Load("/path/that/does/not/exist.yaml")

	if err != nil {
		t.Errorf("Load should not error on non-existent file, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load should return default config for non-existent file")
	}

	// Should return default config
	if len(cfg.IgnoreFields) == 0 {
		t.Error("Should return default config with ignore fields")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")

	if err != nil {
		t.Errorf("Load should not error on empty path, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load should return default config for empty path")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `ignore_fields:
  - customField1
  - customField2
ignore_patterns:
  - ".*Custom$"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.IgnoreFields) != 2 {
		t.Errorf("Expected 2 ignore fields, got %d", len(cfg.IgnoreFields))
	}

	if cfg.IgnoreFields[0] != "customField1" {
		t.Errorf("Expected first field to be 'customField1', got %q", cfg.IgnoreFields[0])
	}

	if len(cfg.IgnorePatterns) != 1 {
		t.Errorf("Expected 1 ignore pattern, got %d", len(cfg.IgnorePatterns))
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `this is not: valid: yaml: content:`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load should error on invalid YAML")
	}
}

func TestLoad_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.yaml")

	emptyContent := ``
	if err := os.WriteFile(configPath, []byte(emptyContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should return defaults for empty config
	if len(cfg.IgnoreFields) == 0 {
		t.Error("Empty config should merge with defaults")
	}
}

func TestShouldIgnore(t *testing.T) {
	cfg := &Config{
		IgnoreFields: []string{
			"id",
			"metadata.creationTimestamp",
			"selfLink",
		},
	}

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"exact match", "id", true},
		{"nested field", "metadata.creationTimestamp", true},
		{"not in list", "name", false},
		{"partial match", "metadata", false},
		{"case sensitive", "ID", false},
		{"selfLink exact", "selfLink", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.ShouldIgnore(tt.field)
			if result != tt.expected {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestShouldIgnore_EmptyConfig(t *testing.T) {
	cfg := &Config{
		IgnoreFields: []string{},
	}

	if cfg.ShouldIgnore("anyfield") {
		t.Error("Empty config should not ignore any fields")
	}
}
