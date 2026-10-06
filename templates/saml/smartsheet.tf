# Smartsheet — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_smartsheet_app_smartsheet_com" {
  fqdn         = "app.smartsheet.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Smartsheet"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_smartsheet_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Smartsheet"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_smartsheet" {
  name         = "Smartsheet"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration tool to assign tasks, track project process, manage calendars, and share documents."
  url          = "https://app.smartsheet.com/b/home"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Smartsheet"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sso.smartsheet.com/Shibboleth.sso/SAML2/POST"
    audience          = "https://sso.smartsheet.com/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_smartsheet_app_smartsheet_com,
    citrixspa_routing_domain.rd_smartsheet_customer_fqdn,
  ]
}
