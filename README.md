# SPA Terraform Provider

This Terraform provider manages Citrix SPA (Secure Private Access) resources through the SPA API.

## Features

- **Applications**: Manage web, SaaS, and ZTNA applications
- **Access Policies**: Configure access control policies
- **Security Groups**: Manage user and group security configurations
- **Routing Domains**: Configure domain routing for applications
- **Certificates**: Manage SSL certificates for applications
- **Browser Mode**: Query and manage browser mode configuration (CEB/CEP)
- **Hybrid Configuration**: Query hybrid deployment settings
- **Last Activity**: Track last activity timestamps
- **Terminate Machine Access**: Manage machine access termination

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
go build -o terraform-provider-spa
```

## Authentication

The provider supports the following authentication methods:

### Environment Variables

```bash
export CITRIX_CUSTOMER_ID="your-customer-id"
export CITRIX_AUTH_TOKEN="your-auth-token"
```

### Provider Configuration

```hcl
provider "spa" {
  base_url    = "https://api.cloud.com/accessSecurity"
  customer_id = "your-customer-id"
  auth_token  = "your-auth-token"
}
```

## Usage Examples

### Web Application

```hcl
resource "spa_application" "web_app" {
  name        = "My Web Application"
  type        = "web"
  description = "A sample web application"
  url         = "https://example.com"
  category    = "Productivity"

  using_template = true
  related_urls   = ["https://example.com", "https://api.example.com"]

  hidden           = false
  agentless_access = true
  mobile_security  = true

  keywords = ["productivity", "web", "example"]
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

  destination {
    destination = "10.0.1.0/24"
    port        = "3306"
    protocol    = "PROTOCOL_TCP"
    subtype     = "SUBTYPE_IP_AND_CIDR"
  }
}
```

### Access Policy

```hcl
resource "spa_access_policy" "require_mfa" {
  name        = "Require MFA for Sensitive Apps"
  description = "Require multi-factor authentication for sensitive applications"
  type        = "access"
  enabled     = true
  priority    = 100

  conditions = {
    user_risk    = "high"
    device_trust = "untrusted"
  }

  actions = {
    require_mfa = "true"
    block_access = "false"
  }
}
```

### Security Group

```hcl
resource "spa_security_group" "developers" {
  name        = "Developers"
  description = "Development team security group"
  type        = "user_group"

  users = [
    "user1@example.com",
    "user2@example.com"
  ]

  groups = [
    "CN=Developers,OU=Groups,DC=example,DC=com"
  ]

  rules {
    type     = "access"
    action   = "allow"
    priority = 1

    conditions = {
      time_of_day = "business_hours"
      location    = "office"
    }
  }
}
```

### Routing Domain

```hcl
resource "spa_routing_domain" "internal_domain" {
  fqdn     = "internal.example.com"
  type     = "internal"
  app_type = "web"
  comment  = "Internal domain for web applications"
  flag     = "enabled"

  location_ids = [
    "location-uuid-1",
    "location-uuid-2"
  ]
}
```

### Certificate

```hcl
resource "spa_certificate" "app_cert" {
  certificate_name = "example-app-cert"
  certificate      = base64encode(file("${path.module}/certs/app.pfx"))

  # Optional: Assign to application
  application_id = spa_application.web_app.id
  domain         = "example.com"
}
```

### Browser Mode Configuration

```hcl
# Query current browser mode
data "spa_browser_mode" "current" {}

# Browser mode resource (typically read-only)
resource "spa_browser_mode" "config" {
  browser_mode = "CEB"  # or "CEP"
}
```

### Hybrid Configuration

```hcl
# Query hybrid configuration
data "spa_hybrid_config" "current" {}

output "is_hybrid_deployment" {
  value = data.spa_hybrid_config.current.is_hybrid
}
```

### Last Activity Tracking

```hcl
# Get last activity timestamp
data "spa_last_activity" "current" {}

output "last_activity_time" {
  value = data.spa_last_activity.current.last_activity
}
```

### Terminate Machine Access

```hcl
# Manage machine access termination
resource "spa_terminate_machine_access" "machine" {
  name   = "workstation-001"
  status = "active"
}

