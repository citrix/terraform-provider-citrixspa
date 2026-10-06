# Nmbrs — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_nmbrs_your_organisation_nmbrs_nl" {
  fqdn         = "<your-organisation>.nmbrs.nl"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Nmbrs"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_nmbrs_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Nmbrs"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_nmbrs" {
  name         = "Nmbrs"
  type         = "saas"
  state        = "complete"
  description  = "Cloud HR and payroll software for businesses."
  url          = "https://<your-organisation>.nmbrs.nl/applications/Common/Login.aspx"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Nmbrs"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organisation>.nmbrs.nl/applications/common/externalactions.aspx?login=samlresponse"
    relay_state       = "https://<your-organisation>.nmbrs.nl"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_nmbrs_your_organisation_nmbrs_nl,
    citrixspa_routing_domain.rd_nmbrs_customer_fqdn,
  ]
}
