# ClearSlide — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_clearslide_www_clearslide_com" {
  fqdn         = "www.clearslide.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ClearSlide"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_clearslide_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ClearSlide"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_clearslide" {
  name         = "ClearSlide"
  type         = "saas"
  state        = "complete"
  description  = "Sales engagement tool to let users share content and sales material for customer interaction."
  url          = "https://www.clearslide.com/me"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "ClearSlide"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.clearslide.com/auth/saml_login_tx?tm=<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_clearslide_www_clearslide_com,
    citrixspa_routing_domain.rd_clearslide_customer_fqdn,
  ]
}
