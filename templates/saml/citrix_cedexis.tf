# Citrix Cedexis — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_citrix_cedexis_portal_cedexis_com" {
  fqdn         = "portal.cedexis.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Citrix Cedexis"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_citrix_cedexis_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Citrix Cedexis"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_citrix_cedexis" {
  name         = "Citrix Cedexis"
  type         = "saas"
  state        = "complete"
  description  = "Traffic management tool for large websites to leverage multivendor sourcing of data centers, cloud providers, and content delivery networks."
  url          = "https://portal.cedexis.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Citrix Cedexis"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://portal.cedexis.com/api/v2/sso/saml.json/<customer_id>/consume"
    audience          = "https://portal.cedexis.com/api/v2/sso/saml.json/<customer_id>/sp/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_citrix_cedexis_portal_cedexis_com,
    citrixspa_routing_domain.rd_citrix_cedexis_customer_fqdn,
  ]
}
