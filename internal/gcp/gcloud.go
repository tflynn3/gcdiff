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
