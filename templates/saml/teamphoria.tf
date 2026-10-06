# teamphoria — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_teamphoria_your_organisation_teamphoria_com" {
  fqdn         = "<your-organisation>.teamphoria.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "teamphoria"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_teamphoria_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "teamphoria"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_teamphoria" {
  name         = "teamphoria"
  type         = "saas"
  state        = "complete"
  description  = "Software to provide real-time employee engagement metrics, employee reviews, and recognition."
  url          = "https://<your-organisation>.teamphoria.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "teamphoria"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.teamphoria.com/auth/saml/login/callback/<your-orgid>"
    audience          = "https://app.teamphoria.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_teamphoria_your_organisation_teamphoria_com,
    citrixspa_routing_domain.rd_teamphoria_customer_fqdn,
  ]
}
