# Meisterplan — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_meisterplan_eu_meisterplan_com" {
  fqdn         = "eu.meisterplan.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Meisterplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_meisterplan_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Meisterplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_meisterplan" {
  name         = "Meisterplan"
  type         = "saas"
  state        = "complete"
  description  = "The cloud PPM software Meisterplan uses Lean PPM principles, so designing project portfolios that really work is fast and easy."
  url          = "https://eu.meisterplan.com/<customer_domain>#/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Meisterplan"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.eu.meisterplan.com/api/public/v1/saml/loginResponse"
    audience          = "meisterplan.com/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_meisterplan_eu_meisterplan_com,
    citrixspa_routing_domain.rd_meisterplan_customer_fqdn,
  ]
}
