# QAComplete — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_qacomplete_app_qacomplete_smartbear_com" {
  fqdn         = "app.qacomplete.smartbear.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "QAComplete"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_qacomplete_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "QAComplete"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_qacomplete" {
  name         = "QAComplete"
  type         = "saas"
  state        = "complete"
  description  = "Software test management tool."
  url          = "https://app.qacomplete.smartbear.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "QAComplete"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer_domain>.qacomplete.smartbear.com/sso/callback"
    audience          = "https://qacomplete.smartbear.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_qacomplete_app_qacomplete_smartbear_com,
    citrixspa_routing_domain.rd_qacomplete_customer_fqdn,
  ]
}
