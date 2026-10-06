# EZRentOut — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ezrentout_company_name_ezrentout_com" {
  fqdn         = "<company-name>.ezrentout.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EZRentOut"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ezrentout_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EZRentOut"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ezrentout" {
  name         = "EZRentOut"
  type         = "saas"
  state        = "complete"
  description  = "Equipment rental tool to track equipment quality and availability."
  url          = "https://<company-name>.ezrentout.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "EZRentOut"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-name>.ezrentout.com/users/auth/saml/callback"
    audience          = "https://www.ezrentout.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_ezrentout_company_name_ezrentout_com,
    citrixspa_routing_domain.rd_ezrentout_customer_fqdn,
  ]
}