# Query specific machine access
data "spa_terminate_machine_access" "machine_info" {
  id = spa_terminate_machine_access.machine.id
}
```

## Data Sources

### Application Data Source

```hcl
data "spa_application" "existing_app" {
  name = "Existing Application"
}

# Or by ID
data "spa_application" "existing_app_by_id" {
  id = "app-id-123"
}
```

### Access Policy Data Source

```hcl
data "spa_access_policy" "existing_policy" {
  name = "Default Policy"
}
```

### Security Group Data Source

```hcl
data "spa_security_group" "existing_group" {
  name = "All Users"
}
```

### Routing Domain Data Source

```hcl
data "spa_routing_domain" "existing_domain" {
  fqdn = "api.example.com"
}
```

## Resource Reference

### Application Types

- `web`: Web applications accessed via URL
- `saas`: Software-as-a-Service applications
- `ztna`: Zero Trust Network Access applications

### Destination Protocols

- `PROTOCOL_TCP`: TCP protocol
- `PROTOCOL_UDP`: UDP protocol

### Destination Subtypes

- `SUBTYPE_HOSTNAME`: Hostname-based destination
- `SUBTYPE_IP_AND_CIDR`: IP address and CIDR notation
- `SUBTYPE_IP_RANGE`: IP address range

### Routing Entry Types

- `internal`: Internal routing
- `external`: External routing
- `external_via_connector`: External routing via connector
- `conflicting`: Conflicting routing entry
- `internal_bypass_proxy`: Internal routing bypassing proxy

## Error Handling

The provider includes comprehensive error handling for:

- Authentication failures
- API rate limiting
- Resource not found errors
- Validation errors
- Network connectivity issues

## Local Development

This repository includes a local testing environment for development and testing purposes.

### Prerequisites

- Go 1.21 or higher
- Terraform 1.0 or higher
- Citrix Cloud API credentials (either direct token or service principal)

### Quick Start

1. **Clone the repository**:

   ```bash
   git clone https://github.com/citrix/terraform-provider-spa.git
   cd terraform-provider-spa
   ```

2. **Set up authentication** (choose one method):

   **Option A: Direct Token Authentication**

   ```bash
   export SPA_TOKEN="your-bearer-token"
   export SPA_CUSTOMER_ID="your-customer-id"
   ```

   **Option B: Service Principal Authentication (Recommended)**

   ```bash
   export SPA_CLIENT_ID="your-client-id"
   export SPA_CLIENT_SECRET="your-client-secret"
   export SPA_CUSTOMER_ID="your-customer-id"
   ```

3. **Run the local test environment**:

   ```bash
   # Test with direct token
   ./test-local.sh

   # Test with service principal
   ./test-local.sh --service-principal

   # Apply changes instead of just planning
   ./test-local.sh --sp --apply
   ```

### Plugin Directory Configuration

The local testing environment uses a persistent plugin directory to avoid rebuilding across system reboots:

- **Default Location**: `~/.terraform/plugins`
- **Environment Variable**: `TF_PLUGIN_DIR=/custom/path ./test-local.sh`
- **Command Line Option**: `./test-local.sh --plugin-dir /custom/path`

Example with custom plugin directory:

```bash
# Using command line argument
./test-local.sh --plugin-dir /tmp/tf-plugins

# Using environment variable
export TF_PLUGIN_DIR=/tmp/tf-plugins
./test-local.sh
```

### Available Make Targets

- `make build` - Build the provider
- `make install-local` - Install provider locally for testing
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make docs` - Generate documentation

### Testing with Different Authentication Methods

The test environment supports both authentication methods:

1. **Direct Token**: Quick testing with existing token
2. **Service Principal**: Production-like authentication flow

Use `./test-local.sh --help` to see all available options.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Architecture

### Overview

This Terraform provider enables Infrastructure as Code management of Citrix SPA (Secure Private Access) resources using the OpenAPI specification. The provider supports full CRUD operations for key SPA resources and follows Terraform best practices.

### Provider Structure

