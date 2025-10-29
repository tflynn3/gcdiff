package gcp

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/option"
)

// ComputeClient wraps the GCP Compute API client
type ComputeClient struct {
	instances *compute.InstancesClient
}

// NewComputeClient creates a new ComputeClient using Application Default Credentials
func NewComputeClient(ctx context.Context, opts ...option.ClientOption) (*ComputeClient, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}

	return &ComputeClient{
		instances: instancesClient,
	}, nil
}

// GetInstance retrieves a compute instance
func (c *ComputeClient) GetInstance(ctx context.Context, project, zone, instance string) (*computepb.Instance, error) {
	req := &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instance,
	}

	return c.instances.Get(ctx, req)
}

// Close closes the client
func (c *ComputeClient) Close() error {
	return c.instances.Close()
}
