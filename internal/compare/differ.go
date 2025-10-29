package compare

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/tflynn3/gcdiff/internal/config"
)

// DiffType represents the type of difference
type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
	DiffTypeEqual    DiffType = "equal"
)

// Diff represents a difference between two values
type Diff struct {
	Path     string                 `json:"path"`
	Type     DiffType               `json:"type"`
	Value1   interface{}            `json:"value1,omitempty"`
	Value2   interface{}            `json:"value2,omitempty"`
	Children map[string]*Diff       `json:"children,omitempty"`
}

// Differ performs deep comparison of objects
type Differ struct {
	config  *config.Config
	showAll bool
}

// NewDiffer creates a new Differ
func NewDiffer(cfg *config.Config, showAll bool) *Differ {
	if cfg == nil {
		cfg = config.Default()
	}
	return &Differ{
		config:  cfg,
		showAll: showAll,
	}
}

// Compare compares two objects and returns differences
func (d *Differ) Compare(obj1, obj2 map[string]interface{}) *Diff {
	return d.compareObjects(obj1, obj2, "")
}

func (d *Differ) compareObjects(obj1, obj2 map[string]interface{}, path string) *Diff {
	diff := &Diff{
		Path:     path,
		Type:     DiffTypeEqual,
		Children: make(map[string]*Diff),
	}

	// Get all keys from both objects
	keys := make(map[string]bool)
	for k := range obj1 {
		keys[k] = true
	}
	for k := range obj2 {
		keys[k] = true
	}

	// Compare each key
	for key := range keys {
		fieldPath := key
		if path != "" {
			fieldPath = path + "." + key
		}

		// Skip ignored fields unless showAll is true
		if !d.showAll && d.config.ShouldIgnore(fieldPath) {
			continue
		}

		val1, exists1 := obj1[key]
		val2, exists2 := obj2[key]

		if !exists1 && exists2 {
			diff.Children[key] = &Diff{
				Path:   fieldPath,
				Type:   DiffTypeAdded,
				Value2: val2,
			}
			diff.Type = DiffTypeModified
		} else if exists1 && !exists2 {
			diff.Children[key] = &Diff{
				Path:   fieldPath,
				Type:   DiffTypeRemoved,
				Value1: val1,
			}
			diff.Type = DiffTypeModified
		} else {
			childDiff := d.compareValues(val1, val2, fieldPath)
			if childDiff.Type != DiffTypeEqual {
				diff.Children[key] = childDiff
				diff.Type = DiffTypeModified
			}
		}
	}

	return diff
}

func (d *Differ) compareValues(val1, val2 interface{}, path string) *Diff {
	// Handle nil values
	if val1 == nil && val2 == nil {
		return &Diff{Path: path, Type: DiffTypeEqual}
	}
	if val1 == nil {
		return &Diff{Path: path, Type: DiffTypeAdded, Value2: val2}
	}
	if val2 == nil {
		return &Diff{Path: path, Type: DiffTypeRemoved, Value1: val1}
	}

	// Check if types match
	type1 := reflect.TypeOf(val1)
	type2 := reflect.TypeOf(val2)
	if type1 != type2 {
		return &Diff{
			Path:   path,
			Type:   DiffTypeModified,
			Value1: val1,
			Value2: val2,
		}
	}

	// Handle different types
	switch v1 := val1.(type) {
	case map[string]interface{}:
		v2 := val2.(map[string]interface{})
		return d.compareObjects(v1, v2, path)
	case []interface{}:
		v2 := val2.([]interface{})
		return d.compareArrays(v1, v2, path)
	default:
		if reflect.DeepEqual(val1, val2) {
			return &Diff{Path: path, Type: DiffTypeEqual}
		}
		return &Diff{
			Path:   path,
			Type:   DiffTypeModified,
			Value1: val1,
			Value2: val2,
		}
	}
}

func (d *Differ) compareArrays(arr1, arr2 []interface{}, path string) *Diff {
	diff := &Diff{
		Path:     path,
		Type:     DiffTypeEqual,
		Children: make(map[string]*Diff),
	}

	maxLen := len(arr1)
	if len(arr2) > maxLen {
		maxLen = len(arr2)
	}

	for i := 0; i < maxLen; i++ {
		indexPath := fmt.Sprintf("%s[%d]", path, i)
		key := fmt.Sprintf("[%d]", i)

		// Element exists in both arrays - compare them
		if i < len(arr1) && i < len(arr2) {
			childDiff := d.compareValues(arr1[i], arr2[i], indexPath)
			if childDiff.Type != DiffTypeEqual {
				diff.Children[key] = childDiff
				diff.Type = DiffTypeModified
			}
		} else if i >= len(arr1) {
			// Element only exists in arr2 - it was added
			diff.Children[key] = &Diff{
				Path:   indexPath,
				Type:   DiffTypeAdded,
				Value2: arr2[i],
			}
			diff.Type = DiffTypeModified
		} else {
			// Element only exists in arr1 - it was removed
			diff.Children[key] = &Diff{
				Path:   indexPath,
				Type:   DiffTypeRemoved,
				Value1: arr1[i],
			}
			diff.Type = DiffTypeModified
		}
	}

	return diff
}

// GetAllDiffs returns a flat list of all differences
func GetAllDiffs(diff *Diff) []*Diff {
	var diffs []*Diff

	if diff.Type != DiffTypeEqual && len(diff.Children) == 0 {
		diffs = append(diffs, diff)
	}

	// Get sorted keys for consistent ordering
	keys := make([]string, 0, len(diff.Children))
	for k := range diff.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		child := diff.Children[key]
		diffs = append(diffs, GetAllDiffs(child)...)
	}

	return diffs
}