```
terraform-provider-spa/
├── main.go                          # Provider entry point
├── go.mod                          # Go module dependencies
├── internal/
│   └── provider/
│       ├── provider.go             # Main provider implementation
│       ├── client.go               # API client with HTTP operations
│       ├── resource_application.go # Application resource (web, SaaS, ZTNA)
│       ├── resource_access_policy.go # Access policy resource
│       ├── resource_security_group.go # Security group resource
│       ├── resource_routing_domain.go # Routing domain resource
│       ├── resource_certificate.go # Certificate resource
│       ├── resource_browser_mode.go # Browser mode resource
│       ├── resource_terminate_machine_access.go # Machine access termination
│       └── data_source_*.go        # Various data sources
├── examples/                       # Usage examples
├── resource-listing-tool/          # Resource discovery and TF generation
├── Makefile                        # Build and development tasks
└── README.md                       # This documentation
```

### Key Components

#### API Client (`client.go`)

- HTTP client with proper header management
- Citrix Cloud authentication headers (updated to use `Citrix-CustomerId`)
- Base path support for `/accessSecurity` endpoint
- Comprehensive error handling and request/response logging
- Support for all SPA API endpoints

#### Resources and Data Sources

The provider includes comprehensive resource management for:

- **Applications**: Web, SaaS, and ZTNA applications with full lifecycle management
- **Access Policies**: Policy conditions and actions with priority-based ordering
- **Security Groups**: User and group management with rule-based access control
- **Routing Domains**: Domain routing configuration with location support
- **Certificates**: SSL certificate management with application binding
- **Browser Mode**: Browser mode configuration and querying
- **Terminate Machine Access**: Machine access termination management

For detailed architecture information, see the complete provider structure and component documentation.

## Authentication Methods

The provider supports multiple authentication methods to accommodate different deployment scenarios and security requirements.

### Direct Token Authentication

For quick testing and development environments:

```bash
# Environment variables
export CITRIX_CUSTOMER_ID="your-customer-id"
export CITRIX_AUTH_TOKEN="your-auth-token"

# Alternative variable names
export SPA_CUSTOMER_ID="your-customer-id"
export SPA_TOKEN="your-bearer-token"
```

```hcl
# Provider configuration
provider "spa" {
  base_url    = "https://api.cloud.com/accessSecurity"
  customer_id = "your-customer-id"
  auth_token  = "your-auth-token"
}
```

### Service Principal Authentication (Recommended)

For production environments and CI/CD pipelines:

```bash
# Environment variables
export SPA_CLIENT_ID="your-client-id"
export SPA_CLIENT_SECRET="your-client-secret"
export SPA_CUSTOMER_ID="your-customer-id"
```

```hcl
# Provider configuration
provider "spa" {
  base_url      = "https://api.cloud.com/accessSecurity"
  customer_id   = "your-customer-id"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}
```

### Authentication Flow

1. **Token Validation**: The provider validates the provided credentials
2. **OAuth2 Flow**: For service principal authentication, the provider exchanges credentials for access tokens
3. **API Authorization**: All API requests include proper Citrix Cloud authentication headers
4. **Token Refresh**: Automatic token refresh for long-running operations

### Security Best Practices

- Use service principal authentication for production deployments
- Store credentials in secure environment variables or secret management systems
- Implement proper credential rotation policies
- Use least-privilege access principles for service principals
- Enable audit logging for authentication events

## Development Guide

### Prerequisites

- Go 1.21 or higher
- Terraform 1.0 or higher
- Git version control
- Make build tool
- Citrix Cloud API credentials

### Development Environment Setup

1. **Clone the repository**:

   ```bash
   git clone https://github.com/citrix/terraform-provider-spa.git
   cd terraform-provider-spa
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   go mod tidy
   ```

3. **Set up authentication** (choose one method):

   **Direct Token Authentication**:

   ```bash
   export SPA_TOKEN="your-bearer-token"
   export SPA_CUSTOMER_ID="your-customer-id"
   ```

   **Service Principal Authentication**:

   ```bash
   export SPA_CLIENT_ID="your-client-id"
   export SPA_CLIENT_SECRET="your-client-secret"
   export SPA_CUSTOMER_ID="your-customer-id"
   ```

4. **Build and install the provider**:

   ```bash
   make build
   make install-local
   ```

### Available Make Targets

