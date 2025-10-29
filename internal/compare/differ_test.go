package compare

import (
	"testing"

	"github.com/tflynn3/gcdiff/internal/config"
)

func TestNewDiffer(t *testing.T) {
	cfg := &config.Config{
		IgnoreFields: []string{"id"},
	}

	d := NewDiffer(cfg, false)
	if d == nil {
		t.Fatal("NewDiffer returned nil")
	}

	if d.config != cfg {
		t.Error("Config not set correctly")
	}

	if d.showAll != false {
		t.Error("showAll not set correctly")
	}
}

func TestNewDiffer_NilConfig(t *testing.T) {
	d := NewDiffer(nil, false)
	if d == nil {
		t.Fatal("NewDiffer returned nil")
	}

	if d.config == nil {
		t.Error("Should use default config when nil is passed")
	}
}

func TestCompare_Equal(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"name":   "test",
		"value":  123,
		"active": true,
	}

	obj2 := map[string]interface{}{
		"name":   "test",
		"value":  123,
		"active": true,
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeEqual {
		t.Errorf("Expected DiffTypeEqual, got %v", diff.Type)
	}
}

func TestCompare_Modified(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"name":  "test1",
		"value": 123,
	}

	obj2 := map[string]interface{}{
		"name":  "test2",
		"value": 123,
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	if _, exists := diff.Children["name"]; !exists {
		t.Error("Expected 'name' field in children")
	}

	nameDiff := diff.Children["name"]
	if nameDiff.Type != DiffTypeModified {
		t.Error("Expected 'name' field to be modified")
	}

	if nameDiff.Value1 != "test1" || nameDiff.Value2 != "test2" {
		t.Error("Values not captured correctly")
	}
}

func TestCompare_Added(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"name": "test",
	}

	obj2 := map[string]interface{}{
		"name":  "test",
		"extra": "field",
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	if extraDiff, exists := diff.Children["extra"]; !exists {
		t.Error("Expected 'extra' field in children")
	} else {
		if extraDiff.Type != DiffTypeAdded {
			t.Errorf("Expected DiffTypeAdded, got %v", extraDiff.Type)
		}
		if extraDiff.Value2 != "field" {
			t.Errorf("Expected value2 to be 'field', got %v", extraDiff.Value2)
		}
	}
}

func TestCompare_Removed(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"name":  "test",
		"extra": "field",
	}

	obj2 := map[string]interface{}{
		"name": "test",
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	if extraDiff, exists := diff.Children["extra"]; !exists {
		t.Error("Expected 'extra' field in children")
	} else {
		if extraDiff.Type != DiffTypeRemoved {
			t.Errorf("Expected DiffTypeRemoved, got %v", extraDiff.Type)
		}
		if extraDiff.Value1 != "field" {
			t.Errorf("Expected value1 to be 'field', got %v", extraDiff.Value1)
		}
	}
}

func TestCompare_NestedObjects(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"name": "test",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"author":  "alice",
		},
	}

	obj2 := map[string]interface{}{
		"name": "test",
		"metadata": map[string]interface{}{
			"version": "2.0",
			"author":  "alice",
		},
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	metadataDiff, exists := diff.Children["metadata"]
	if !exists {
		t.Fatal("Expected 'metadata' field in children")
	}

	versionDiff, exists := metadataDiff.Children["version"]
	if !exists {
		t.Fatal("Expected 'version' field in metadata children")
	}

	if versionDiff.Type != DiffTypeModified {
		t.Error("Expected version to be modified")
	}

	if versionDiff.Value1 != "1.0" || versionDiff.Value2 != "2.0" {
		t.Error("Version values not captured correctly")
	}
}

func TestCompare_Arrays(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	}

	obj2 := map[string]interface{}{
		"tags": []interface{}{"tag1", "tag2", "tag4"},
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	tagsDiff := diff.Children["tags"]
	if tagsDiff == nil {
		t.Fatal("Expected 'tags' field in children")
	}

	// Check that array difference is detected
	if tagsDiff.Type != DiffTypeModified {
		t.Error("Expected tags array to be modified")
	}
}

func TestCompare_ArrayDifferentLength(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"items": []interface{}{"a", "b"},
	}

	obj2 := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified, got %v", diff.Type)
	}

	itemsDiff := diff.Children["items"]
	if itemsDiff.Type != DiffTypeModified {
		t.Error("Expected items to be modified")
	}
}

