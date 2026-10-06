# GoToMeeting — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_gotomeeting_organization_logmeininc_com" {
  fqdn         = "organization.logmeininc.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GoToMeeting"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_gotomeeting_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GoToMeeting"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_gotomeeting" {
  name         = "GoToMeeting"
  type         = "saas"
  state        = "complete"
  description  = "Online meeting software with HD Video Conferencing capabilities."
  url          = "https://organization.logmeininc.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "GoToMeeting"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://authentication.logmeininc.com/saml/acs"
    audience          = "https://authentication.logmeininc.com/saml/sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_gotomeeting_organization_logmeininc_com,
    citrixspa_routing_domain.rd_gotomeeting_customer_fqdn,
  ]
}
