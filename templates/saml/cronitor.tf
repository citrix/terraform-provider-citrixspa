# Cronitor — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cronitor_cronitor_io" {
  fqdn         = "cronitor.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cronitor"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cronitor_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cronitor"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cronitor" {
  name         = "Cronitor"
  type         = "saas"
  state        = "complete"
  description  = "Monitoring tool for cron jobs."
  url          = "https://cronitor.io"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Cronitor"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://cronitor.io/auth/saml/acs/<Customer_id>"
    audience          = "https://cronitor.io/auth/saml/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_cronitor_cronitor_io,
    citrixspa_routing_domain.rd_cronitor_customer_fqdn,
  ]
}
