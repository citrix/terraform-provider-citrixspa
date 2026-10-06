# WORKPATH — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_workpath_www_workpath_com" {
  fqdn         = "www.workpath.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WORKPATH"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workpath_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WORKPATH"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_workpath" {
  name         = "WORKPATH"
  type         = "saas"
  state        = "complete"
  description  = "Tool to manage goals and performance of the organization."
  url          = "https://www.workpath.com/en/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "WORKPATH"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.workpath.com/v1/saml/assert/<your-account-url>"
    audience          = "https://api.workpath.com/v1/saml/metadata/<your-account-url>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "first_name"
        value  = "ns_user_name"
        format = "unspecified"
      },
      {
        name   = "last_name"
        value  = "aaa.USER.ATTRIBUTE(\"sn\")"
        format = "unspecified"
      },
      {
        name   = "email"
        value  = "ns_user_email"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_workpath_www_workpath_com,
    citrixspa_routing_domain.rd_workpath_customer_fqdn,
  ]
}
