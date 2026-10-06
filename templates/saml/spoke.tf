# Spoke — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_spoke_customer_domain_askspoke_com" {
  fqdn         = "<customer-domain>.askspoke.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Spoke"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_spoke_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Spoke"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_spoke" {
  name         = "Spoke"
  type         = "saas"
  state        = "complete"
  description  = "Service desk tool to file service tickets."
  url          = "https://<customer-domain>.askspoke.com/requests/inbox"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Spoke"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.askspoke.com/saml/callback"
    audience          = "https://askspoke.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_spoke_customer_domain_askspoke_com,
    citrixspa_routing_domain.rd_spoke_customer_fqdn,
  ]
}
