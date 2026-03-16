terraform {
  required_providers {
    spa = {
      source = "code.citrite.net/csvc/spa-sdk/terraform-provider-spa"
    }
  }
}

# Configure the SPA Provider
provider "spa" {
  # Configuration can be provided via environment variables:
  # CITRIX_CUSTOMER_ID
  # CITRIX_AUTH_TOKEN (or CITRIX_CLIENT_ID + CITRIX_CLIENT_SECRET)
  # base_url = "https://api.cloud.com/accessSecurity"
}

# Data source for listing all routing domains
data "spa_routing_domains" "all" {
  # Optional pagination parameters
  # offset = 0
  # limit = 100
}

# Output the routing domains
output "routing_domains" {
  value = data.spa_routing_domains.all.routing_domains
}

output "routing_domains_count" {
  value = data.spa_routing_domains.all.count
}

output "routing_domains_total" {
  value = data.spa_routing_domains.all.total
}
