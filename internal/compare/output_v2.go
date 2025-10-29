package compare

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"
)

var (
	gray = color.New(color.FgHiBlack).SprintFunc()
)

// PrintGitStyleDiffV2 prints a diff with arrays shown inline with markers
func PrintGitStyleDiffV2(w io.Writer, diff *Diff, name1, name2 string) {
	fmt.Fprintf(w, "%s\n", bold(fmt.Sprintf("Comparing: %s <-> %s", name1, name2)))
	fmt.Fprintln(w, strings.Repeat("-", 80))

	if diff.Type == DiffTypeEqual {
		fmt.Fprintf(w, "%s\n", green("✓ No differences found"))
		return
	}

	// Group top-level differences
	topLevelDiffs := getTopLevelDiffs(diff)

	if len(topLevelDiffs) == 0 {
		fmt.Fprintf(w, "%s\n", green("✓ No differences found"))
		return
	}

	// Print each top-level field with its differences
	fmt.Fprintln(w)
	for _, fieldName := range getSortedKeys(topLevelDiffs) {
		fieldDiff := topLevelDiffs[fieldName]
		printFieldDiff(w, fieldName, fieldDiff, 0)
		fmt.Fprintln(w)
	}
}

// getTopLevelDiffs groups diffs by their top-level field name
func getTopLevelDiffs(diff *Diff) map[string]*Diff {
	result := make(map[string]*Diff)

	for key, child := range diff.Children {
		// Extract top-level field name (before any brackets or dots)
		topField := extractTopLevelField(key)
		if existing, ok := result[topField]; ok {
			// Merge if same top field
			if existing.Children == nil {
				existing.Children = make(map[string]*Diff)
			}
			existing.Children[key] = child
			if child.Type != DiffTypeEqual {
				existing.Type = DiffTypeModified
			}
		} else {
			result[topField] = child
		}
	}

	return result
}

func extractTopLevelField(path string) string {
	// Extract field name before any [index] or .subfield
	if idx := strings.IndexAny(path, ".["); idx != -1 {
		return path[:idx]
	}
	return path
}

func getSortedKeys(m map[string]*Diff) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func printFieldDiff(w io.Writer, fieldName string, fieldDiff *Diff, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Check if this is an array diff
	if isArrayDiff(fieldDiff) {
		printArrayDiff(w, fieldName, fieldDiff, indent)
		return
	}

	// Check if this is an object diff
	if len(fieldDiff.Children) > 0 && fieldDiff.Type == DiffTypeModified {
		fmt.Fprintf(w, "%s%s %s\n", indentStr, yellow("~"), cyan(fieldName))
		for _, childKey := range getSortedKeys(fieldDiff.Children) {
			childDiff := fieldDiff.Children[childKey]
			printFieldDiff(w, childKey, childDiff, indent+1)
		}
		return
	}

	// Simple field diff
	switch fieldDiff.Type {
	case DiffTypeAdded:
		fmt.Fprintf(w, "%s%s %s\n", indentStr, green("+"), cyan(fieldName))
		printValue(w, indentStr+"    ", fieldDiff.Value2, green)
	case DiffTypeRemoved:
		fmt.Fprintf(w, "%s%s %s\n", indentStr, red("-"), cyan(fieldName))
		printValue(w, indentStr+"    ", fieldDiff.Value1, red)
	case DiffTypeModified:
		fmt.Fprintf(w, "%s%s %s\n", indentStr, yellow("~"), cyan(fieldName))
		fmt.Fprintf(w, "%s    %s ", indentStr, red("-"))
		printValue(w, indentStr+"      ", fieldDiff.Value1, red)
		fmt.Fprintf(w, "%s    %s ", indentStr, green("+"))
		printValue(w, indentStr+"      ", fieldDiff.Value2, green)
	}
}

func isArrayDiff(diff *Diff) bool {
	if len(diff.Children) == 0 {
		return false
	}

	// Check if all children are array indices
	for key := range diff.Children {
		if !strings.HasPrefix(key, "[") {
			return false
		}
	}
	return true
}

