# FOSSA — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_fossa_app_fossa_io" {
  fqdn         = "app.fossa.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "FOSSA"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_fossa_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "FOSSA"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_fossa" {
  name         = "FOSSA"
  type         = "saas"
  state        = "complete"
  description  = "FOSSA - Open Source Management for Modern Development. Automated open source license scanning and vulnerability management tools built natively into CI CD."
  url          = "https://app.fossa.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "FOSSA"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.fossa.io/account/saml/<customer_id>/callback"
    audience          = "https://app.fossa.io/account/saml/<customer_id>/<organization_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_fossa_app_fossa_io,
    citrixspa_routing_domain.rd_fossa_customer_fqdn,
  ]
}
