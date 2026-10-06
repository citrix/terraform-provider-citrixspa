# Tallie — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_tallie_usetallie_com" {
  fqdn         = "usetallie.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Tallie"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_tallie_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Tallie"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_tallie" {
  name         = "Tallie"
  type         = "saas"
  state        = "complete"
  description  = "Tool to capture and upload receipts, generate expense reports, and customize expense details."
  url          = "https://usetallie.com/<Account-id>/Expense/Tile"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Tallie"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://usetallie.com/sso/samlauth/authresponse?providerId=<Customer-id>"
    audience          = "https://usetallie.com/sso"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_tallie_usetallie_com,
    citrixspa_routing_domain.rd_tallie_customer_fqdn,
  ]
}
