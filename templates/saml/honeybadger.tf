# Honeybadger — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_honeybadger_app_honeybadger_io" {
  fqdn         = "app.honeybadger.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Honeybadger"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_honeybadger_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Honeybadger"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_honeybadger" {
  name         = "Honeybadger"
  type         = "saas"
  state        = "complete"
  description  = "Exception, uptime, and check-in monitoring system in a single platform."
  url          = "https://app.honeybadger.io"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Honeybadger"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.honeybadger.io/saml/<Customer_id>"
    audience          = "https://saml.honeybadger.io"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_honeybadger_app_honeybadger_io,
    citrixspa_routing_domain.rd_honeybadger_customer_fqdn,
  ]
}
