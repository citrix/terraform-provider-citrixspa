# Test configuration for debug logging
terraform {
  required_providers {
    spa = {
      source = "local/spa"
    }
  }
}

provider "spa" {
  base_url    = "https://api.cloud.com"
  customer_id = "test-customer"
  auth_token  = "test-token"
  debug       = true  # Enable debug logging
}

# Test resource for import
resource "spa_application" "debug_test" {
  name = "Debug Test Application"
  type = "web"
}
