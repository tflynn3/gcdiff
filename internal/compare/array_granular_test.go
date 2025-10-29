package compare

import (
	"testing"

	"github.com/tflynn3/gcdiff/internal/config"
)

// TestCompare_ArrayGranularDiff tests that array differences are shown at element level
func TestCompare_ArrayGranularDiff(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	// Simulate firewall rules with array of allowed ports
	obj1 := map[string]interface{}{
		"allowed": []interface{}{
			map[string]interface{}{
				"IPProtocol": "tcp",
				"ports":      []interface{}{"80"},
			},
			map[string]interface{}{
				"IPProtocol": "tcp",
				"ports":      []interface{}{"443"},
			},
		},
	}

	obj2 := map[string]interface{}{
		"allowed": []interface{}{
			map[string]interface{}{
				"IPProtocol": "tcp",
				"ports":      []interface{}{"80"},
			},
			map[string]interface{}{
				"IPProtocol": "tcp",
				"ports":      []interface{}{"443"},
			},
			map[string]interface{}{
				"IPProtocol": "tcp",
				"ports":      []interface{}{"8080"},
			},
		},
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	allowedDiff := diff.Children["allowed"]
	if allowedDiff == nil {
		t.Fatal("Expected 'allowed' field in children")
	}

	// Should have element [2] as added
	elem2Diff := allowedDiff.Children["[2]"]
	if elem2Diff == nil {
		t.Fatal("Expected '[2]' element in allowed children")
	}

	if elem2Diff.Type != DiffTypeAdded {
		t.Errorf("Expected element [2] to be DiffTypeAdded, got %v", elem2Diff.Type)
	}

	// Elements [0] and [1] should not be in children (they're equal)
	if _, exists := allowedDiff.Children["[0]"]; exists {
		t.Error("Element [0] should not be in children since it's equal")
	}

	if _, exists := allowedDiff.Children["[1]"]; exists {
		t.Error("Element [1] should not be in children since it's equal")
	}
}

// TestCompare_ArrayElementRemoved tests element removal detection
func TestCompare_ArrayElementRemoved(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	obj2 := map[string]interface{}{
		"items": []interface{}{"a", "b"},
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	itemsDiff := diff.Children["items"]
	if itemsDiff == nil {
		t.Fatal("Expected 'items' field in children")
	}

	// Element [2] should be marked as removed
	elem2Diff := itemsDiff.Children["[2]"]
	if elem2Diff == nil {
		t.Fatal("Expected '[2]' element in items children")
	}

	if elem2Diff.Type != DiffTypeRemoved {
		t.Errorf("Expected element [2] to be DiffTypeRemoved, got %v", elem2Diff.Type)
	}

	if elem2Diff.Value1 != "c" {
		t.Errorf("Expected Value1 to be 'c', got %v", elem2Diff.Value1)
	}
}

// TestCompare_ArrayComplexNestedChange tests nested object changes within arrays
func TestCompare_ArrayComplexNestedChange(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"protocol": "tcp",
				"port":     80,
			},
			map[string]interface{}{
				"protocol": "tcp",
				"port":     443,
			},
		},
	}

	obj2 := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"protocol": "tcp",
				"port":     80,
			},
			map[string]interface{}{
				"protocol": "tcp",
				"port":     8443, // Changed port
			},
		},
	}

	diff := d.Compare(obj1, obj2)

	rulesDiff := diff.Children["rules"]
	if rulesDiff == nil {
		t.Fatal("Expected 'rules' field in children")
	}

	// Element [1] should show nested change in port field
	elem1Diff := rulesDiff.Children["[1]"]
	if elem1Diff == nil {
		t.Fatal("Expected '[1]' element in rules children")
	}

	if elem1Diff.Type != DiffTypeModified {
		t.Errorf("Expected element [1] to be DiffTypeModified, got %v", elem1Diff.Type)
	}

	// Should have nested port change
	portDiff := elem1Diff.Children["port"]
	if portDiff == nil {
		t.Fatal("Expected 'port' field in element [1] children")
	}

	if portDiff.Type != DiffTypeModified {
		t.Errorf("Expected port to be DiffTypeModified, got %v", portDiff.Type)
	}

	// Verify the actual values (JSON numbers are int, not float64 in this case)
	if portDiff.Value1 != 443 || portDiff.Value2 != 8443 {
		t.Errorf("Expected port change from 443 to 8443, got %v to %v", portDiff.Value1, portDiff.Value2)
	}
}
