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
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

// PrintGitStyleDiff prints a git-style diff to the writer
func PrintGitStyleDiff(w io.Writer, diff *Diff, name1, name2 string) {
	fmt.Fprintf(w, "%s\n", bold(fmt.Sprintf("Comparing: %s <-> %s", name1, name2)))
	fmt.Fprintln(w, strings.Repeat("-", 80))

	if diff.Type == DiffTypeEqual {
		fmt.Fprintf(w, "%s\n", green("✓ No differences found"))
		return
	}

	diffs := GetAllDiffs(diff)
	if len(diffs) == 0 {
		fmt.Fprintf(w, "%s\n", green("✓ No differences found"))
		return
	}

	// Group diffs by type
	added := []*Diff{}
	removed := []*Diff{}
	modified := []*Diff{}

	for _, d := range diffs {
		switch d.Type {
		case DiffTypeAdded:
			added = append(added, d)
		case DiffTypeRemoved:
			removed = append(removed, d)
		case DiffTypeModified:
			modified = append(modified, d)
		}
	}

	// Print summary
	total := len(added) + len(removed) + len(modified)
	fmt.Fprintf(w, "\n%s\n", bold(fmt.Sprintf("Summary: %d difference(s) found", total)))
	if len(added) > 0 {
		fmt.Fprintf(w, "  %s %d field(s)\n", green("+"), len(added))
	}
	if len(removed) > 0 {
		fmt.Fprintf(w, "  %s %d field(s)\n", red("-"), len(removed))
	}
	if len(modified) > 0 {
		fmt.Fprintf(w, "  %s %d field(s)\n", yellow("~"), len(modified))
	}
	fmt.Fprintln(w)

	// Print differences
	printDiffSection(w, "Added Fields", added, DiffTypeAdded)
	printDiffSection(w, "Removed Fields", removed, DiffTypeRemoved)
	printDiffSection(w, "Modified Fields", modified, DiffTypeModified)
}

func printDiffSection(w io.Writer, title string, diffs []*Diff, diffType DiffType) {
	if len(diffs) == 0 {
		return
	}

	// Sort by path for consistent output
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Path < diffs[j].Path
	})

	fmt.Fprintf(w, "%s\n", bold(title+":"))
	fmt.Fprintln(w)

	for _, d := range diffs {
		switch diffType {
		case DiffTypeAdded:
			fmt.Fprintf(w, "  %s %s\n", green("+"), cyan(d.Path))
			printValue(w, "      ", d.Value2, green)
		case DiffTypeRemoved:
			fmt.Fprintf(w, "  %s %s\n", red("-"), cyan(d.Path))
			printValue(w, "      ", d.Value1, red)
		case DiffTypeModified:
			fmt.Fprintf(w, "  %s %s\n", yellow("~"), cyan(d.Path))
			fmt.Fprintf(w, "      %s ", red("-"))
			printValue(w, "        ", d.Value1, red)
			fmt.Fprintf(w, "      %s ", green("+"))
			printValue(w, "        ", d.Value2, green)
		}
		fmt.Fprintln(w)
	}
}

func printValue(w io.Writer, indent string, value interface{}, colorFunc func(...interface{}) string) {
	if value == nil {
		fmt.Fprintf(w, "%s\n", colorFunc("<nil>"))
		return
	}

	// Try to format as JSON for complex types
	switch v := value.(type) {
	case map[string]interface{}, []interface{}:
		jsonBytes, err := json.MarshalIndent(v, indent, "  ")
		if err != nil {
			fmt.Fprintf(w, "%v\n", colorFunc(fmt.Sprintf("%v", value)))
		} else {
			lines := strings.Split(string(jsonBytes), "\n")
			for i, line := range lines {
				if i == 0 {
					fmt.Fprintf(w, "%s\n", colorFunc(line))
				} else {
					fmt.Fprintf(w, "%s%s\n", indent, colorFunc(line))
				}
			}
		}
	case string:
		fmt.Fprintf(w, "%s\n", colorFunc(fmt.Sprintf("%q", v)))
	default:
		fmt.Fprintf(w, "%v\n", colorFunc(fmt.Sprintf("%v", value)))
	}
}
