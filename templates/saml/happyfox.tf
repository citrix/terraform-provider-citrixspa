# happyfox — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_happyfox_company_id_happyfox_com" {
  fqdn         = "<company-id>.happyfox.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "happyfox"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_happyfox_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "happyfox"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_happyfox" {
  name         = "happyfox"
  type         = "saas"
  state        = "complete"
  description  = "Online help desk software and web based support ticket system."
  url          = "https://<company-id>.happyfox.com/staff/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "happyfox"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-id>.happyfox.com/staff/saml/callback"
    audience          = "https://<company-id>.happyfox.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "User.FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "User.LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_happyfox_company_id_happyfox_com,
    citrixspa_routing_domain.rd_happyfox_customer_fqdn,
  ]
}
