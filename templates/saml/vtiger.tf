# Vtiger — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_vtiger_customer_domain_od2_vtiger_com" {
  fqdn         = "<Customer-domain>.od2.vtiger.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Vtiger"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_vtiger_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Vtiger"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_vtiger" {
  name         = "Vtiger"
  type         = "saas"
  state        = "complete"
  description  = "Customer Relationship Management Software"
  url          = "https://<Customer-domain>.od2.vtiger.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Vtiger"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.od2.vtiger.com/sso/saml/?acs"
    audience          = "https://<Customer-domain>.od2.vtiger.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_vtiger_customer_domain_od2_vtiger_com,
    citrixspa_routing_domain.rd_vtiger_customer_fqdn,
  ]
}
