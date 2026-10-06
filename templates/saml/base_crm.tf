# Base CRM — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_base_crm_app_futuresimple_com" {
  fqdn         = "app.futuresimple.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Base CRM"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_base_crm_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Base CRM"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_base_crm" {
  name         = "Base CRM"
  type         = "saas"
  state        = "complete"
  description  = "Sales management tool to manage emails, phone calls, and notes."
  url          = "https://app.futuresimple.com/sales"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Base CRM"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://core.futuresimple.com/sso/saml/<customer_UUID>/consume"
    audience          = "https://core.futuresimple.com/sso/saml/<customer_UUID>/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_base_crm_app_futuresimple_com,
    citrixspa_routing_domain.rd_base_crm_customer_fqdn,
  ]
}
