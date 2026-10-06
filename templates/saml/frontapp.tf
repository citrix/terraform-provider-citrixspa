# FrontApp — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_frontapp_customer_domain_frontapp_com" {
  fqdn         = "<customer-domain>.frontapp.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "FrontApp"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_frontapp_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "FrontApp"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_frontapp" {
  name         = "FrontApp"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration tool to manage all conversations in one place."
  url          = "https://<customer-domain>.frontapp.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "FrontApp"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.frontapp.com/sso/saml/callback"
    audience          = "https://<customer-domain>.frontapp.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_frontapp_customer_domain_frontapp_com,
    citrixspa_routing_domain.rd_frontapp_customer_fqdn,
  ]
}
