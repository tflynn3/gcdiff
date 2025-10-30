package cmd

import (
	"github.com/spf13/cobra"
)

// computeCmd is an alias for "resource compute" for backward compatibility
var computeCmd = &cobra.Command{
	Use:   "compute [instance-name-1] [instance-name-2]",
	Short: "Compare two GCP compute instances (alias for 'resource compute')",
	Long: `Compare two GCP compute instances, either within the same project or across projects.

This is a convenience alias for 'gcdiff resource compute'.

Examples:
  # Compare two instances in the same project
  gcdiff compute instance-1 instance-2 --project1=my-project --zone1=us-central1-a

  # Compare instances across projects
  gcdiff compute prod-instance staging-instance --project1=prod-project --project2=staging-project --zone1=us-central1-a

  # Show all fields (including normally ignored ones)
  gcdiff compute instance-1 instance-2 --project1=my-project --zone1=us-central1-a --show-all`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to resource command with "compute instances" as the resource type
		newArgs := append([]string{"compute instances"}, args...)
		return runResource(cmd, newArgs)
	},
}

func init() {
	rootCmd.AddCommand(computeCmd)
	computeCmd.Flags().String("zone1", "", "Zone for first instance (required)")
	computeCmd.Flags().String("zone2", "", "Zone for second instance (defaults to zone1)")
	_ = computeCmd.MarkFlagRequired("zone1")
}

// Removed runCompute - now using runResource via delegation
