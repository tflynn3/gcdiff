package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tflynn3/gcdiff/internal/compare"
	"github.com/tflynn3/gcdiff/internal/config"
	"github.com/tflynn3/gcdiff/internal/gcp"
)

var computeCmd = &cobra.Command{
	Use:   "compute [instance-name-1] [instance-name-2]",
	Short: "Compare two GCP compute instances",
	Long: `Compare two GCP compute instances, either within the same project or across projects.

Examples:
  # Compare two instances in the same project
  gcdiff compute instance-1 instance-2 --project1=my-project

  # Compare instances across projects
  gcdiff compute prod-instance staging-instance --project1=prod-project --project2=staging-project

  # Show all fields (including normally ignored ones)
  gcdiff compute instance-1 instance-2 --project1=my-project --show-all`,
	Args: cobra.ExactArgs(2),
	RunE: runCompute,
}

func init() {
	rootCmd.AddCommand(computeCmd)
	computeCmd.Flags().String("zone1", "", "Zone for first instance (required)")
	computeCmd.Flags().String("zone2", "", "Zone for second instance (defaults to zone1)")
	computeCmd.MarkFlagRequired("zone1")
}

func runCompute(cmd *cobra.Command, args []string) error {
	instance1 := args[0]
	instance2 := args[1]

	project1 := viper.GetString("project1")
	project2 := viper.GetString("project2")
	if project2 == "" {
		project2 = project1
	}

	zone1, _ := cmd.Flags().GetString("zone1")
	zone2, _ := cmd.Flags().GetString("zone2")
	if zone2 == "" {
		zone2 = zone1
	}

	if project1 == "" {
		return fmt.Errorf("--project1 is required")
	}

	ctx := context.Background()

	// Initialize GCP client
	client, err := gcp.NewComputeClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}
	defer client.Close()

	// Fetch instances
	fmt.Fprintf(cmd.OutOrStderr(), "Fetching instance %s from %s/%s...\n", instance1, project1, zone1)
	inst1, err := client.GetInstance(ctx, project1, zone1, instance1)
	if err != nil {
		return fmt.Errorf("failed to fetch instance %s: %w", instance1, err)
	}

	fmt.Fprintf(cmd.OutOrStderr(), "Fetching instance %s from %s/%s...\n", instance2, project2, zone2)
	inst2, err := client.GetInstance(ctx, project2, zone2, instance2)
	if err != nil {
		return fmt.Errorf("failed to fetch instance %s: %w", instance2, err)
	}

	// Load config for field filtering
	cfg, err := config.Load(viper.ConfigFileUsed())
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	// Convert to JSON for comparison
	data1, err := json.Marshal(inst1)
	if err != nil {
		return fmt.Errorf("failed to marshal instance1: %w", err)
	}

	data2, err := json.Marshal(inst2)
	if err != nil {
		return fmt.Errorf("failed to marshal instance2: %w", err)
	}

	var obj1, obj2 map[string]interface{}
	json.Unmarshal(data1, &obj1)
	json.Unmarshal(data2, &obj2)

	// Compare and output
	differ := compare.NewDiffer(cfg, viper.GetBool("show-all"))
	diff := differ.Compare(obj1, obj2)

	format := viper.GetString("format")
	switch format {
	case "json":
		output, _ := json.MarshalIndent(diff, "", "  ")
		fmt.Println(string(output))
	case "diff":
		fallthrough
	default:
		compare.PrintGitStyleDiff(cmd.OutOrStdout(), diff, instance1, instance2)
	}

	return nil
}
