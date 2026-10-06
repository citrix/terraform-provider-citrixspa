# GetGuru — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_getguru_app_getguru_com" {
  fqdn         = "app.getguru.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GetGuru"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_getguru_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GetGuru"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_getguru" {
  name         = "GetGuru"
  type         = "saas"
  state        = "complete"
  description  = "Revenue empowerment network to empower your revenue teams."
  url          = "https://app.getguru.com/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "GetGuru"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.getguru.com/samlsso/<customer_id>"
    audience          = "getguru.com/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_getguru_app_getguru_com,
    citrixspa_routing_domain.rd_getguru_customer_fqdn,
  ]
}
