# gcdiff

[![Tests](https://github.com/tflynn3/gcdiff/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/tflynn3/gcdiff/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tflynn3/gcdiff)](https://goreportcard.com/report/github.com/tflynn3/gcdiff)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/tflynn3/gcdiff)](https://github.com/tflynn3/gcdiff/releases)
[![Homebrew](https://img.shields.io/badge/Homebrew-tflynn3%2Ftap%2Fgcdiff-orange)](https://github.com/tflynn3/homebrew-tap)

A terminal tool for comparing and auditing GCP resources across projects.

`gcdiff` helps you identify exact differences between GCP resources, making it especially useful for:
- Validating Terraform configurations match live resources
- Replicating environments across projects
- Auditing infrastructure changes
- Cross-project resource comparison

## Features

- **Dynamic resource support** - compare **ANY** GCP resource type without explicit code
- **Deep comparison** of GCP resources with field-level granularity
- **Git-style diff output** with color-coded changes
- **Configurable field filtering** to ignore auto-generated or timestamp fields
- **Uses gcloud auth** - leverages your existing authentication
- **Cross-project support** - compare resources across different GCP projects
- **Easy installation** via Homebrew

### Instantly Supported Resources

Because gcdiff uses `gcloud` CLI under the hood, **all GCP resources are automatically supported**:
- Compute Engine (instances, disks, snapshots, images, etc.)
- Cloud Storage (buckets)
- Networking (VPCs, subnets, firewalls, routes, load balancers)
- Cloud Run (services, jobs)
- Cloud SQL (instances, databases)
- Pub/Sub (topics, subscriptions)
- IAM (service accounts, roles)
- GKE (clusters, node pools)
- And literally **every other GCP resource** that gcloud supports!

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tflynn3/tap/gcdiff
```

### From source

```bash
go install github.com/tflynn3/gcdiff/cmd/gcdiff@latest
```

### Binary releases

Download pre-built binaries from the [releases page](https://github.com/tflynn3/gcdiff/releases).

## Quick Start

### Compute Instances
```bash
# Compare two instances (backward-compatible command)
gcdiff compute instance-1 instance-2 \
  --project1=my-project \
  --zone1=us-central1-a

# Or use the universal resource command
gcdiff resource compute instance-1 instance-2 \
  --project1=my-project \
  --zone1=us-central1-a
```

### Any GCP Resource
```bash
# Firewall rules
gcdiff resource firewall my-rule-1 my-rule-2 \
  --project1=my-project

# Storage buckets
gcdiff resource storage my-bucket-1 my-bucket-2 \
  --project1=my-project

# Cloud Run services
gcdiff resource run my-service-1 my-service-2 \
  --project1=my-project \
  --region1=us-central1

# VPC networks
gcdiff resource network my-vpc-1 my-vpc-2 \
  --project1=my-project

# Cloud SQL instances
gcdiff resource sql my-db-1 my-db-2 \
  --project1=my-project
```

### Cross-Project Comparison
```bash
# Verify staging matches production
gcdiff resource compute web-server web-server \
  --project1=my-prod-project \
  --project2=my-staging-project \
  --zone1=us-central1-a
```

## Authentication

`gcdiff` uses [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). Make sure you're authenticated with gcloud:

```bash
gcloud auth application-default login
```

## Same-Project vs Cross-Project Comparisons

`gcdiff` automatically adjusts its behavior based on comparison context:

### Same-Project Comparison
When comparing resources within the same project (default behavior), `gcdiff` automatically ignores:
- Resource names (must be unique within a project)
- Self-links (contain resource names)
- Auto-generated IDs and timestamps

This is useful for comparing similar resources in the same environment.

### Cross-Project Comparison
When comparing resources across different projects (`--project1` ≠ `--project2`), `gcdiff` **includes** resource names and identifiers in the comparison. This is perfect for:
- Validating that a resource in one project matches another project (same name, same config)
- Terraform validation: ensuring your Terraform creates resources that match production
- Environment parity checks: prod vs staging

Example cross-project comparison:
```bash
# Verify staging instance matches prod configuration
gcdiff compute web-server web-server \
  --project1=my-prod-project \
  --project2=my-staging-project \
  --zone1=us-central1-a
```

If both instances are named "web-server" and have identical configurations, gcdiff will report no differences (except auto-generated fields like timestamps).

## Configuration

Create a `.gcdiff.yaml` file in your home directory or current directory to customize field filtering:

```yaml
# Fields to ignore when comparing (exact matches)
ignore_fields:
  - id
  - selfLink
  - creationTimestamp
  - fingerprint
  - kind
  - etag

# Regex patterns for fields to ignore
ignore_patterns:
  - ".*Timestamp$"
  - ".*Fingerprint$"
```

You can override the config location with `--config`:

```bash
gcdiff compute instance-1 instance-2 --config=/path/to/config.yaml
```

## Supported Resources

**All GCP resources are supported dynamically!**

The tool uses `gcloud` CLI commands under the hood, which means any resource that gcloud can describe is automatically supported. No code changes needed to add support for new resource types!

### Built-in Shortcuts

For convenience, these resource types have built-in shortcuts:
- `compute` - Compute Engine instances
- `firewall` - Compute Engine firewall rules
- `network` - VPC networks
- `subnet` - VPC subnets
- `disk` - Compute Engine disks
- `storage` - Cloud Storage buckets
- `run` - Cloud Run services
- `sql` - Cloud SQL instances
- `pubsub-topic` - Pub/Sub topics
- `pubsub-subscription` - Pub/Sub subscriptions
- `iam-service-account` - IAM service accounts

### Advanced Usage

You can also compare any GCP resource by providing the gcloud resource path:

```bash
# GKE clusters
gcdiff resource "container clusters" cluster-1 cluster-2 \
  --project1=my-project \
  --zone1=us-central1-a

# Cloud Functions
gcdiff resource "functions" function-1 function-2 \
  --project1=my-project \
  --region1=us-central1
```

## Example Output

```
Comparing: prod-web <-> staging-web
--------------------------------------------------------------------------------

Summary: 3 difference(s) found
  + 0 field(s)
  - 0 field(s)
  ~ 3 field(s)

Modified Fields:

  ~ machineType
      - "n1-standard-2"
      + "n1-standard-4"

  ~ disks[0].diskSizeGb
      - 50
      + 100

  ~ networkInterfaces[0].accessConfigs[0].natIP
      - "35.123.45.67"
      + "35.234.56.78"
```

## Use Cases

### Terraform Validation

When replicating a GCP environment with Terraform:

1. Deploy your Terraform configuration
2. Use `gcdiff` to compare the Terraform-created resource with the original
3. Identify any fields that don't match
4. Adjust your Terraform config or use `--show-all` to see ignored fields

```bash
gcdiff compute terraform-instance original-instance \
  --project1=terraform-project \
  --project2=original-project \
  --zone1=us-central1-a
```

### Environment Replication Audit

Verify that staging matches production:

```bash
gcdiff compute prod-app staging-app \
  --project1=prod \
  --project2=staging \
  --zone1=us-central1-a
```

## Development

```bash
# Clone the repo
git clone https://github.com/tflynn3/gcdiff.git
cd gcdiff

# Build
make build
# or: go build -o gcdiff ./cmd/gcdiff

# Run tests
make test

# Run tests with coverage
make coverage

# Format and lint
make fmt
make lint

# Run all checks
make check

# Run the binary
./gcdiff --help
```

### Running Tests

The project has comprehensive unit tests with >95% coverage for core logic:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...
```

Test structure:
- `internal/config/config_test.go` - Configuration loading and filtering tests
- `internal/compare/differ_test.go` - Comparison logic tests
- `internal/compare/output_test.go` - Output formatting tests
- `internal/compare/integration_test.go` - End-to-end integration tests

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.