func printArrayDiff(w io.Writer, fieldName string, arrayDiff *Diff, indent int) {
	indentStr := strings.Repeat("  ", indent)

	fmt.Fprintf(w, "%s%s %s (array with changes)\n", indentStr, yellow("~"), cyan(fieldName))

	// Get all array indices
	indices := make([]int, 0)
	childMap := make(map[int]*Diff)

	for key, child := range arrayDiff.Children {
		var idx int
		fmt.Sscanf(key, "[%d]", &idx)
		indices = append(indices, idx)
		childMap[idx] = child
	}
	sort.Ints(indices)

	// Print each array element with diff markers
	for _, idx := range indices {
		child := childMap[idx]
		elementIndent := indentStr + "    "

		switch child.Type {
		case DiffTypeAdded:
			fmt.Fprintf(w, "%s%s [%d] ", elementIndent, green("+"), idx)
			printInlineValue(w, child.Value2, green)
		case DiffTypeRemoved:
			fmt.Fprintf(w, "%s%s [%d] ", elementIndent, red("-"), idx)
			printInlineValue(w, child.Value1, red)
		case DiffTypeModified:
			// Show the element with nested changes
			if len(child.Children) > 0 {
				fmt.Fprintf(w, "%s%s [%d] (modified)\n", elementIndent, yellow("~"), idx)
				for _, childKey := range getSortedKeys(child.Children) {
					childDiff := child.Children[childKey]
					printNestedChange(w, elementIndent+"  ", childKey, childDiff)
				}
			} else {
				// Simple value change
				fmt.Fprintf(w, "%s%s [%d]\n", elementIndent, yellow("~"), idx)
				fmt.Fprintf(w, "%s    %s ", elementIndent, red("-"))
				printInlineValue(w, child.Value1, red)
				fmt.Fprintf(w, "%s    %s ", elementIndent, green("+"))
				printInlineValue(w, child.Value2, green)
			}
		}
	}
}

func printNestedChange(w io.Writer, indent string, key string, diff *Diff) {
	switch diff.Type {
	case DiffTypeAdded:
		fmt.Fprintf(w, "%s  %s %s: ", indent, green("+"), key)
		printInlineValue(w, diff.Value2, green)
	case DiffTypeRemoved:
		fmt.Fprintf(w, "%s  %s %s: ", indent, red("-"), key)
		printInlineValue(w, diff.Value1, red)
	case DiffTypeModified:
		fmt.Fprintf(w, "%s  %s %s\n", indent, yellow("~"), key)
		if len(diff.Children) > 0 {
			// Nested object changes
			for _, childKey := range getSortedKeys(diff.Children) {
				printNestedChange(w, indent+"  ", childKey, diff.Children[childKey])
			}
		} else {
			fmt.Fprintf(w, "%s      %s ", indent, red("-"))
			printInlineValue(w, diff.Value1, red)
			fmt.Fprintf(w, "%s      %s ", indent, green("+"))
			printInlineValue(w, diff.Value2, green)
		}
	}
}

func printInlineValue(w io.Writer, value interface{}, colorFunc func(...interface{}) string) {
	if value == nil {
		fmt.Fprintf(w, "%s\n", colorFunc("<nil>"))
		return
	}

	switch v := value.(type) {
	case map[string]interface{}:
		jsonBytes, _ := json.Marshal(v)
		fmt.Fprintf(w, "%s\n", colorFunc(string(jsonBytes)))
	case []interface{}:
		jsonBytes, _ := json.Marshal(v)
		fmt.Fprintf(w, "%s\n", colorFunc(string(jsonBytes)))
	case string:
		fmt.Fprintf(w, "%s\n", colorFunc(fmt.Sprintf("%q", v)))
	default:
		fmt.Fprintf(w, "%s\n", colorFunc(fmt.Sprintf("%v", v)))
	}
}
