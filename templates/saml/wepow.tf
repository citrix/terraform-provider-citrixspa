# Wepow — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_wepow_your_organization_wepowapp_com" {
  fqdn         = "<your-organization>.wepowapp.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Wepow"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_wepow_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Wepow"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_wepow" {
  name         = "Wepow"
  type         = "saas"
  state        = "complete"
  description  = "Tool to connect recruiters, job candidates, and employers through mobile and video interviewing solution."
  url          = "https://<your-organization>.wepowapp.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Wepow"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.wepowapp.com/sso/saml/consume"
    audience          = "https://wepowapp.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "first_name"
        value = "aaa.USER.ATTRIBUTE(\"givenName\")"
      },
      {
        name  = "last_name"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "is_owner"
        value = "\"false\""
      },
      {
        name  = "wepow_role_1"
        value = "\"Marketing\""
      },
      {
        name  = "wepow_team_1"
        value = "\"admin\""
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_wepow_your_organization_wepowapp_com,
    citrixspa_routing_domain.rd_wepow_customer_fqdn,
  ]
}
