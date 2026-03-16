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
  
  # Service Principal authentication
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
}

variable "base_url" {
  description = "SPA API Base URL"
  type        = string
  default     = "https://api.cloud.com"
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Citrix Cloud Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret"
  type        = string
  sensitive   = true
}

# Example: Create a simple resource for testing
resource "spa_application" "test_app" {
  name = "Test Application - Service Principal Auth"
  type = "web"
  # Add other required attributes based on your resource schema
}

# Output to verify the test
output "test_app_id" {
  value = spa_application.test_app.id
}

output "provider_test_status" {
  value = "Service Principal authentication successful"
}

output "authentication_method" {
  value = "OAuth2 Service Principal"
}
