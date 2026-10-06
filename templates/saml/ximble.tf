# Ximble — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ximble_id_ximble_com" {
  fqdn         = "id.ximble.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ximble"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ximble_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ximble"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ximble" {
  name         = "Ximble"
  type         = "saas"
  state        = "complete"
  description  = "Tool for employee scheduling and time tracking."
  url          = "https://id.ximble.com/identity/AuthServices/Acs"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Ximble"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://id.ximble.com/identity/AuthServices/Acs"
    audience          = "https://id.ximble.com/identity/AuthServices"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_ximble_id_ximble_com,
    citrixspa_routing_domain.rd_ximble_customer_fqdn,
  ]
}
