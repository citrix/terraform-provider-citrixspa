# Signavio — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_signavio_app_au_signavio_com" {
  fqdn         = "app-au.signavio.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Signavio"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_signavio_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Signavio"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_signavio" {
  name         = "Signavio"
  type         = "saas"
  state        = "complete"
  description  = "A business process modelling tool."
  url          = "https://app-au.signavio.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Signavio"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app-au.signavio.com/intralink/saml2endpoint"
    audience          = "app-au.signavio.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_signavio_app_au_signavio_com,
    citrixspa_routing_domain.rd_signavio_customer_fqdn,
  ]
}