- `make build` - Build the provider binary
- `make install-local` - Install provider locally for testing
- `make clean` - Clean build artifacts and temporary files
- `make test` - Run unit tests
- `make testacc` - Run acceptance tests (requires API credentials)
- `make docs` - Generate documentation
- `make fmt` - Format Go code
- `make lint` - Run linting tools
- `make dev-install` - Combined build and install for development

### Development Workflow

1. **Feature Development**:

   - Create a feature branch from main
   - Implement changes following Go best practices
   - Add comprehensive tests for new functionality
   - Update documentation as needed

2. **Testing**:

   - Run unit tests: `make test`
   - Run acceptance tests: `make testacc`
   - Test with local examples: `./test-local.sh`

3. **Code Quality**:

   - Format code: `make fmt`
   - Run linting: `make lint`
   - Ensure all tests pass

4. **Documentation**:
   - Update README for new features
   - Generate provider documentation: `make docs`
   - Update examples as needed

### Project Structure

```
terraform-provider-spa/
├── internal/provider/       # Provider implementation
│   ├── provider.go         # Main provider setup
│   ├── client.go           # API client
│   ├── resource_*.go       # Resource implementations
│   ├── data_source_*.go    # Data source implementations
│   └── provider_test.go    # Test setup
├── examples/               # Usage examples
├── docs/                   # Generated documentation
├── resource-listing-tool/  # Resource discovery tool
├── test-local/            # Local testing environment
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── main.go                # Provider entry point
```

### Adding New Resources

1. **Create resource file**: `internal/provider/resource_new_resource.go`
2. **Implement CRUD operations**: Create, Read, Update, Delete
3. **Add schema definition**: Define Terraform resource schema
4. **Add to provider**: Register resource in `provider.go`
5. **Write tests**: Add unit and acceptance tests
6. **Update documentation**: Add usage examples

### Adding New Data Sources

1. **Create data source file**: `internal/provider/data_source_new_data.go`
2. **Implement read operation**: Data retrieval logic
3. **Add schema definition**: Define data source schema
4. **Add to provider**: Register data source in `provider.go`
5. **Write tests**: Add comprehensive tests
6. **Update documentation**: Add usage examples

### Testing Guidelines

- Write unit tests for all new functionality
- Use acceptance tests for API integration testing
- Mock external dependencies in unit tests
- Test error scenarios and edge cases
- Maintain test coverage above 80%

### Contributing Guidelines

1. Fork the repository
2. Create a feature branch
3. Follow Go coding standards
4. Write comprehensive tests
5. Update documentation
6. Submit a pull request

## Local Testing

### Overview

The local testing environment provides a comprehensive setup for testing the SPA Terraform provider in a controlled environment before deploying to production.

### Testing Environment Structure

```
test-local/
├── main.tf                 # Test configuration
├── terraform.tfvars       # Local variables (gitignored)
├── terraform.tfvars.example # Example variables
├── tfplan                 # Terraform plan output
└── terraform-debug.log    # Debug log output
```

### Prerequisites

- Terraform 1.0 or higher
- Go 1.21 or higher for building from source
- Valid Citrix Cloud credentials
- Network connectivity to Citrix Cloud API

### Quick Start

1. **Set up credentials**:

   ```bash
   cd test-local
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your credentials
   ```

2. **Run tests**:

   ```bash
   # From project root
   ./test-local.sh

   # With service principal
   ./test-local.sh --service-principal

   # Apply changes
   ./test-local.sh --apply
   ```

### Authentication Configuration

#### Direct Token Authentication

```hcl
# In terraform.tfvars
spa_token = "your-bearer-token"
spa_customer_id = "your-customer-id"
```

#### Service Principal Authentication

```hcl
# In terraform.tfvars
spa_client_id = "your-client-id"
spa_client_secret = "your-client-secret"
spa_customer_id = "your-customer-id"
```

### Test Configuration Options

The test environment supports various configuration options:

#### Environment Variables

```bash
# Base URL configuration
export SPA_BASE_URL="https://api.cloud.com/accessSecurity"

# Plugin directory
export TF_PLUGIN_DIR="~/.terraform/plugins"

# Debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH="terraform-debug.log"
```

#### Command Line Options

