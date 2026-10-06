# Bugsnag — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bugsnag_app_bugsnag_com" {
  fqdn         = "app.bugsnag.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bugsnag"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bugsnag_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bugsnag"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bugsnag" {
  name         = "Bugsnag"
  type         = "saas"
  state        = "complete"
  description  = "Monitoring tool to manage application stability and report errors and diagnostic data."
  url          = "https://app.bugsnag.com/settings/<customer_id>/projects"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Bugsnag"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.bugsnag.com/user/sign_in/saml/<customer_id>/acs"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_bugsnag_app_bugsnag_com,
    citrixspa_routing_domain.rd_bugsnag_customer_fqdn,
  ]
}
