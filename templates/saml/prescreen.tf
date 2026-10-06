# Prescreen — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_prescreen_customer_domain_prescreenapp_io" {
  fqdn         = "<customer-domain>.prescreenapp.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Prescreen"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_prescreen_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Prescreen"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_prescreen" {
  name         = "Prescreen"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based applicant tracking system to publish job vacancies online and offline."
  url          = "https://<customer-domain>.prescreenapp.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Prescreen"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.prescreenapp.io/saml/acs"
    audience          = "https://prescreenapp.io/saml/sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_prescreen_customer_domain_prescreenapp_io,
    citrixspa_routing_domain.rd_prescreen_customer_fqdn,
  ]
}
