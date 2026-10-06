# HelloSign — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_hellosign_app_hellosign_com" {
  fqdn         = "app.hellosign.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HelloSign"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_hellosign_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HelloSign"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_hellosign" {
  name         = "HelloSign"
  type         = "saas"
  state        = "complete"
  description  = "Esigning interface to enable signing from anywhere, at any time, on any device."
  url          = "https://app.hellosign.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "HelloSign"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.hellosign.com/account/ssoLogIn"
    audience          = "https://app.hellosign.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_hellosign_app_hellosign_com,
    citrixspa_routing_domain.rd_hellosign_customer_fqdn,
  ]
}
