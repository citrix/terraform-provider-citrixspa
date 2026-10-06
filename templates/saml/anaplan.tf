# Anaplan — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_anaplan_www_anaplan_com" {
  fqdn         = "www.anaplan.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Anaplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_anaplan_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Anaplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_anaplan" {
  name         = "Anaplan"
  type         = "saas"
  state        = "complete"
  description  = "Planning tool to help organizations with decision making by connecting data, people, and plans."
  url          = "https://www.anaplan.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Anaplan"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sdp.anaplan.com/frontdoor/saml/<customer_domain>"
    audience          = "https://sdp.anaplan.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_anaplan_www_anaplan_com,
    citrixspa_routing_domain.rd_anaplan_customer_fqdn,
  ]
}
