# Rollbar — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_rollbar_rollbar_com" {
  fqdn         = "rollbar.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Rollbar"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_rollbar_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Rollbar"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_rollbar" {
  name         = "Rollbar"
  type         = "saas"
  state        = "complete"
  description  = "Real-time error alerting and debugging tools for developers."
  url          = "https://rollbar.com/netscaler/saml/login/other/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Rollbar"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://rollbar.com/netscaler/saml/sso/other/"
    audience          = "https://saml.rollbar.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_rollbar_rollbar_com,
    citrixspa_routing_domain.rd_rollbar_customer_fqdn,
  ]
}
