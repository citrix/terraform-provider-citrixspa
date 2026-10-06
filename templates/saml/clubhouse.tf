# Clubhouse — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_clubhouse_app_clubhouse_io" {
  fqdn         = "app.clubhouse.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Clubhouse"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_clubhouse_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Clubhouse"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_clubhouse" {
  name         = "Clubhouse"
  type         = "saas"
  state        = "complete"
  description  = "Project management tool for software development."
  url          = "https://app.clubhouse.io"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Clubhouse"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sso.clubhouse.io/login/saml2/<Customer_id>"
    audience          = "https://sso.clubhouse.io/saml2sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_clubhouse_app_clubhouse_io,
    citrixspa_routing_domain.rd_clubhouse_customer_fqdn,
  ]
}
