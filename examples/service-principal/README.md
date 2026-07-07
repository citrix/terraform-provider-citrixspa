# SPA Terraform Provider - Service Principal Authentication Example

This example demonstrates how to use the SPA Terraform provider with Service Principal authentication.

## Prerequisites

1. Citrix Cloud account with SPA service access
2. Service Principal created in Citrix Cloud with appropriate permissions
3. Terraform installed (version 0.13+)

## Setup Steps

### 1. Create Service Principal in Citrix Cloud

1. Log into [Citrix Cloud](https://citrix.cloud.com/)
2. Navigate to **Identity and Access Management** > **API Access** > **Service Principals**
3. Click **Create Service Principal**
4. Provide a name and description
5. Set appropriate permissions for SPA service
6. **Important**: Copy the Client ID and Client Secret - the secret is only shown once!

### 2. Configure Terraform Variables

Create `terraform.tfvars` with your credentials:

```hcl
# terraform.tfvars
citrix_customer_id   = "your-customer-id-here"
citrix_client_id     = "your-client-id-here"
citrix_client_secret = "your-client-secret-here"
```

### 3. Apply the Configuration

```bash
terraform init
terraform plan
terraform apply
```

## Configuration Files

### main.tf

```hcl
terraform {
  required_version = ">= 0.13"
  required_providers {
    spa = {
      source  = "citrix/citrixspa"
      version = "~> 0.1.0"
    }
  }
}

provider "spa" {
  base_url      = "https://api.cloud.com/accessSecurity"
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
}

# Variables
variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Service Principal Client Secret"
  type        = string
  sensitive   = true
}

# Data sources for existing resources
data "spa_application" "existing_apps" {
  # List existing applications
}

# Create a new application
resource "spa_application" "web_app" {
  name        = "My Web Application"
  description = "Web application managed by Terraform"
  # Add other required attributes based on your provider schema
}

# Create a security group
resource "spa_security_group" "web_sg" {
  name        = "web-security-group"
  description = "Security group for web applications"
  # Add security rules and other attributes
}

# Create an access policy
resource "spa_access_policy" "web_policy" {
  name        = "web-access-policy"
  description = "Access policy for web applications"
  # Configure access rules
}

# Outputs
output "application_id" {
  description = "ID of the created application"
  value       = spa_application.web_app.id
}

output "security_group_id" {
  description = "ID of the created security group"
  value       = spa_security_group.web_sg.id
}

output "access_policy_id" {
  description = "ID of the created access policy"
  value       = spa_access_policy.web_policy.id
}
```

### terraform.tfvars.example

```hcl
# Copy this file to terraform.tfvars and fill in your actual values
# DO NOT commit terraform.tfvars to version control

citrix_customer_id   = "your-customer-id-here"
citrix_client_id     = "your-client-id-here"
citrix_client_secret = "your-client-secret-here"
```

### .gitignore

```gitignore
# Terraform files
*.tfstate
*.tfstate.*
*.tfplan
*.tfplan.*
.terraform/
.terraform.lock.hcl

# Sensitive variable files
terraform.tfvars
*.auto.tfvars

# Override files
override.tf
override.tf.json
*_override.tf
*_override.tf.json

# Crash log files
crash.log
crash.*.log

# Exclude all .tfvars files, which are likely to contain sensitive data
*.tfvars
*.tfvars.json
```

## Alternative: Using Environment Variables

Instead of using `terraform.tfvars`, you can set environment variables:

```bash
export CITRIX_CUSTOMER_ID="your-customer-id-here"
export CITRIX_CLIENT_ID="your-client-id-here"
export CITRIX_CLIENT_SECRET="your-client-secret-here"

terraform init
terraform plan
terraform apply
```

## For Japan Region

If your Citrix Cloud account is in the Japan region, modify the provider configuration:

```hcl
provider "spa" {
  base_url      = "https://api.citrixcloud.jp"
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
}
```

## Security Considerations

1. **Never commit `terraform.tfvars`** to version control
2. **Use a secret management system** in production environments
3. **Rotate service principal secrets** regularly
4. **Use separate service principals** for different environments
5. **Monitor service principal usage** through Citrix Cloud audit logs

## Troubleshooting

### Common Issues

1. **Authentication Errors**

   - Verify client ID and secret are correct
   - Check service principal has necessary permissions
   - Ensure customer ID is correct

2. **Connection Errors**

   - Verify base URL is correct for your region
   - Check network connectivity to Citrix Cloud

3. **Permission Errors**
   - Ensure service principal has appropriate roles
   - Check if specific SPA permissions are required

### Debug Mode

To enable debug logging:

```bash
export TF_LOG=DEBUG
terraform plan
```

## Next Steps

1. Customize the configuration for your specific SPA resources
2. Add more complex resource configurations
3. Implement proper secret management for production use
4. Set up CI/CD integration with service principal authentication
