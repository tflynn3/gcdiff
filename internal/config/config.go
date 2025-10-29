package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the gcdiff configuration
type Config struct {
	// IgnoreFields is a list of field paths to ignore when comparing resources
	// Supports nested paths like "metadata.creationTimestamp"
	IgnoreFields []string `yaml:"ignore_fields"`

	// IgnorePatterns is a list of regex patterns for fields to ignore
	IgnorePatterns []string `yaml:"ignore_patterns"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		IgnoreFields: []string{
			"id",
			"selfLink",
			"creationTimestamp",
			"lastModifiedTimestamp",
			"fingerprint",
			"kind",
			"etag",
		},
		IgnorePatterns: []string{
			".*Timestamp$",
			".*Fingerprint$",
		},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	if path == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Merge with defaults if empty
	if len(cfg.IgnoreFields) == 0 && len(cfg.IgnorePatterns) == 0 {
		return Default(), nil
	}

	return &cfg, nil
}

// ShouldIgnore checks if a field should be ignored based on config
func (c *Config) ShouldIgnore(fieldPath string) bool {
	// Check exact matches
	for _, field := range c.IgnoreFields {
		if field == fieldPath {
			return true
		}
	}

	// TODO: Add regex pattern matching for IgnorePatterns
	// This would require importing regexp package

	return false
}
