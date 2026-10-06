# Float — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_float_customer_domain_float_com" {
  fqdn         = "<Customer-domain>.float.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Float"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_float_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Float"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_float" {
  name         = "Float"
  type         = "saas"
  state        = "complete"
  description  = "Resource planning tool for project scheduling and managing the teams' utilization."
  url          = "https://<Customer-domain>.float.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Float"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.float.com/sso/okta/?company=<Customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "name"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_float_customer_domain_float_com,
    citrixspa_routing_domain.rd_float_customer_fqdn,
  ]
}
