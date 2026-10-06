# Proxyclick — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_proxyclick_app_proxyclick_com" {
  fqdn         = "app.proxyclick.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Proxyclick"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_proxyclick_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Proxyclick"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_proxyclick" {
  name         = "Proxyclick"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based visitor management solution to manage visitors, build their brand image, and ensure the security."
  url          = "https://app.proxyclick.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Proxyclick"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://saml.proxyclick.com/consume/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_proxyclick_app_proxyclick_com,
    citrixspa_routing_domain.rd_proxyclick_customer_fqdn,
  ]
}
