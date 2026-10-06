# Pingboard — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_pingboard_company_name_pingboard_com" {
  fqdn         = "<company-name>.pingboard.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pingboard"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_pingboard_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pingboard"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_pingboard" {
  name         = "Pingboard"
  type         = "saas"
  state        = "complete"
  description  = "Tool to build organization charts for organizing teams and workforce planning."
  url          = "https://<company-name>.pingboard.com/users/directory"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Pingboard"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-name>.pingboard.com/auth/saml/consume"
    audience          = "http://app.pingboard.com/sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_pingboard_company_name_pingboard_com,
    citrixspa_routing_domain.rd_pingboard_customer_fqdn,
  ]
}