```bash
# Show help
./test-local.sh --help

# Use service principal authentication
./test-local.sh --service-principal
./test-local.sh --sp

# Apply changes instead of just planning
./test-local.sh --apply

# Use custom plugin directory
./test-local.sh --plugin-dir /custom/path

# Skip provider build
./test-local.sh --skip-build
```

### Test Scenarios

#### Basic Resource Testing

```hcl
# Test web application
resource "spa_application" "test_web" {
  name           = "Test Web App"
  type           = "web"
  url            = "https://example.com"
  using_template = true
  related_urls   = ["https://example.com"]
}

# Test ZTNA application
resource "spa_application" "test_ztna" {
  name = "Test ZTNA App"
  type = "ztna"

  destination {
    destination = "internal.example.com"
    port        = "443"
    protocol    = "PROTOCOL_TCP"
    subtype     = "SUBTYPE_HOSTNAME"
  }
}
```

#### Access Policy Testing

```hcl
resource "spa_access_policy" "test_policy" {
  name     = "Test Policy"
  enabled  = true
  priority = 100

  conditions = {
    user_risk = "medium"
  }

  actions = {
    require_mfa = "true"
  }
}
```

#### Security Group Testing

```hcl
resource "spa_security_group" "test_group" {
  name = "Test Security Group"
  type = "user_group"

  users = [
    "testuser@example.com"
  ]

  rules {
    type     = "access"
    action   = "allow"
    priority = 1
  }
}
```

### Plugin Directory Configuration

The test environment uses a persistent plugin directory to avoid rebuilding:

```bash
# Default location
~/.terraform/plugins

# Custom location via environment variable
export TF_PLUGIN_DIR=/tmp/tf-plugins
./test-local.sh

# Custom location via command line
./test-local.sh --plugin-dir /tmp/tf-plugins
```

### Debug Logging

Enable detailed logging for troubleshooting:

```bash
# Enable debug logging
export TF_LOG=DEBUG
export TF_LOG_PATH="terraform-debug.log"

# Run with debug output
./test-local.sh --debug
```

### Troubleshooting

#### Common Issues

1. **Authentication Errors**:

   - Verify credentials in `terraform.tfvars`
   - Check environment variables
   - Ensure correct base URL

2. **Provider Not Found**:

   - Rebuild provider: `make build`
   - Check plugin directory permissions
   - Verify provider installation

3. **API Errors**:
   - Check network connectivity
   - Verify API endpoint availability
   - Review debug logs

#### Debug Steps

1. **Check provider build**:

   ```bash
   make build
   ls -la terraform-provider-spa
   ```

2. **Verify plugin installation**:

   ```bash
   ls -la ~/.terraform/plugins/
   ```

3. **Test authentication**:

   ```bash
   export TF_LOG=DEBUG
   ./test-local.sh --debug
   ```

4. **Review logs**:
   ```bash
   tail -f terraform-debug.log
   ```

### Performance Testing

The test environment supports performance testing scenarios:

```bash
# Test with multiple resources
./test-local.sh --scale-test

# Measure execution time
time ./test-local.sh --apply

# Test concurrent operations
./test-local.sh --parallel
```

### Cleanup

Clean up test resources:

```bash
# Destroy test infrastructure
cd test-local
terraform destroy

# Clean temporary files
rm -f terraform-debug.log tfplan
```

## Base URL Configuration

### Overview

The SPA Terraform provider supports configurable base URLs to accommodate different Citrix Cloud deployments, regions, and environments. This flexibility allows the provider to work with various Citrix Cloud instances worldwide.

### Default Configuration

The provider uses the following default base URL:

```
https://api.cloud.com
```

### Configuration Methods

#### 1. Environment Variable

```bash
export SPA_BASE_URL="https://api.cloud.com/accessSecurity"
```

#### 2. Provider Configuration

```hcl
provider "spa" {
  base_url = "https://api.cloud.com/accessSecurity"
  # ... other configuration
}
```

#### 3. Terraform Variables

```hcl
variable "spa_base_url" {
  description = "Base URL for SPA API"
  type        = string
  default     = "https://api.cloud.com/accessSecurity"
}

provider "spa" {
  base_url = var.spa_base_url
}
```

### Regional Endpoints

Different regions may have different base URLs:

