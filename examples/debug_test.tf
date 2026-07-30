# Test configuration for debug logging
terraform {
  required_providers {
    citrixspa = {
      source = "local/spa"
    }
  }
}

provider "citrixspa" {
  base_url    = "https://api.cloud.com"
  customer_id = "test-customer"
  auth_token  = "test-token"
  debug       = true  # Enable debug logging
}

# Test resource for import
resource "citrixspa_application" "debug_test" {
  name = "Debug Test Application"
  type = "web"
}
