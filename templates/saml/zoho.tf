# Zoho — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zoho_www_zoho_com" {
  fqdn         = "www.zoho.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zoho"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zoho_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zoho"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zoho" {
  name         = "Zoho"
  type         = "saas"
  state        = "complete"
  description  = "Business application suite."
  url          = "https://www.zoho.com/accounts/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zoho"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://accounts.zoho.com/samlresponse/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_zoho_www_zoho_com,
    citrixspa_routing_domain.rd_zoho_customer_fqdn,
  ]
}
