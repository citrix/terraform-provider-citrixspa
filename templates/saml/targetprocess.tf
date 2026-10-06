# Targetprocess — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_targetprocess_customer_domain_tpondemand_com" {
  fqdn         = "<customer-domain>.tpondemand.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Targetprocess"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_targetprocess_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Targetprocess"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_targetprocess" {
  name         = "Targetprocess"
  type         = "saas"
  state        = "complete"
  description  = "Agile project management software to Scrum, Kanban, SAFe and so on."
  url          = "https://<customer-domain>.tpondemand.com/RestUI/Board.aspx#page=profile"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Targetprocess"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.tpondemand.com/api/sso/saml2"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_targetprocess_customer_domain_tpondemand_com,
    citrixspa_routing_domain.rd_targetprocess_customer_fqdn,
  ]
}
