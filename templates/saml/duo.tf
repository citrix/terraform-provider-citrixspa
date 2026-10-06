# DUO — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_duo_customer_domain_duosecurity_com" {
  fqdn         = "<customer-domain>.duosecurity.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "DUO"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_duo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "DUO"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_duo" {
  name         = "DUO"
  type         = "saas"
  state        = "complete"
  description  = "Security tool to provide secure access to your applications."
  url          = "https://<customer-domain>.duosecurity.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "DUO"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.duosecurity.com/saml/<customer_id>/acs"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_duo_customer_domain_duosecurity_com,
    citrixspa_routing_domain.rd_duo_customer_fqdn,
  ]
}