func TestCompare_WithIgnoreFields(t *testing.T) {
	cfg := &config.Config{
		IgnoreFields: []string{"id", "timestamp"},
	}
	d := NewDiffer(cfg, false)

	obj1 := map[string]interface{}{
		"id":        "123",
		"timestamp": "2024-01-01",
		"name":      "test",
	}

	obj2 := map[string]interface{}{
		"id":        "456",
		"timestamp": "2024-01-02",
		"name":      "test",
	}

	diff := d.Compare(obj1, obj2)

	// Should be equal because id and timestamp are ignored
	if diff.Type != DiffTypeEqual {
		t.Errorf("Expected DiffTypeEqual with ignored fields, got %v", diff.Type)
	}
}

func TestCompare_ShowAll(t *testing.T) {
	cfg := &config.Config{
		IgnoreFields: []string{"id"},
	}
	d := NewDiffer(cfg, true) // showAll = true

	obj1 := map[string]interface{}{
		"id":   "123",
		"name": "test",
	}

	obj2 := map[string]interface{}{
		"id":   "456",
		"name": "test",
	}

	diff := d.Compare(obj1, obj2)

	// Should NOT be equal because showAll ignores the ignore list
	if diff.Type != DiffTypeModified {
		t.Errorf("Expected DiffTypeModified with showAll=true, got %v", diff.Type)
	}

	if _, exists := diff.Children["id"]; !exists {
		t.Error("Expected 'id' field to be compared when showAll=true")
	}
}

func TestCompare_NilValues(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"field": nil,
	}

	obj2 := map[string]interface{}{
		"field": nil,
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeEqual {
		t.Error("Nil values should be equal")
	}
}

func TestCompare_NilToValue(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"field": nil,
	}

	obj2 := map[string]interface{}{
		"field": "value",
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Error("Expected modification from nil to value")
	}

	fieldDiff := diff.Children["field"]
	if fieldDiff.Type != DiffTypeAdded {
		t.Errorf("Expected DiffTypeAdded for nil->value, got %v", fieldDiff.Type)
	}
}

func TestCompare_DifferentTypes(t *testing.T) {
	d := NewDiffer(config.Default(), false)

	obj1 := map[string]interface{}{
		"field": "123",
	}

	obj2 := map[string]interface{}{
		"field": 123,
	}

	diff := d.Compare(obj1, obj2)

	if diff.Type != DiffTypeModified {
		t.Error("Different types should be detected as modified")
	}

	fieldDiff := diff.Children["field"]
	if fieldDiff.Type != DiffTypeModified {
		t.Error("Expected field to be modified due to type difference")
	}
}

func TestGetAllDiffs(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"field1": {
				Path:   "field1",
				Type:   DiffTypeModified,
				Value1: "old",
				Value2: "new",
			},
			"field2": {
				Path:   "field2",
				Type:   DiffTypeAdded,
				Value2: "added",
			},
		},
	}

	diffs := GetAllDiffs(diff)

	if len(diffs) != 2 {
		t.Errorf("Expected 2 diffs, got %d", len(diffs))
	}

	// Check that both diffs are included
	foundField1 := false
	foundField2 := false
	for _, d := range diffs {
		if d.Path == "field1" {
			foundField1 = true
		}
		if d.Path == "field2" {
			foundField2 = true
		}
	}

	if !foundField1 || !foundField2 {
		t.Error("Not all diffs were included in GetAllDiffs")
	}
}

func TestGetAllDiffs_Empty(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeEqual,
	}

	diffs := GetAllDiffs(diff)

	if len(diffs) != 0 {
		t.Errorf("Expected 0 diffs for equal diff, got %d", len(diffs))
	}
}

func TestGetAllDiffs_Nested(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"metadata": {
				Path: "metadata",
				Type: DiffTypeModified,
				Children: map[string]*Diff{
					"version": {
						Path:   "metadata.version",
						Type:   DiffTypeModified,
						Value1: "1.0",
						Value2: "2.0",
					},
				},
			},
		},
	}

	diffs := GetAllDiffs(diff)

	if len(diffs) != 1 {
		t.Errorf("Expected 1 diff, got %d", len(diffs))
	}

	if diffs[0].Path != "metadata.version" {
		t.Errorf("Expected path 'metadata.version', got %q", diffs[0].Path)
	}
}
