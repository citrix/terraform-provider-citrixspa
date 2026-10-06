# ezeep — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ezeep_portal_ezeep_com" {
  fqdn         = "portal.ezeep.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ezeep"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ezeep_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ezeep"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ezeep" {
  name         = "ezeep"
  type         = "saas"
  state        = "complete"
  description  = "Print infrastructure management tool to print from any device, any location to any printer in the Cloud."
  url          = "https://portal.ezeep.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "ezeep"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://accounts.ezeep.com/auth/saml/<customer_id>/?acs"
    audience          = "https://accounts.ezeep.com/auth/saml/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_ezeep_portal_ezeep_com,
    citrixspa_routing_domain.rd_ezeep_customer_fqdn,
  ]
}
