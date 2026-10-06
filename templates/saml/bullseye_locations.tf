# Bullseye Locations — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bullseye_locations_app_bullseyelocations_com" {
  fqdn         = "app.bullseyelocations.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bullseye Locations"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bullseye_locations_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bullseye Locations"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bullseye_locations" {
  name         = "Bullseye Locations"
  type         = "saas"
  state        = "complete"
  description  = "Store locator tool to locate a store or dealer on a device."
  url          = "https://app.bullseyelocations.com/Admin/Dashboard.aspx"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Bullseye Locations"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.bullseyelocations.com/Saml/AssertionConsumerService.aspx"
    audience          = "https://bullseyelocations.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_bullseye_locations_app_bullseyelocations_com,
    citrixspa_routing_domain.rd_bullseye_locations_customer_fqdn,
  ]
}
