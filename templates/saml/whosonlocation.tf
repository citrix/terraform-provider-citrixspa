# WhosOnLocation — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_whosonlocation_us_whosonlocation_com" {
  fqdn         = "us.whosonlocation.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WhosOnLocation"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_whosonlocation_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WhosOnLocation"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_whosonlocation" {
  name         = "WhosOnLocation"
  type         = "saas"
  state        = "complete"
  description  = "Tool to track the flow of people through sites and zones."
  url          = "https://us.whosonlocation.com/home"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "WhosOnLocation"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.whosonlocation.com/saml/acs/<customer_id>"
    audience          = "https://login.whosonlocation.com/saml/metadata/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_whosonlocation_us_whosonlocation_com,
    citrixspa_routing_domain.rd_whosonlocation_customer_fqdn,
  ]
}
