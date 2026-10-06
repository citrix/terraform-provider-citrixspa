# Feeder — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_feeder_feeder_co" {
  fqdn         = "feeder.co"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Feeder"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_feeder_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Feeder"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_feeder" {
  name         = "Feeder"
  type         = "saas"
  state        = "complete"
  description  = "Follow everything you care about. Faster. Easier. In a unique way."
  url          = "https://feeder.co/reader"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Feeder"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://feeder.co/-saml/<customer_id>/callback"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_feeder_feeder_co,
    citrixspa_routing_domain.rd_feeder_customer_fqdn,
  ]
}
