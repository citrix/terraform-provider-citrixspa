# Runscope — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_runscope_www_runscope_com" {
  fqdn         = "www.runscope.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Runscope"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_runscope_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Runscope"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_runscope" {
  name         = "Runscope"
  type         = "saas"
  state        = "complete"
  description  = "Tool to create, manage and execute functional API tests and monitors."
  url          = "https://www.runscope.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Runscope"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.runscope.com/saml-signin"
    audience          = "https://www.runscope.com"
    relay_state       = "c67f4a1b-571e-4b15-8ff5-2b3cf2fa436d"
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
        name  = "email_address"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_runscope_www_runscope_com,
    citrixspa_routing_domain.rd_runscope_customer_fqdn,
  ]
}
