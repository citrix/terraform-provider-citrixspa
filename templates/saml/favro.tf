# Favro — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_favro_favro_com" {
  fqdn         = "favro.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Favro"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_favro_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Favro"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_favro" {
  name         = "Favro"
  type         = "saas"
  state        = "complete"
  description  = "Planning and collaboration tool for organizational flow."
  url          = "https://favro.com/organization/<customer_id>/blank"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Favro"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://favro.com/saml/assert"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_favro_favro_com,
    citrixspa_routing_domain.rd_favro_customer_fqdn,
  ]
}
