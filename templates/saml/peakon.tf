# Peakon — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_peakon_app_peakon_com" {
  fqdn         = "app.peakon.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Peakon"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_peakon_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Peakon"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_peakon" {
  name         = "Peakon"
  type         = "saas"
  state        = "complete"
  description  = "Tool to measure and improve employee engagement."
  url          = "https://app.peakon.com/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Peakon"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.peakon.com/saml/<customer_id>/assert"
    audience          = "https://app.peakon.com/saml/<customer_id>/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_peakon_app_peakon_com,
    citrixspa_routing_domain.rd_peakon_customer_fqdn,
  ]
}
