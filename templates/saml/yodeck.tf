# Yodeck — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_yodeck_app_yodeck_com" {
  fqdn         = "app.yodeck.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Yodeck"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_yodeck_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Yodeck"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_yodeck" {
  name         = "Yodeck"
  type         = "saas"
  state        = "complete"
  description  = "Tool to manage screens remotely, through the web or mobile."
  url          = "https://app.yodeck.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Yodeck"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.yodeck.com/api/v1/user/acs/"
    audience          = "https://app.yodeck.com/api/v1/account/metadata/"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_yodeck_app_yodeck_com,
    citrixspa_routing_domain.rd_yodeck_customer_fqdn,
  ]
}
