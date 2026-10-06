# Helpmonks — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_helpmonks_customer_domain_helpmonks_com" {
  fqdn         = "<Customer-domain>.helpmonks.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Helpmonks"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_helpmonks_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Helpmonks"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_helpmonks" {
  name         = "Helpmonks"
  type         = "saas"
  state        = "complete"
  description  = "The collaborative email platform for your team..."
  url          = "https://<Customer-domain>.helpmonks.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Helpmonks"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.helpmonks.com/p/sso/saml/assert?h=<Customer-id>"
    audience          = "https://app.helpmonks.com/p/sso/saml/metadata?h=<Customer-id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_helpmonks_customer_domain_helpmonks_com,
    citrixspa_routing_domain.rd_helpmonks_customer_fqdn,
  ]
}
