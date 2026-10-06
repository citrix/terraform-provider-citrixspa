# Beekeeper — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_beekeeper_customer_id_beekeeper_io" {
  fqdn         = "<customer-id>.beekeeper.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Beekeeper"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_beekeeper_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Beekeeper"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_beekeeper" {
  name         = "Beekeeper"
  type         = "saas"
  state        = "complete"
  description  = "The Secure Employee App"
  url          = "https://<customer-id>.beekeeper.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Beekeeper"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-id>.beekeeper.io/saml/sso"
    audience          = "https://<customer-id>.beekeeper.io/saml/sso/metadata.xml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_beekeeper_customer_id_beekeeper_io,
    citrixspa_routing_domain.rd_beekeeper_customer_fqdn,
  ]
}
