# ServiceDesk Plus — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_servicedesk_plus_sdpondemand_manageengine_com" {
  fqdn         = "sdpondemand.manageengine.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ServiceDesk Plus"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_servicedesk_plus_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ServiceDesk Plus"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_servicedesk_plus" {
  name         = "ServiceDesk Plus"
  type         = "saas"
  state        = "complete"
  description  = "Tool for IT service desk."
  url          = "https://sdpondemand.manageengine.com/app/itdesk/HomePage.do"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "ServiceDesk Plus"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://accounts.zoho.com/samlresponse/<your-userid>"
    audience          = "zoho.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_servicedesk_plus_sdpondemand_manageengine_com,
    citrixspa_routing_domain.rd_servicedesk_plus_customer_fqdn,
  ]
}
