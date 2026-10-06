# Absorb — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_absorb_customer_domain_myabsorb_com_au" {
  fqdn         = "<customer-domain>.myabsorb.com.au"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Absorb"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_absorb_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Absorb"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_absorb" {
  name         = "Absorb"
  type         = "saas"
  state        = "complete"
  description  = "Learning management tool."
  url          = "https://<customer-domain>.myabsorb.com.au/#/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Absorb"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.myabsorb.com.au/api/rest/v2/authentication/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_absorb_customer_domain_myabsorb_com_au,
    citrixspa_routing_domain.rd_absorb_customer_fqdn,
  ]
}
