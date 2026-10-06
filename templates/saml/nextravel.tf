# NexTravel — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_nextravel_www_nextravel_com" {
  fqdn         = "www.nextravel.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "NexTravel"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_nextravel_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "NexTravel"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_nextravel" {
  name         = "NexTravel"
  type         = "saas"
  state        = "complete"
  description  = "Save time and money with NexTravel's corporate travel management software - the simplest travel booking and management system."
  url          = "https://www.nextravel.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "NexTravel"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.nextravel.com/saml/<customer_id>/acs"
    audience          = "https://www.nextravel.com/saml/<customer_id>/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_nextravel_www_nextravel_com,
    citrixspa_routing_domain.rd_nextravel_customer_fqdn,
  ]
}
