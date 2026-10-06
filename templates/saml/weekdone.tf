# Weekdone — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_weekdone_weekdone_com" {
  fqdn         = "weekdone.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Weekdone"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_weekdone_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Weekdone"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_weekdone" {
  name         = "Weekdone"
  type         = "saas"
  state        = "complete"
  description  = "Tool to create managers' dashboard and team management service for companies."
  url          = "https://weekdone.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Weekdone"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://weekdone.com/a/<customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_weekdone_weekdone_com,
    citrixspa_routing_domain.rd_weekdone_customer_fqdn,
  ]
}
