# Bananatag — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bananatag_app_bananatag_com" {
  fqdn         = "app.bananatag.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bananatag"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bananatag_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bananatag"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bananatag" {
  name         = "Bananatag"
  type         = "saas"
  state        = "complete"
  description  = "Tool to track and schedule emails, track files and create email templates"
  url          = "https://app.bananatag.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Bananatag"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login-service.bananatag.com/sso/acs/<customer_domain>.com"
    audience          = "urn:bananatag"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_bananatag_app_bananatag_com,
    citrixspa_routing_domain.rd_bananatag_customer_fqdn,
  ]
}
