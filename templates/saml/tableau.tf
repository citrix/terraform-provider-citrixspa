# Tableau — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_tableau_sso_online_tableau_com" {
  fqdn         = "sso.online.tableau.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Tableau"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_tableau_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Tableau"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_tableau" {
  name         = "Tableau"
  type         = "saas"
  state        = "complete"
  description  = "Tool to create interactive data visualization."
  url          = "https://sso.online.tableau.com/public/idp/SSO"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Tableau"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sso.online.tableau.com/public/sp/SSO?alias=<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_tableau_sso_online_tableau_com,
    citrixspa_routing_domain.rd_tableau_customer_fqdn,
  ]
}
