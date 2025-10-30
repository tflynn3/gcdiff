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
This command dynamically supports ALL GCP resources.

Use the gcloud resource path (in quotes) matching the gcloud command structure:

Examples:
  # Compute instances (from: gcloud compute instances describe)
  gcdiff resource "compute instances" instance-1 instance-2 --project1=proj --zone1=us-central1-a

  # Storage buckets (from: gcloud storage buckets describe)
  gcdiff resource "storage buckets" bucket-1 bucket-2 --project1=proj

  # Firewall rules (from: gcloud compute firewall-rules describe)
  gcdiff resource "compute firewall-rules" rule-1 rule-2 --project1=proj

  # Cloud Run services (from: gcloud run services describe)
  gcdiff resource "run services" service-1 service-2 --project1=proj --region1=us-central1

  # Pub/Sub subscriptions (from: gcloud pubsub subscriptions describe)
  gcdiff resource "pubsub subscriptions" sub-1 sub-2 --project1=proj

  # Include IAM bindings in comparison
  gcdiff resource "pubsub subscriptions" sub-1 sub-2 --project1=proj --iam

  # GKE clusters (from: gcloud container clusters describe)
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

	// IAM policy flag
	resourceCmd.Flags().Bool("iam", false, "Include IAM policy bindings in comparison (fetches both resource and IAM policy)")
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

	includeIAM, _ := cmd.Flags().GetBool("iam")

	ctx := context.Background()
	fetcher := gcp.NewResourceFetcher()

	// Build flags for resource 1
	flags1 := buildResourceFlags(cmd, "1")

	// Build flags for resource 2
	flags2 := buildResourceFlags(cmd, "2")

	// Build gcloud commands
	gcloudCmd1 := buildGcloudCommand(resourceTypeStr, name1, project1, flags1)
	gcloudCmd2 := buildGcloudCommand(resourceTypeStr, name2, project2, flags2)

	// Fetch resources
	fmt.Fprintf(cmd.OutOrStderr(), "Fetching resource with: gcloud %s...\n", gcloudCmd1)
	resource1, err := fetcher.FetchResourceGeneric(ctx, gcloudCmd1)
	if err != nil {
		return fmt.Errorf("failed to fetch resource: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStderr(), "Fetching resource with: gcloud %s...\n", gcloudCmd2)
	resource2, err := fetcher.FetchResourceGeneric(ctx, gcloudCmd2)
	if err != nil {
		return fmt.Errorf("failed to fetch resource: %w", err)
	}

	// If --iam flag is set, also fetch IAM policies and merge them
	if includeIAM {
		iamCmd1 := buildGcloudIAMCommand(resourceTypeStr, name1, project1, flags1)
		iamCmd2 := buildGcloudIAMCommand(resourceTypeStr, name2, project2, flags2)

		fmt.Fprintf(cmd.OutOrStderr(), "Fetching IAM policy with: gcloud %s...\n", iamCmd1)
		iamPolicy1, err := fetcher.FetchResourceGeneric(ctx, iamCmd1)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: could not fetch IAM policy for %s: %v\n", name1, err)
		} else {
			resource1["iamPolicy"] = iamPolicy1
		}

		fmt.Fprintf(cmd.OutOrStderr(), "Fetching IAM policy with: gcloud %s...\n", iamCmd2)
		iamPolicy2, err := fetcher.FetchResourceGeneric(ctx, iamCmd2)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: could not fetch IAM policy for %s: %v\n", name2, err)
		} else {
			resource2["iamPolicy"] = iamPolicy2
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

func buildGcloudIAMCommand(resourcePath, name, project string, flags map[string]string) string {
	parts := []string{resourcePath, "get-iam-policy", name}

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
