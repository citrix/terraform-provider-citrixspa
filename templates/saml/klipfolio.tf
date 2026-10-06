# Klipfolio — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_klipfolio_app_klipfolio_com" {
  fqdn         = "app.klipfolio.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Klipfolio"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_klipfolio_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Klipfolio"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_klipfolio" {
  name         = "Klipfolio"
  type         = "saas"
  state        = "complete"
  description  = "Online dashboard platform for building powerful real-time business dashboards for your team or your clients."
  url          = "https://app.klipfolio.com/tabs/index"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Klipfolio"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.klipfolio.com/SAML/consume"
    audience          = "https://app.klipfolio.com/SAML/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_klipfolio_app_klipfolio_com,
    citrixspa_routing_domain.rd_klipfolio_customer_fqdn,
  ]
}
