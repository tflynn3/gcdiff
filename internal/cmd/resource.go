package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tflynn3/gcdiff/internal/compare"
	"github.com/tflynn3/gcdiff/internal/config"
	"github.com/tflynn3/gcdiff/internal/gcp"
)

var resourceCmd = &cobra.Command{
	Use:   "resource [resource-type] [name-1] [name-2]",
	Short: "Compare any two GCP resources (universal command)",
	Long: `Compare any two GCP resources using gcloud CLI.
This command dynamically supports ALL GCP resources without requiring explicit code.

Supported resource types (built-in shortcuts):
  compute       - Compute Engine instances
  storage       - Cloud Storage buckets
  firewall      - Compute Engine firewall rules
  network       - VPC networks
  subnet        - VPC subnets
  disk          - Compute Engine disks
  run           - Cloud Run services
  sql           - Cloud SQL instances
  pubsub-topic  - Pub/Sub topics
  pubsub-subscription - Pub/Sub subscriptions
  iam-service-account - IAM service accounts

You can also use ANY gcloud resource by providing the full gcloud path!

Examples:
  # Compute instances
  gcdiff resource compute instance-1 instance-2 --project1=proj --zone1=us-central1-a

  # Storage buckets
  gcdiff resource storage bucket-1 bucket-2 --project1=proj

  # Firewall rules
  gcdiff resource firewall rule-1 rule-2 --project1=proj

  # Cloud Run services
  gcdiff resource run service-1 service-2 --project1=proj --region1=us-central1

  # Any resource via custom gcloud path (advanced)
  gcdiff resource "container clusters" cluster-1 cluster-2 --project1=proj --zone1=us-central1-a`,
	Args: cobra.ExactArgs(3),
	RunE: runResource,
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	// Generic location flags that work for most resources
	resourceCmd.Flags().String("zone1", "", "Zone for first resource (for zonal resources)")
	resourceCmd.Flags().String("zone2", "", "Zone for second resource (defaults to zone1)")
	resourceCmd.Flags().String("region1", "", "Region for first resource (for regional resources)")
	resourceCmd.Flags().String("region2", "", "Region for second resource (defaults to region1)")
	resourceCmd.Flags().String("location1", "", "Location for first resource (alternative to zone/region)")
	resourceCmd.Flags().String("location2", "", "Location for second resource (defaults to location1)")
}

func runResource(cmd *cobra.Command, args []string) error {
	resourceTypeStr := args[0]
	name1 := args[1]
	name2 := args[2]

	project1 := viper.GetString("project1")
	project2 := viper.GetString("project2")
	if project2 == "" {
		project2 = project1
	}

	if project1 == "" {
		return fmt.Errorf("--project1 is required")
	}

	ctx := context.Background()
	fetcher := gcp.NewResourceFetcher()

	// Check if this is a known resource type
	resourceType, isKnown := gcp.ResourceTypes[resourceTypeStr]

	// Build flags for resource 1
	flags1 := buildResourceFlags(cmd, "1")

	// Build flags for resource 2
	flags2 := buildResourceFlags(cmd, "2")

	// Fetch resources
	var resource1, resource2 map[string]interface{}
	var err error

	if isKnown {
		// Use structured fetcher for known resource types
		fmt.Fprintf(cmd.OutOrStderr(), "Fetching %s '%s' from %s...\n", resourceTypeStr, name1, project1)
		resource1, err = fetcher.FetchResource(ctx, resourceType, name1, project1, flags1)
		if err != nil {
			return fmt.Errorf("failed to fetch %s '%s': %w", resourceTypeStr, name1, err)
		}

		fmt.Fprintf(cmd.OutOrStderr(), "Fetching %s '%s' from %s...\n", resourceTypeStr, name2, project2)
		resource2, err = fetcher.FetchResource(ctx, resourceType, name2, project2, flags2)
		if err != nil {
			return fmt.Errorf("failed to fetch %s '%s': %w", resourceTypeStr, name2, err)
		}
	} else {
		// Treat resourceTypeStr as a custom gcloud resource path
		// Build gcloud commands manually
		gcloudCmd1 := buildGcloudCommand(resourceTypeStr, name1, project1, flags1)
		gcloudCmd2 := buildGcloudCommand(resourceTypeStr, name2, project2, flags2)

		fmt.Fprintf(cmd.OutOrStderr(), "Fetching resource with: gcloud %s...\n", gcloudCmd1)
		resource1, err = fetcher.FetchResourceGeneric(ctx, gcloudCmd1)
		if err != nil {
			return fmt.Errorf("failed to fetch resource: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStderr(), "Fetching resource with: gcloud %s...\n", gcloudCmd2)
		resource2, err = fetcher.FetchResourceGeneric(ctx, gcloudCmd2)
		if err != nil {
			return fmt.Errorf("failed to fetch resource: %w", err)
		}
	}

	// Load config for field filtering
	cfg, err := config.Load(viper.ConfigFileUsed())
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	// If comparing within the same project, ignore resource-specific identifiers
	if project1 == project2 {
		cfg.IgnoreFields = append(cfg.IgnoreFields,
			"name",
			"self_link",
			"selfLink",
		)
	}

	// Compare and output
	differ := compare.NewDiffer(cfg, viper.GetBool("show-all"))
	diff := differ.Compare(resource1, resource2)

	format := viper.GetString("format")
	switch format {
	case "json":
		output, _ := json.MarshalIndent(diff, "", "  ")
		fmt.Println(string(output))
	case "diff":
		fallthrough
	default:
		compare.PrintGitStyleDiffV2(cmd.OutOrStdout(), diff, name1, name2)
	}

	return nil
}

func buildResourceFlags(cmd *cobra.Command, suffix string) map[string]string {
	flags := make(map[string]string)

	if zone, _ := cmd.Flags().GetString("zone" + suffix); zone != "" {
		flags["zone"] = zone
	} else if zone1, _ := cmd.Flags().GetString("zone1"); suffix == "2" && zone1 != "" {
		// Default to zone1 for resource 2
		flags["zone"] = zone1
	}

	if region, _ := cmd.Flags().GetString("region" + suffix); region != "" {
		flags["region"] = region
	} else if region1, _ := cmd.Flags().GetString("region1"); suffix == "2" && region1 != "" {
		// Default to region1 for resource 2
		flags["region"] = region1
	}

	if location, _ := cmd.Flags().GetString("location" + suffix); location != "" {
		flags["location"] = location
	} else if location1, _ := cmd.Flags().GetString("location1"); suffix == "2" && location1 != "" {
		// Default to location1 for resource 2
		flags["location"] = location1
	}

	return flags
}

func buildGcloudCommand(resourcePath, name, project string, flags map[string]string) string {
	parts := []string{resourcePath, "describe", name}

	if project != "" {
		parts = append(parts, "--project="+project)
	}

	for key, value := range flags {
		if value != "" {
			parts = append(parts, fmt.Sprintf("--%s=%s", key, value))
		}
	}

	return strings.Join(parts, " ")
}
