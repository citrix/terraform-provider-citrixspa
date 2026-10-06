# Receptive — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_receptive_receptive_io" {
  fqdn         = "receptive.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Receptive"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_receptive_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Receptive"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_receptive" {
  name         = "Receptive"
  type         = "saas"
  state        = "complete"
  description  = "Leading SaaS companies use Receptive to build winning products. Gather prioritized feedback from your customers, teams and the market - all in one place."
  url          = "https://receptive.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Receptive"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.eu-west-1.receptive.io/saml/consume?vid=[vid]&user=[vu/eu]"
    audience          = "https://receptive.io/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_receptive_receptive_io,
    citrixspa_routing_domain.rd_receptive_customer_fqdn,
  ]
}
