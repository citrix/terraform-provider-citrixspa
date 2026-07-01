terraform {
  required_providers {
    spa = {
      source  = "registry.terraform.io/citrix/spa"
      version = "0.1.0"
    }
  }
}

provider "spa" {
  # Configuration options
  base_url    = var.base_url
  customer_id = var.citrix_customer_id
  
  # Authentication Method 1: Direct token authentication
  auth_token = var.citrix_auth_token
  
  # Authentication Method 2: Service Principal authentication
  # Use these instead of auth_token for OAuth2 flow
  # client_id     = var.citrix_client_id
  # client_secret = var.citrix_client_secret
}

variable "base_url" {
  description = "SPA API Base URL"
  type        = string
  default     = "https://api.cloud.com/accessSecurity"
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_auth_token" {
  description = "Citrix Cloud Auth Token (for direct authentication)"
  type        = string
  sensitive   = true
  default     = null
}

variable "citrix_client_id" {
  description = "Citrix Cloud Service Principal Client ID (for OAuth2 authentication)"
  type        = string
  sensitive   = true
  default     = null
}

variable "citrix_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret (for OAuth2 authentication)"
  type        = string
  sensitive   = true
  default     = null
}

# Example: Data source to test connectivity
# GET the first 2 applications to verify the provider can connect and authenticate successfully
data "spa_applications" "test_apps" {
  offset = 0
  limit  = 2
}

output "apps" {
  value = data.spa_applications.test_apps
}

# Example: Create a simple resource for testing
# resource "spa_application" "test_app" {
#   name = "Test Application - Local Provider"
#   type = "web"
#   # Add other required attributes based on your resource schema
# }

# # Output to verify the test
# output "test_app_id" {
#   value = spa_application.test_app.id
# }

# output "provider_test_status" {
#   value = "Local provider successfully loaded and configured"
# }
