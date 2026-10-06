# Azendoo — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_azendoo_app_azendoo_com" {
  fqdn         = "app.azendoo.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Azendoo"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_azendoo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Azendoo"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_azendoo" {
  name         = "Azendoo"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration tool for teams to converse and collaborate."
  url          = "https://app.azendoo.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Azendoo"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.azendoo.com/authorizations/callbacks/saml"
    audience          = "https://app.azendoo.com/api/organizations/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_azendoo_app_azendoo_com,
    citrixspa_routing_domain.rd_azendoo_customer_fqdn,
  ]
}
