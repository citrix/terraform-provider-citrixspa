# OneDesk — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_onedesk_your_organization_onedesk_com" {
  fqdn         = "<your-organization>.onedesk.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "OneDesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_onedesk_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "OneDesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_onedesk" {
  name         = "OneDesk"
  type         = "saas"
  state        = "complete"
  description  = "Project management and helpdesk software to connect with and support your customers."
  url          = "https://<your-organization>.onedesk.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "OneDesk"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.onedesk.com/sso/saml/SSO/alias/onedesk.com_<your-organization>"
    audience          = "onedesk.com_<your-organization>"
    sign_assertion    = "RESPONSE"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "EmailAddress"
        value  = "ns_user_email"
        format = "unspecified"
      },
      {
        name   = "FirstName"
        value  = "ns_user_name"
        format = "unspecified"
      },
      {
        name   = "LastName"
        value  = "aaa.USER.ATTRIBUTE(\"sn\")"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_onedesk_your_organization_onedesk_com,
    citrixspa_routing_domain.rd_onedesk_customer_fqdn,
  ]
}
