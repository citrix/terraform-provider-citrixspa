# Cloudability — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cloudability_app_cloudability_com" {
  fqdn         = "app.cloudability.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cloudability"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cloudability_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cloudability"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cloudability" {
  name         = "Cloudability"
  type         = "saas"
  state        = "complete"
  description  = "Cloud Cost Management, Visibility and Optimization."
  url          = "https://app.cloudability.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Cloudability"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.cloudability.com/user/auth/saml/callback?idp=<customer_id>"
    audience          = "https://saml.cloudability.com/sp"
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
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_cloudability_app_cloudability_com,
    citrixspa_routing_domain.rd_cloudability_customer_fqdn,
  ]
}
