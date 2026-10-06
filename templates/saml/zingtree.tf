# Zingtree — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zingtree_zingtree_com" {
  fqdn         = "zingtree.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zingtree"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zingtree_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zingtree"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zingtree" {
  name         = "Zingtree"
  type         = "saas"
  state        = "complete"
  description  = "A toolkit for creating interactive decision trees and troubleshooters."
  url          = "https://zingtree.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zingtree"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://zingtree.com/sso-author/login-callback/<customer_id>"
    audience          = "https://zingtree.com/sso-author/metadata/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_zingtree_zingtree_com,
    citrixspa_routing_domain.rd_zingtree_customer_fqdn,
  ]
}
