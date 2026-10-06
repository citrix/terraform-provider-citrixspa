# UserEcho — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_userecho_customer_domain_userecho_com" {
  fqdn         = "<customer-domain>.userecho.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "UserEcho"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_userecho_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "UserEcho"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_userecho" {
  name         = "UserEcho"
  type         = "saas"
  state        = "complete"
  description  = "Community forum tool that helps businesses manage customer feedback."
  url          = "https://<customer-domain>.userecho.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "UserEcho"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.userecho.com/saml/acs/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "First name"
        value  = "aaa.user.attribute(\"givenName\")"
        format = "unspecified"
      },
      {
        name   = "Last name"
        value  = "aaa.user.attribute(\"sn\")"
        format = "unspecified"
      },
      {
        name   = "Email"
        value  = "ns_user_email"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_userecho_customer_domain_userecho_com,
    citrixspa_routing_domain.rd_userecho_customer_fqdn,
  ]
}
