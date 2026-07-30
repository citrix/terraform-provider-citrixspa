# Example Terraform configuration for SPA Provider Import

terraform {
  required_providers {
    citrixspa = {
      source = "local/spa"
    }
  }
}

# Provider configuration
provider "citrixspa" {
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
resource "citrixspa_application" "web_app" {
  # Will be populated after import
  # Import with: terraform import citrixspa_application.web_app <application_id>
}

resource "citrixspa_application" "api_app" {
  # Will be populated after import
  # Import with: terraform import citrixspa_application.api_app <application_id>
}

# Access Policies
resource "citrixspa_access_policy" "default_policy" {
  # Will be populated after import
  # Import with: terraform import citrixspa_access_policy.default_policy <policy_id>
}

# Security Groups
resource "citrixspa_security_group" "web_sg" {
  # Will be populated after import
  # Import with: terraform import citrixspa_security_group.web_sg <security_group_id>
}

# Routing Domains
resource "citrixspa_routing_domain" "main_domain" {
  # Will be populated after import
  # Import with: terraform import citrixspa_routing_domain.main_domain <domain_id>
}

# Certificates
resource "citrixspa_certificate" "ssl_cert" {
  # Will be populated after import
  # Import with: terraform import citrixspa_certificate.ssl_cert <certificate_id>
}

# Browser Mode (singleton resource)
resource "citrixspa_browser_mode" "browser_config" {
  # Will be populated after import
  # Import with: terraform import citrixspa_browser_mode.browser_config browser_mode
}

# Terminate Machine Access
resource "citrixspa_terminate_machine_access" "expired_session" {
  # Will be populated after import
  # Import with: terraform import citrixspa_terminate_machine_access.expired_session <termination_id>
}

# Data sources for discovery (optional)
data "citrixspa_application" "existing_apps" {
  # Use this to discover existing applications
  offset = 0
  limit  = 100
}

# Output discovered resource IDs
output "discovered_application_ids" {
  value       = data.citrixspa_application.existing_apps.applications[*].id
  description = "IDs of all discovered applications"
}

output "discovered_application_names" {
  value       = data.citrixspa_application.existing_apps.applications[*].name
  description = "Names of all discovered applications"
}