````hcl
```hcl
# US region
provider "spa" {
  base_url = "https://api.cloud.com/accessSecurity"
}

# EU region
provider "spa" {
  base_url = "https://api.eu.cloud.com/accessSecurity"
}
````

# Asia-Pacific region

provider "spa" {
base_url = "https://api.ap.cloud.com"

````

### Environment-Specific URLs

Configure different URLs for different environments:

```hcl
# Development environment
provider "spa" {
  base_url = "https://api-dev.cloud.com"
}

# Staging environment
provider "spa" {
  base_url = "https://api-staging.cloud.com/accessSecurity"
}

# Production environment
provider "spa" {
  base_url = "https://api.cloud.com/accessSecurity"
}
````

### URL Validation

The provider validates base URLs to ensure they:

- Use HTTPS protocol
- Have valid hostname format
- Are reachable and responsive
- Return expected API responses

### Migration Between URLs

When migrating between different base URLs:

1. **Update configuration**:

   ```hcl
   provider "spa" {
     base_url = "https://new-api.cloud.com/accessSecurity"
   }
   ```

2. **Test connectivity**:

   ```bash
   terraform plan
   ```

3. **Apply changes**:
   ```bash
   terraform apply
   ```

### Best Practices

- Use environment variables for different deployment environments
- Validate URL changes in non-production environments first
- Document URL changes in infrastructure documentation
- Monitor API connectivity after URL changes
- Use regional endpoints for optimal performance

## Debug Logging

### Overview

The SPA Terraform provider supports comprehensive debug logging to help troubleshoot issues, monitor API interactions, and understand provider behavior during operations.

### Enabling Debug Logging

#### Environment Variables

```bash
# Enable debug logging
export TF_LOG=DEBUG

# Specify log file location
export TF_LOG_PATH="terraform-debug.log"

# Enable provider-specific logging
export TF_LOG_PROVIDER=DEBUG
```

#### Log Levels

Available log levels (in order of verbosity):

- `TRACE` - Most verbose, shows all details
- `DEBUG` - Detailed information for debugging
- `INFO` - General information about operations
- `WARN` - Warning messages
- `ERROR` - Error messages only

```bash
# Set specific log level
export TF_LOG=TRACE
export TF_LOG=DEBUG
export TF_LOG=INFO
export TF_LOG=WARN
export TF_LOG=ERROR
```

### Provider-Specific Logging

The SPA provider includes detailed logging for:

- **API Requests**: HTTP method, URL, headers, and request body
- **API Responses**: Status code, headers, and response body
- **Authentication**: Token validation and refresh operations
- **Resource Operations**: Create, read, update, and delete operations
- **Error Handling**: Detailed error messages and stack traces

### Debug Log Examples

#### API Request Logging

```
2024-01-15T10:30:00.000Z [DEBUG] provider.terraform-provider-spa: API Request:
  Method: GET
  URL: https://api.cloud.com/accessSecurity/applications
  Headers:
    Citrix-CustomerId: your-customer-id
    Authorization: Bearer your-token
    Content-Type: application/json
```

#### API Response Logging

```
2024-01-15T10:30:00.100Z [DEBUG] provider.terraform-provider-spa: API Response:
  Status: 200 OK
  Headers:
    Content-Type: application/json
    Content-Length: 1234
  Body: {"items":[{"id":"app1","name":"Test App"}]}
```

#### Resource Operation Logging

```
2024-01-15T10:30:00.200Z [DEBUG] provider.terraform-provider-spa: Creating application:
  Name: Test Application
  Type: web
  URL: https://example.com

2024-01-15T10:30:00.300Z [DEBUG] provider.terraform-provider-spa: Application created:
  ID: app-123456
  Status: active
```

### Structured Logging

The provider uses structured logging with consistent fields:

```json
{
  "timestamp": "2024-01-15T10:30:00.000Z",
  "level": "DEBUG",
  "component": "spa-provider",
  "operation": "create_application",
  "resource_type": "spa_application",
  "resource_name": "test_app",
  "message": "Creating application resource",
  "details": {
    "api_endpoint": "/accessSecurity/applications",
    "request_id": "req-123456"
  }
}
```

### Log Analysis

#### Common Debug Patterns

1. **Authentication Issues**:

   ```bash
   grep -i "auth\|token\|401\|403" terraform-debug.log
   ```

