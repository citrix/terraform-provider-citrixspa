# Example Terraform configuration for SPA Provider Import

terraform {
  required_providers {
    spa = {
      source = "local/spa"
    }
  }
}

# Provider configuration
provider "spa" {
  # Option 1: Using direct authentication token
  base_url    = "https://your-spa-instance.com"
  customer_id = "your-customer-id"
  auth_token  = "your-auth-token"
  
  # Option 2: Using service principal (OAuth2)
  # base_url      = "https://your-spa-instance.com"
  # customer_id   = "your-customer-id"
  # client_id     = "your-client-id"
  # client_secret = "your-client-secret"
}

# Example resource configurations for import
# After importing, these will be populated with actual values

# Applications
resource "spa_application" "web_app" {
  # Will be populated after import
  # Import with: terraform import spa_application.web_app <application_id>
}

resource "spa_application" "api_app" {
  # Will be populated after import
  # Import with: terraform import spa_application.api_app <application_id>
}

# Access Policies
resource "spa_access_policy" "default_policy" {
  # Will be populated after import
  # Import with: terraform import spa_access_policy.default_policy <policy_id>
}

# Security Groups
resource "spa_security_group" "web_sg" {
  # Will be populated after import
  # Import with: terraform import spa_security_group.web_sg <security_group_id>
}

# Routing Domains
resource "spa_routing_domain" "main_domain" {
  # Will be populated after import
  # Import with: terraform import spa_routing_domain.main_domain <domain_id>
}

# Certificates
resource "spa_certificate" "ssl_cert" {
  # Will be populated after import
  # Import with: terraform import spa_certificate.ssl_cert <certificate_id>
}

# Browser Mode (singleton resource)
resource "spa_browser_mode" "browser_config" {
  # Will be populated after import
  # Import with: terraform import spa_browser_mode.browser_config browser_mode
}

# Terminate Machine Access
resource "spa_terminate_machine_access" "expired_session" {
  # Will be populated after import
  # Import with: terraform import spa_terminate_machine_access.expired_session <termination_id>
}

# Data sources for discovery (optional)
data "spa_application" "existing_apps" {
  # Use this to discover existing applications
  offset = 0
  limit  = 100
}

# Output discovered resource IDs
output "discovered_application_ids" {
  value       = data.spa_application.existing_apps.applications[*].id
  description = "IDs of all discovered applications"
}

output "discovered_application_names" {
  value       = data.spa_application.existing_apps.applications[*].name
  description = "Names of all discovered applications"
}
