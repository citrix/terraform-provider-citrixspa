# Helpjuice — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_helpjuice_helpjuice_com" {
  fqdn         = "helpjuice.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Helpjuice"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_helpjuice_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Helpjuice"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_helpjuice" {
  name         = "Helpjuice"
  type         = "saas"
  state        = "complete"
  description  = "Knowledge management solution to create and maintain knowledge bases."
  url          = "https://helpjuice.com/users/sign_in"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Helpjuice"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://helpjuice.com/saml/okta"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "firstName"
        value  = "ns_user_name"
        format = "basic"
      },
      {
        name   = "lastName"
        value  = "aaa.USER.ATTRIBUTE(\"sn\")"
        format = "basic"
      },
      {
        name   = "email"
        value  = "ns_user_email"
        format = "basic"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_helpjuice_helpjuice_com,
    citrixspa_routing_domain.rd_helpjuice_customer_fqdn,
  ]
}