2. **API Errors**:

   ```bash
   grep -i "error\|400\|500" terraform-debug.log
   ```

3. **Resource Operations**:

   ```bash
   grep -i "creating\|updating\|deleting" terraform-debug.log
   ```

4. **Performance Issues**:
   ```bash
   grep -i "timeout\|slow\|latency" terraform-debug.log
   ```

### Security Considerations

#### Sensitive Data Redaction

The provider automatically redacts sensitive information:

- Authentication tokens
- Client secrets
- Password fields
- Certificate private keys

#### Safe Logging Practices

- Never commit log files to version control
- Use secure file permissions for log files
- Rotate log files regularly
- Filter sensitive data from shared logs

### Log File Management

#### Log Rotation

```bash
# Rotate logs daily
export TF_LOG_PATH="terraform-$(date +%Y%m%d).log"

# Compress old logs
gzip terraform-*.log
```

#### Log Cleanup

```bash
# Remove old log files
find . -name "terraform-*.log" -mtime +7 -delete

# Clean up test logs
rm -f test-local/terraform-debug.log
```

### Troubleshooting with Logs

#### Step-by-Step Debugging

1. **Enable debug logging**:

   ```bash
   export TF_LOG=DEBUG
   export TF_LOG_PATH="debug.log"
   ```

2. **Run problematic operation**:

   ```bash
   terraform plan
   ```

3. **Analyze logs**:

   ```bash
   # Check for errors
   grep -i error debug.log

   # Check API calls
   grep -i "api request\|api response" debug.log

   # Check authentication
   grep -i "auth\|token" debug.log
   ```

4. **Identify issues**:
   - Authentication failures
   - API endpoint errors
   - Resource conflicts
   - Network connectivity issues

### Performance Monitoring

#### Request Timing

```bash
# Monitor API response times
grep -i "response time\|duration" terraform-debug.log

# Check for slow operations
grep -i "slow\|timeout" terraform-debug.log
```

#### Resource Metrics

```bash
# Count API calls by endpoint
grep "API Request" terraform-debug.log | \
  grep -o "/accessSecurity/[^?]*" | \
  sort | uniq -c
```

### Integration with Monitoring Tools

#### Log Aggregation

```bash
# Send logs to syslog
export TF_LOG_PATH="/dev/stdout" | logger -t terraform-spa

# Forward to logging service
export TF_LOG_PATH="/dev/stdout" | \
  curl -X POST -H "Content-Type: application/json" \
  -d @- https://logs.example.com/api/v1/logs
```

#### Monitoring Alerts

Set up alerts for:

- Authentication failures
- API error rates
- Resource operation failures
- Performance degradation

## Service Principal Authentication Enhancement

### Overview

The SPA Terraform provider has been enhanced to support Citrix Cloud Service Principal authentication in addition to the existing direct token authentication. This enhancement provides a more secure and scalable authentication method suitable for production environments and CI/CD pipelines.

### Background

Previously, the provider only supported direct token authentication using bearer tokens. While functional, this approach had limitations:

- Manual token management
- Limited token lifecycle control
- Less suitable for automated environments
- Potential security concerns with long-lived tokens

### Service Principal Authentication

Service Principal authentication uses OAuth2 client credentials flow, providing:

- Automated token management
- Improved security through short-lived tokens
- Better integration with CI/CD pipelines
- Enhanced audit capabilities

### Implementation Details

#### Authentication Flow

1. **Client Credentials Exchange**: The provider exchanges client ID and secret for an access token
2. **Token Validation**: The received token is validated for proper format and expiration
3. **API Authorization**: The token is used to authenticate API requests
4. **Token Refresh**: Tokens are automatically refreshed when approaching expiration

#### OAuth2 Endpoint

The provider uses the Citrix Cloud OAuth2 endpoint:

```
https://api.cloud.com/cctrustoauth2/root/tokens/clients
```

#### Token Management

- **Token Caching**: Tokens are cached in memory during provider execution
- **Automatic Refresh**: Tokens are refreshed automatically before expiration
- **Concurrent Safety**: Token operations are thread-safe for parallel resource operations

### Configuration

#### Environment Variables

