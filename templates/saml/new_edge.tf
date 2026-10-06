# New Edge — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_new_edge_customer_domain_newedge_io" {
  fqdn         = "<customer-domain>.newedge.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "New Edge"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_new_edge_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "New Edge"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_new_edge" {
  name         = "New Edge"
  type         = "saas"
  state        = "complete"
  description  = "Secure application networking service for Hybrid IT."
  url          = "https://<customer-domain>.newedge.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "New Edge"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.newedge.io/auth/saml/callback"
    audience          = "https://<customer-domain>.newedge.io"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_new_edge_customer_domain_newedge_io,
    citrixspa_routing_domain.rd_new_edge_customer_fqdn,
  ]
}
