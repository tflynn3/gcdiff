# gcdiff

A terminal tool for comparing and auditing GCP resources across projects.

`gcdiff` helps you identify exact differences between GCP resources, making it especially useful for:
- Validating Terraform configurations match live resources
- Replicating environments across projects
- Auditing infrastructure changes
- Cross-project resource comparison

## Features

- 🔍 **Deep comparison** of GCP resources with field-level granularity
- 🎨 **Git-style diff output** with color-coded changes
- ⚙️ **Configurable field filtering** to ignore auto-generated or timestamp fields
- 🔐 **Uses gcloud auth** - leverages your existing authentication
- 🌍 **Cross-project support** - compare resources across different GCP projects
- 📦 **Easy installation** via Homebrew

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

```bash
# Compare two compute instances in the same project
gcdiff compute instance-1 instance-2 \
  --project1=my-project \
  --zone1=us-central1-a

# Compare instances across projects
gcdiff compute prod-web staging-web \
  --project1=prod-project \
  --project2=staging-project \
  --zone1=us-central1-a \
  --zone2=us-west1-b

# Show all fields (including normally ignored ones)
gcdiff compute instance-1 instance-2 \
  --project1=my-project \
  --zone1=us-central1-a \
  --show-all
```

## Authentication

`gcdiff` uses [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). Make sure you're authenticated with gcloud:

```bash
gcloud auth application-default login
```

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

Currently supported:
- ✅ Compute Instances

Coming soon:
- 🔜 VPC Networks & Subnets
- 🔜 Firewall Rules
- 🔜 Cloud Storage Buckets
- 🔜 IAM Policies
- 🔜 And more...

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
