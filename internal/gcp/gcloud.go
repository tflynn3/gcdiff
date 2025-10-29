package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ResourceFetcher fetches GCP resources using gcloud CLI
type ResourceFetcher struct{}

// NewResourceFetcher creates a new ResourceFetcher
func NewResourceFetcher() *ResourceFetcher {
	return &ResourceFetcher{}
}

// ResourceType represents a GCP resource type with its gcloud command structure
type ResourceType struct {
	// Service is the gcloud service (e.g., "compute", "storage", "run")
	Service string
	// Resource is the resource type (e.g., "instances", "buckets", "services")
	Resource string
	// DescribeCommand is the describe subcommand (usually "describe")
	DescribeCommand string
	// RequiredFlags are flags required for this resource type
	RequiredFlags []string
}

// Common resource type mappings
var ResourceTypes = map[string]ResourceType{
	"compute": {
		Service:         "compute",
		Resource:        "instances",
		DescribeCommand: "describe",
		RequiredFlags:   []string{"zone"},
	},
	"storage": {
		Service:         "storage",
		Resource:        "buckets",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"firewall": {
		Service:         "compute",
		Resource:        "firewall-rules",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"network": {
		Service:         "compute",
		Resource:        "networks",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"subnet": {
		Service:         "compute",
		Resource:        "networks subnets",
		DescribeCommand: "describe",
		RequiredFlags:   []string{"region"},
	},
	"disk": {
		Service:         "compute",
		Resource:        "disks",
		DescribeCommand: "describe",
		RequiredFlags:   []string{"zone"},
	},
	"run": {
		Service:         "run",
		Resource:        "services",
		DescribeCommand: "describe",
		RequiredFlags:   []string{"region"},
	},
	"sql": {
		Service:         "sql",
		Resource:        "instances",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"pubsub-topic": {
		Service:         "pubsub",
		Resource:        "topics",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"pubsub-subscription": {
		Service:         "pubsub",
		Resource:        "subscriptions",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"iam-service-account": {
		Service:         "iam",
		Resource:        "service-accounts",
		DescribeCommand: "describe",
		RequiredFlags:   []string{},
	},
	"iam-policy": {
		Service:         "iam",
		Resource:        "service-accounts",
		DescribeCommand: "get-iam-policy",
		RequiredFlags:   []string{},
	},
	"pubsub-topic-iam-policy": {
		Service:         "pubsub",
		Resource:        "topics",
		DescribeCommand: "get-iam-policy",
		RequiredFlags:   []string{},
	},
}

// FetchResource fetches a resource using gcloud CLI
func (f *ResourceFetcher) FetchResource(ctx context.Context, resourceType ResourceType, name, project string, flags map[string]string) (map[string]interface{}, error) {
	// Build gcloud command
	args := []string{resourceType.Service}

	// Add resource type (may have multiple parts like "networks subnets")
	resourceParts := strings.Fields(resourceType.Resource)
	args = append(args, resourceParts...)

	// Add describe command
	args = append(args, resourceType.DescribeCommand)

	// Add resource name
	args = append(args, name)

	// Add project
	if project != "" {
		args = append(args, "--project="+project)
	}

	// Add required flags
	for _, flagName := range resourceType.RequiredFlags {
		if flagValue, exists := flags[flagName]; exists && flagValue != "" {
			args = append(args, fmt.Sprintf("--%s=%s", flagName, flagValue))
		}
	}

	// Add format flag for JSON output
	args = append(args, "--format=json")

	// Execute gcloud command
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gcloud command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gcloud output: %w\nOutput: %s", err, string(output))
	}

	return result, nil
}

// FetchResourceGeneric fetches any resource using a generic gcloud command
// This allows for maximum flexibility without predefined resource types
func (f *ResourceFetcher) FetchResourceGeneric(ctx context.Context, gcloudCommand string) (map[string]interface{}, error) {
	// Parse the gcloud command
	parts := strings.Fields(gcloudCommand)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty gcloud command")
	}

	// Add format flag if not present
	hasFormat := false
	for _, part := range parts {
		if strings.HasPrefix(part, "--format=") {
			hasFormat = true
			break
		}
	}

	if !hasFormat {
		parts = append(parts, "--format=json")
	}

	// Execute gcloud command
	cmd := exec.CommandContext(ctx, "gcloud", parts...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gcloud command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gcloud output: %w\nOutput: %s", err, string(output))
	}

	return result, nil
}

// ListSupportedResources returns a list of built-in supported resource types
func ListSupportedResources() []string {
	types := make([]string, 0, len(ResourceTypes))
	for k := range ResourceTypes {
		types = append(types, k)
	}
	return types
}
