package compare

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tflynn3/gcdiff/internal/config"
)

// TestEndToEnd_ComputeInstanceDiff simulates comparing two compute instances
func TestEndToEnd_ComputeInstanceDiff(t *testing.T) {
	// Simulated compute instance data (simplified)
	instance1 := map[string]interface{}{
		"id":                "1234567890",
		"name":              "prod-web-01",
		"machineType":       "n1-standard-2",
		"status":            "RUNNING",
		"zone":              "us-central1-a",
		"creationTimestamp": "2024-01-01T00:00:00Z",
		"networkInterfaces": []interface{}{
			map[string]interface{}{
				"network": "default",
				"accessConfigs": []interface{}{
					map[string]interface{}{
						"natIP": "35.123.45.67",
						"type":  "ONE_TO_ONE_NAT",
					},
				},
			},
		},
		"disks": []interface{}{
			map[string]interface{}{
				"boot":       true,
				"diskSizeGb": 50,
				"type":       "PERSISTENT",
			},
		},
	}

	instance2 := map[string]interface{}{
		"id":                "9876543210",
		"name":              "staging-web-01",
		"machineType":       "n1-standard-4",
		"status":            "RUNNING",
		"zone":              "us-west1-b",
		"creationTimestamp": "2024-02-01T00:00:00Z",
		"networkInterfaces": []interface{}{
			map[string]interface{}{
				"network": "default",
				"accessConfigs": []interface{}{
					map[string]interface{}{
						"natIP": "35.234.56.78",
						"type":  "ONE_TO_ONE_NAT",
					},
				},
			},
		},
		"disks": []interface{}{
			map[string]interface{}{
				"boot":       true,
				"diskSizeGb": 100,
				"type":       "PERSISTENT",
			},
		},
	}

	// Use default config which ignores id and creationTimestamp
	cfg := config.Default()
	differ := NewDiffer(cfg, false)

	diff := differ.Compare(instance1, instance2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected instances to have differences, got %v", diff.Type)
	}

	// Get all diffs
	allDiffs := GetAllDiffs(diff)

	// Check that we found the expected differences
	expectedDiffs := map[string]bool{
		"name":        false,
		"machineType": false,
		"zone":        false,
	}

	for _, d := range allDiffs {
		if _, expected := expectedDiffs[d.Path]; expected {
			expectedDiffs[d.Path] = true
		}
		// id and creationTimestamp should be ignored
		if d.Path == "id" || d.Path == "creationTimestamp" {
			t.Errorf("Field %q should be ignored by default config", d.Path)
		}
	}

	for field, found := range expectedDiffs {
		if !found {
			t.Errorf("Expected to find difference in field %q", field)
		}
	}

	// Test output formatting
	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "prod-web-01", "staging-web-01")

	output := buf.String()

	if !strings.Contains(output, "machineType") {
		t.Error("Output should show machineType difference")
	}

	if !strings.Contains(output, "n1-standard-2") {
		t.Error("Output should show old machine type")
	}

	if !strings.Contains(output, "n1-standard-4") {
		t.Error("Output should show new machine type")
	}
}

// TestEndToEnd_WithShowAll verifies showAll flag shows ignored fields
func TestEndToEnd_WithShowAll(t *testing.T) {
	obj1 := map[string]interface{}{
		"id":   "123",
		"name": "test",
	}

	obj2 := map[string]interface{}{
		"id":   "456",
		"name": "test",
	}

	cfg := config.Default()

	// Without showAll
	differNormal := NewDiffer(cfg, false)
	diffNormal := differNormal.Compare(obj1, obj2)

	if diffNormal.Type != DiffTypeEqual {
		t.Error("With ignored fields, should be equal")
	}

	// With showAll
	differShowAll := NewDiffer(cfg, true)
	diffShowAll := differShowAll.Compare(obj1, obj2)

	if diffShowAll.Type != DiffTypeModified {
		t.Error("With showAll, should show differences in ignored fields")
	}

	allDiffs := GetAllDiffs(diffShowAll)
	foundID := false
	for _, d := range allDiffs {
		if d.Path == "id" {
			foundID = true
		}
	}

	if !foundID {
		t.Error("showAll should include id field differences")
	}
}

