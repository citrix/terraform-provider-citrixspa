# Zillable — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zillable_customer_domain_zillable_com" {
  fqdn         = "<customer-domain>.zillable.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zillable"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zillable_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zillable"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zillable" {
  name         = "Zillable"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration platform with communication capabilities."
  url          = "https://<customer-domain>.zillable.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zillable"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.zillable.com/saml/SSO/alias/<customer_domain>"
    audience          = "https://<customer-domain>.zillable.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_zillable_customer_domain_zillable_com,
    citrixspa_routing_domain.rd_zillable_customer_fqdn,
  ]
}
