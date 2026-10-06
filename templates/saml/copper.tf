# Copper — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_copper_app_prosperworks_com" {
  fqdn         = "app.prosperworks.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Copper"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_copper_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Copper"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_copper" {
  name         = "Copper"
  type         = "saas"
  state        = "complete"
  description  = "Copper is a new kind of productivity crm that's designed to do all your busywork, so you can focus on building long-lasting business relationships."
  url          = "https://app.prosperworks.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Copper"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.prosperworks.com/saml/v2/companies/<customer_id>/saml/consume"
    audience          = "https://app.prosperworks.com/saml/v2/companies/<customer_id>/saml/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_copper_app_prosperworks_com,
    citrixspa_routing_domain.rd_copper_customer_fqdn,
  ]
}
