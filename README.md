# SPA Terraform Provider

Terraform provider for managing Citrix Secure Private Access (SPA) resources through the SPA API.

## Features

- **Applications**: Manage web, SaaS, and ZTNA applications
- **Access Policies**: Configure access control policies with rules and conditions
- **Security Groups**: Manage security group configurations
- **Routing Domains**: Configure domain routing for applications
- **Certificates**: Manage SSL certificates for applications
- **Browser Mode**: Manage browser mode configuration (CEB/CEP)
- **Terminate Access**: Manage machine and user access termination

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for building from source)
- Citrix Cloud account with SPA service enabled

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    spa = {
      source = "citrix/spa"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/citrix/terraform-provider-spa.git
cd terraform-provider-spa
make build
```

## Authentication

The provider supports two authentication methods. Use only one at a time.

### Direct Token

```hcl
provider "spa" {
  base_url    = "https://api.cloud.com/accessSecurity"
  customer_id = "your-customer-id"
  auth_token  = "your-auth-token"
}
```

Or via environment variables:

```bash
export CITRIX_CUSTOMER_ID="your-customer-id"
export CITRIX_AUTH_TOKEN="your-auth-token"
```

### Service Principal (Recommended)

Uses OAuth2 client credentials flow with automatic token management and optional encrypted disk caching.

```hcl
provider "spa" {
  base_url      = "https://api.cloud.com/accessSecurity"
  customer_id   = "your-customer-id"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}
```

Or via environment variables:

```bash
export CITRIX_CUSTOMER_ID="your-customer-id"
export CITRIX_CLIENT_ID="your-client-id"
export CITRIX_CLIENT_SECRET="your-client-secret"
```

## Provider Configuration

```hcl
provider "spa" {
  base_url             = "https://api.cloud.com/accessSecurity"  # Optional, this is the default
  token_url            = "https://api.cloud.com"                 # Optional, derived from base_url
  customer_id          = "your-customer-id"
  auth_token           = "your-auth-token"                       # OR use client_id + client_secret
  rate_limit           = 15                                      # Requests per second (default: 15)
  fetch_details_on_list = false                                  # Batch-fetch details on list calls
  enable_token_cache   = true                                    # Encrypted disk token cache (default: true)
}
```

## Usage Examples

### Web Application

```hcl
resource "spa_application" "web_app" {
  name           = "My Web Application"
  type           = "web"
  description    = "A sample web application"
  url            = "https://example.com"
  category       = "Productivity"
  using_template = true
  related_urls   = ["https://example.com", "https://api.example.com"]
  keywords       = ["productivity", "web"]
}
```

### ZTNA Application

```hcl
resource "spa_application" "ztna_app" {
  name        = "Internal Database"
  type        = "ztna"
  description = "Internal database server"
  category    = "Database"

  destination {
    destination = "database.internal.com"
    port        = "5432"
    protocol    = "PROTOCOL_TCP"
    subtype     = "SUBTYPE_HOSTNAME"
  }
}
```

### Routing Domain

```hcl
resource "spa_routing_domain" "internal_domain" {
  fqdn    = "internal.example.com"
  type    = "internal"
  comment = "Internal domain for web applications"

  location_ids = [
    "location-uuid-1",
    "location-uuid-2"
  ]
}
```

### Data Sources

```hcl
data "spa_application" "existing" {
  name = "Existing Application"
}

data "spa_applications" "all" {
  offset = 0
  limit  = 100
}

data "spa_routing_domain" "domain" {
  fqdn = "api.example.com"
}
```

## Resource Reference

### Application Types

- `web` — Web applications accessed via URL
- `saas` — Software-as-a-Service applications
- `ztna` — Zero Trust Network Access applications

### Destination Protocols

- `PROTOCOL_TCP`, `PROTOCOL_UDP`

### Destination Subtypes

- `SUBTYPE_HOSTNAME`, `SUBTYPE_IP_AND_CIDR`, `SUBTYPE_IP_RANGE`

### Routing Domain Types

- `internal`, `external`, `external_via_connector`, `conflicting`, `internal_bypass_proxy`

## Local Development

### Prerequisites

- Go 1.21+
- Terraform 1.0+
- Citrix Cloud API credentials

### Using the Test Script

The `test-local.sh` script automates the full local dev workflow: build, install, generate `.terraformrc`, and run `terraform init/validate/plan`.

**1. Set up credentials** (choose one method):

For direct token auth:

```bash
cp test-local/terraform.tfvars.example test-local/terraform.tfvars
# Edit test-local/terraform.tfvars with your customer_id and auth_token
```

For service principal auth:

```bash
cp test-local-sp/terraform.tfvars.example test-local-sp/terraform.tfvars
# Edit test-local-sp/terraform.tfvars with your customer_id, client_id, and client_secret
```

**2. Run the script**:

```bash
./test-local.sh                # Plan with direct token auth
./test-local.sh --sp           # Plan with service principal auth
./test-local.sh --apply        # Plan and prompt to apply
./test-local.sh --sp --apply   # Plan with SP auth and prompt to apply
./test-local.sh --cleanup      # Remove all test artifacts (see below)
```

The script performs these steps automatically:

1. `make build` — compiles the provider binary
2. `make install` — copies the binary to `~/.terraform.d/plugins/registry.terraform.io/citrix/spa/0.1.0/<os>_<arch>/`
3. `./generate-terraformrc.sh` — creates a `.terraformrc` file that tells Terraform to use the local binary instead of downloading from the registry
4. `terraform init -reconfigure` — initializes Terraform with the local provider
5. `terraform validate` — validates the configuration
6. `terraform plan` — plans the changes

### Cleanup

When you are done testing, run the following to remove all artifacts created by the script:

```bash
./test-local.sh --cleanup
```

This removes:

| Artifact | Location |
|----------|----------|
| Built binary | `terraform-provider-spa` |
| Generated Terraform CLI config | `.terraformrc` |
| Terraform cache and lock file | `test-local/.terraform/`, `test-local/.terraform.lock.hcl` |
| Terraform cache and lock file | `test-local-sp/.terraform/`, `test-local-sp/.terraform.lock.hcl` |
| Plan files | `test-local/tfplan`, `test-local-sp/tfplan` |
| State files | `test-local/terraform.tfstate*`, `test-local-sp/terraform.tfstate*` |
| Installed plugin | `~/.terraform.d/plugins/registry.terraform.io/citrix/spa/` |

> **Note**: `terraform.tfvars` files in `test-local/` and `test-local-sp/` are **not** removed — your credentials are preserved.

### Manual Steps (if needed)

```bash
make install                                              # Build and install
./generate-terraformrc.sh                                 # Generate .terraformrc
cd test-local                                             # Or test-local-sp for service principal
TF_CLI_CONFIG_FILE=../.terraformrc terraform init         # Init with local provider
TF_CLI_CONFIG_FILE=../.terraformrc terraform plan         # Plan
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the provider binary |
| `make install` | Build and install to local plugin directory |
| `make test` | Run unit tests |
| `make testacc` | Run acceptance tests (requires `TF_ACC=1` and credentials) |
| `make fmt` | Format Go code and Terraform examples |
| `make lint` | Run golangci-lint |
| `make docs` | Generate provider documentation |
| `make clean` | Remove build artifacts |
| `make check` | Run fmt + lint + test |

### Debug Logging

```bash
TF_LOG=DEBUG TF_CLI_CONFIG_FILE=../.terraformrc terraform plan
```

Log levels: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`

To write logs to a file:

```bash
TF_LOG=DEBUG TF_LOG_PATH="debug.log" TF_CLI_CONFIG_FILE=../.terraformrc terraform plan
```

## Project Structure

```
terraform-provider-spa/
├── main.go                          # Provider entry point
├── Makefile                         # Build targets
├── generate-terraformrc.sh          # Generates .terraformrc for local dev
├── test-local.sh                    # Local testing automation script
├── .goreleaser.yml                  # Release build configuration
├── internal/provider/
│   ├── provider.go                  # Provider setup, schema, configuration
│   ├── auth.go                      # OAuth2 service principal authentication
│   ├── client.go                    # HTTP client, API methods, rate limiting
│   ├── token_persistence.go         # Encrypted disk token caching
│   ├── sso_type.go                  # Custom SSO type with semantic equality
│   ├── resource_application.go      # Application resource (CRUD)
│   ├── resource_access_policy.go    # Access policy resource (CRUD)
│   ├── resource_routing_domain.go   # Routing domain resource (CRUD)
│   ├── resource_security_group.go   # Security group resource (CRUD)
│   ├── resource_certificate.go      # Certificate resource
│   ├── resource_browser_mode.go     # Browser mode resource
│   ├── data_source_*.go             # Data source implementations
│   └── *_test.go                    # Tests
├── test-local/                      # Test config for direct token auth
├── test-local-sp/                   # Test config for service principal auth
├── examples/                        # Usage examples
├── docs/                            # Generated documentation
└── resource-listing-tool/           # Resource discovery and TF generation tool
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Run `make check` to verify formatting, linting, and tests
5. Submit a pull request

## Support

- GitHub Issues: [https://github.com/citrix/terraform-provider-spa/issues](https://github.com/citrix/terraform-provider-spa/issues)
- Citrix Developer Documentation: [https://developer.citrix.com](https://developer.citrix.com)