```bash
# Service Principal credentials
export SPA_CLIENT_ID="your-client-id"
export SPA_CLIENT_SECRET="your-client-secret"
export SPA_CUSTOMER_ID="your-customer-id"
```

#### Provider Configuration

```hcl
provider "spa" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
  customer_id   = "your-customer-id"
  base_url      = "https://api.cloud.com/accessSecurity"
}
```

#### Terraform Variables

```hcl
variable "spa_client_id" {
  description = "Citrix Cloud Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "spa_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret"
  type        = string
  sensitive   = true
}

variable "spa_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
}

provider "spa" {
  client_id     = var.spa_client_id
  client_secret = var.spa_client_secret
  customer_id   = var.spa_customer_id
}
```

### Migration Guide

#### From Direct Token to Service Principal

1. **Obtain Service Principal credentials** from Citrix Cloud
2. **Update configuration** to use client credentials instead of direct token
3. **Test authentication** in non-production environment
4. **Deploy to production** with proper secret management

#### Configuration Changes

**Before (Direct Token)**:

```hcl
provider "spa" {
  auth_token  = "your-bearer-token"
  customer_id = "your-customer-id"
}
```

**After (Service Principal)**:

```hcl
provider "spa" {
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
  customer_id   = "your-customer-id"
}
```

### Security Considerations

#### Best Practices

- Store client credentials in secure secret management systems
- Use environment variables instead of hardcoded values
- Implement proper credential rotation policies
- Monitor authentication events and failures
- Use least-privilege principles for service principal permissions

#### Secret Management

```bash
# Using AWS Secrets Manager
export SPA_CLIENT_ID=$(aws secretsmanager get-secret-value \
  --secret-id prod/spa/client-id --query SecretString --output text)

# Using Azure Key Vault
export SPA_CLIENT_SECRET=$(az keyvault secret show \
  --vault-name mykeyvault --name spa-client-secret --query value -o tsv)

# Using HashiCorp Vault
export SPA_CLIENT_ID=$(vault kv get -field=client_id secret/spa)
```

### Testing

#### Test Script Enhancement

The `test-local.sh` script has been enhanced to support both authentication methods:

```bash
# Test with service principal
./test-local.sh --service-principal

# Test with direct token (legacy)
./test-local.sh --token

# Apply changes with service principal
./test-local.sh --sp --apply
```

#### Test Configuration

```hcl
# test-local/terraform.tfvars
spa_client_id = "your-client-id"
spa_client_secret = "your-client-secret"
spa_customer_id = "your-customer-id"
```

### Error Handling

#### Authentication Errors

The provider handles various authentication scenarios:

- Invalid client credentials
- Token expiration
- Network connectivity issues
- OAuth2 service unavailability

#### Error Messages

```
Error: Authentication failed
│
│   with provider["registry.terraform.io/citrix/spa"],
│   on main.tf line 1, in provider "spa":
│    1: provider "spa" {
│
│ Failed to authenticate with service principal: invalid client credentials
```

### Monitoring and Logging

#### Authentication Events

The provider logs authentication events for monitoring:

```
2024-01-15T10:30:00.000Z [INFO] provider.spa: Service principal authentication successful
2024-01-15T10:30:00.100Z [DEBUG] provider.spa: Token expires in 3600 seconds
2024-01-15T10:30:00.200Z [DEBUG] provider.spa: Token refresh scheduled
```

#### Metrics

Monitor key authentication metrics:

- Token acquisition success/failure rates
- Token refresh frequency
- Authentication error patterns
- API request authorization failures

### Backward Compatibility

The enhancement maintains full backward compatibility:

- Existing direct token authentication continues to work
- No breaking changes to provider configuration
- Existing Terraform configurations remain valid
- Gradual migration path available

### Future Enhancements

Planned improvements include:

- Support for additional authentication methods
- Enhanced token caching strategies
- Integration with external identity providers
- Advanced security features

## Support

For support and questions:

- GitHub Issues: [https://github.com/citrix/terraform-provider-spa/issues](https://github.com/citrix/terraform-provider-spa/issues)
- Citrix Developer Documentation: [https://developer.citrix.com](https://developer.citrix.com)
- Citrix Cloud Support: [https://support.citrix.com](https://support.citrix.com)
