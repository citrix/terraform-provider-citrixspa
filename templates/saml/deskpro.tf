# DeskPro — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_deskpro_company_domain_deskpro_com" {
  fqdn         = "<company-domain>.deskpro.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "DeskPro"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_deskpro_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "DeskPro"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_deskpro" {
  name         = "DeskPro"
  type         = "saas"
  state        = "complete"
  description  = "Help desk tool to facilitate ticket management, customer self-help, and customer feedback."
  url          = "https://<company-domain>.deskpro.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "DeskPro"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.deskpro.com/agent/login/authenticate-callback/<int>"
    audience          = "https://<company-domain>.deskpro.com/saml/metadata/<int>.xml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_deskpro_company_domain_deskpro_com,
    citrixspa_routing_domain.rd_deskpro_customer_fqdn,
  ]
}
