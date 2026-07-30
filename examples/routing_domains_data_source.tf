terraform {
  required_providers {
    citrixspa = {
      source = "citrix/citrixspa"
    }
  }
}

# Configure the SPA Provider
provider "citrixspa" {
  # Configuration can be provided via environment variables:
  # CITRIX_CUSTOMER_ID
  # CITRIX_AUTH_TOKEN (or CITRIX_CLIENT_ID + CITRIX_CLIENT_SECRET)
  # base_url = "https://api.cloud.com/accessSecurity"
}

# Data source for listing all routing domains
data "citrixspa_routing_domains" "all" {
  # Optional pagination parameters
  # offset = 0
  # limit = 100
}

# Output the routing domains
output "routing_domains" {
  value = data.citrixspa_routing_domains.all.routing_domains
}

output "routing_domains_count" {
  value = length(data.citrixspa_routing_domains.all.routing_domains)
}

output "routing_domains_total" {
  value = data.citrixspa_routing_domains.all.total
}
