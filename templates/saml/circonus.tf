# Circonus — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_circonus_login_circonus_com" {
  fqdn         = "login.circonus.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Circonus"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_circonus_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Circonus"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_circonus" {
  name         = "Circonus"
  type         = "saas"
  state        = "complete"
  description  = "Data analytics and monitoring tool to deliver alerts, graphs, dashboards and machine-learning intelligence."
  url          = "https://login.circonus.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Circonus"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.circonus.com/login/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "ns_user_name"
      },
      {
        name  = "LastName"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_circonus_login_circonus_com,
    citrixspa_routing_domain.rd_circonus_customer_fqdn,
  ]
}