// TestEndToEnd_IdenticalInstances verifies no false positives
func TestEndToEnd_IdenticalInstances(t *testing.T) {
	instance := map[string]interface{}{
		"name":        "test-instance",
		"machineType": "n1-standard-2",
		"status":      "RUNNING",
		"metadata": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	}

	// Create a copy
	instanceCopy := map[string]interface{}{
		"name":        "test-instance",
		"machineType": "n1-standard-2",
		"status":      "RUNNING",
		"metadata": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	}

	cfg := config.Default()
	differ := NewDiffer(cfg, false)

	diff := differ.Compare(instance, instanceCopy)

	if diff.Type != DiffTypeEqual {
		t.Error("Identical instances should have no differences")
	}

	allDiffs := GetAllDiffs(diff)
	if len(allDiffs) != 0 {
		t.Errorf("Expected 0 differences, got %d", len(allDiffs))
	}

	// Verify output
	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance", "instance-copy")

	output := buf.String()
	if !strings.Contains(output, "No differences found") {
		t.Error("Output should indicate no differences")
	}
}

// TestEndToEnd_CustomConfig tests with custom ignore configuration
func TestEndToEnd_CustomConfig(t *testing.T) {
	obj1 := map[string]interface{}{
		"name":       "resource-1",
		"customID":   "abc123",
		"customTime": "2024-01-01",
		"value":      100,
	}

	obj2 := map[string]interface{}{
		"name":       "resource-1",
		"customID":   "xyz789",
		"customTime": "2024-02-01",
		"value":      200,
	}

	// Custom config that ignores custom fields
	cfg := &config.Config{
		IgnoreFields: []string{"customID", "customTime"},
	}

	differ := NewDiffer(cfg, false)
	diff := differ.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Error("Should detect difference in value field")
	}

	allDiffs := GetAllDiffs(diff)

	// Should only show value difference, not customID or customTime
	if len(allDiffs) != 1 {
		t.Errorf("Expected 1 difference (value), got %d", len(allDiffs))
	}

	if allDiffs[0].Path != "value" {
		t.Errorf("Expected difference in 'value', got %q", allDiffs[0].Path)
	}
}

// TestEndToEnd_ComplexNestedStructures tests deep nested comparisons
func TestEndToEnd_ComplexNestedStructures(t *testing.T) {
	obj1 := map[string]interface{}{
		"config": map[string]interface{}{
			"database": map[string]interface{}{
				"host": "db.example.com",
				"port": 5432,
				"credentials": map[string]interface{}{
					"username": "admin",
					"password": "secret123",
				},
			},
			"cache": map[string]interface{}{
				"enabled": true,
				"ttl":     300,
			},
		},
	}

	obj2 := map[string]interface{}{
		"config": map[string]interface{}{
			"database": map[string]interface{}{
				"host": "db.example.com",
				"port": 5432,
				"credentials": map[string]interface{}{
					"username": "admin",
					"password": "secret456", // Changed
				},
			},
			"cache": map[string]interface{}{
				"enabled": false, // Changed
				"ttl":     300,
			},
		},
	}

	cfg := &config.Config{
		IgnoreFields: []string{},
	}

	differ := NewDiffer(cfg, false)
	diff := differ.Compare(obj1, obj2)

	allDiffs := GetAllDiffs(diff)

	// Should find 2 differences
	if len(allDiffs) != 2 {
		t.Errorf("Expected 2 differences, got %d", len(allDiffs))
	}

	// Check that deep paths are correct
	foundPassword := false
	foundEnabled := false

	for _, d := range allDiffs {
		if strings.Contains(d.Path, "password") {
			foundPassword = true
		}
		if strings.Contains(d.Path, "enabled") {
			foundEnabled = true
		}
	}

	if !foundPassword {
		t.Error("Should find difference in nested password field")
	}

	if !foundEnabled {
		t.Error("Should find difference in nested enabled field")
	}
}
