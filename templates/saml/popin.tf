# POPin — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_popin_app_popinnow_com" {
  fqdn         = "app.popinnow.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "POPin"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_popin_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "POPin"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_popin" {
  name         = "POPin"
  type         = "saas"
  state        = "complete"
  description  = "POPin Knowledge-Sourcing for Ideas, Solutions and Buy-in. POPin knowledge-sourcing software makes it simple for leaders to get prioritized ideas, solutions and buy-in. Have honest conversations to engage employees."
  url          = "https://app.popinnow.com/#/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "POPin"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer_domain>.popinnow.com/saml/sso/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_popin_app_popinnow_com,
    citrixspa_routing_domain.rd_popin_customer_fqdn,
  ]
}
