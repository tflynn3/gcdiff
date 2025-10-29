package compare

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintGitStyleDiff_NoDifferences(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeEqual,
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "instance-1") {
		t.Error("Output should contain first instance name")
	}

	if !strings.Contains(output, "instance-2") {
		t.Error("Output should contain second instance name")
	}

	if !strings.Contains(output, "No differences found") {
		t.Error("Output should indicate no differences")
	}
}

func TestPrintGitStyleDiff_WithDifferences(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"name": {
				Path:   "name",
				Type:   DiffTypeModified,
				Value1: "old-name",
				Value2: "new-name",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "Summary") {
		t.Error("Output should contain summary section")
	}

	if !strings.Contains(output, "difference(s) found") {
		t.Error("Output should show differences count")
	}

	if !strings.Contains(output, "Modified Fields") {
		t.Error("Output should have modified fields section")
	}

	if !strings.Contains(output, "name") {
		t.Error("Output should contain the modified field name")
	}
}

func TestPrintGitStyleDiff_AddedFields(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"newfield": {
				Path:   "newfield",
				Type:   DiffTypeAdded,
				Value2: "value",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "Added Fields") {
		t.Error("Output should have added fields section")
	}

	if !strings.Contains(output, "newfield") {
		t.Error("Output should contain the added field name")
	}

	// Check for + symbol (added)
	if !strings.Contains(output, "+") {
		t.Error("Output should contain + symbol for added fields")
	}
}

func TestPrintGitStyleDiff_RemovedFields(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"oldfield": {
				Path:   "oldfield",
				Type:   DiffTypeRemoved,
				Value1: "value",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "Removed Fields") {
		t.Error("Output should have removed fields section")
	}

	if !strings.Contains(output, "oldfield") {
		t.Error("Output should contain the removed field name")
	}

	// Check for - symbol (removed)
	if !strings.Contains(output, "-") {
		t.Error("Output should contain - symbol for removed fields")
	}
}

func TestPrintGitStyleDiff_ModifiedFields(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"status": {
				Path:   "status",
				Type:   DiffTypeModified,
				Value1: "running",
				Value2: "stopped",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "Modified Fields") {
		t.Error("Output should have modified fields section")
	}

	if !strings.Contains(output, "status") {
		t.Error("Output should contain the modified field name")
	}

	if !strings.Contains(output, "running") {
		t.Error("Output should contain old value")
	}

	if !strings.Contains(output, "stopped") {
		t.Error("Output should contain new value")
	}

	// Both + and - should appear for modified fields
	if !strings.Contains(output, "+") || !strings.Contains(output, "-") {
		t.Error("Output should contain both + and - symbols for modified fields")
	}
}

func TestPrintGitStyleDiff_MixedDifferences(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"added": {
				Path:   "added",
				Type:   DiffTypeAdded,
				Value2: "new",
			},
			"removed": {
				Path:   "removed",
				Type:   DiffTypeRemoved,
				Value1: "old",
			},
			"modified": {
				Path:   "modified",
				Type:   DiffTypeModified,
				Value1: "before",
				Value2: "after",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	// Check all sections are present
	if !strings.Contains(output, "Added Fields") {
		t.Error("Should have Added Fields section")
	}

	if !strings.Contains(output, "Removed Fields") {
		t.Error("Should have Removed Fields section")
	}

	if !strings.Contains(output, "Modified Fields") {
		t.Error("Should have Modified Fields section")
	}

	// Check summary counts
	if !strings.Contains(output, "3 difference(s) found") {
		t.Error("Summary should show 3 differences")
	}
}

func TestPrintGitStyleDiff_ComplexValues(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"config": {
				Path: "config",
				Type: DiffTypeModified,
				Value1: map[string]interface{}{
					"enabled": true,
					"timeout": 30,
				},
				Value2: map[string]interface{}{
					"enabled": false,
					"timeout": 60,
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	// Complex values should be formatted as JSON
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Error("Complex values should be formatted as JSON")
	}

	if !strings.Contains(output, "config") {
		t.Error("Output should contain field name")
	}
}

func TestPrintGitStyleDiff_NilValue(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"field": {
				Path:   "field",
				Type:   DiffTypeModified,
				Value1: nil,
				Value2: "value",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	if !strings.Contains(output, "field") {
		t.Error("Output should contain field name")
	}

	// Should handle nil gracefully
	if !strings.Contains(output, "nil") && !strings.Contains(output, "<nil>") {
		t.Error("Output should represent nil value")
	}
}

func TestPrintGitStyleDiff_StringValue(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"description": {
				Path:   "description",
				Type:   DiffTypeModified,
				Value1: "old description",
				Value2: "new description",
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	// Strings should be quoted
	if !strings.Contains(output, `"old description"`) {
		t.Error("String values should be quoted in output")
	}

	if !strings.Contains(output, `"new description"`) {
		t.Error("String values should be quoted in output")
	}
}

func TestPrintGitStyleDiff_ArrayValue(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"tags": {
				Path:   "tags",
				Type:   DiffTypeModified,
				Value1: []interface{}{"tag1", "tag2"},
				Value2: []interface{}{"tag1", "tag3"},
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	// Arrays should be formatted as JSON
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Array values should be formatted as JSON")
	}
}

func TestGetAllDiffs_Ordering(t *testing.T) {
	diff := &Diff{
		Path: "",
		Type: DiffTypeModified,
		Children: map[string]*Diff{
			"z_field": {
				Path: "z_field",
				Type: DiffTypeModified,
			},
			"a_field": {
				Path: "a_field",
				Type: DiffTypeModified,
			},
			"m_field": {
				Path: "m_field",
				Type: DiffTypeModified,
			},
		},
	}

	var buf bytes.Buffer
	PrintGitStyleDiff(&buf, diff, "instance-1", "instance-2")

	output := buf.String()

	// Find positions of each field in output
	posA := strings.Index(output, "a_field")
	posM := strings.Index(output, "m_field")
	posZ := strings.Index(output, "z_field")

	// All should be found
	if posA == -1 || posM == -1 || posZ == -1 {
		t.Error("All fields should appear in output")
	}

	// Should be in alphabetical order
	if !(posA < posM && posM < posZ) {
		t.Error("Fields should be sorted alphabetically in output")
	}
}
