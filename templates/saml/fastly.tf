# Fastly — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_fastly_manage_fastly_com" {
  fqdn         = "manage.fastly.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Fastly"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_fastly_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Fastly"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_fastly" {
  name         = "Fastly"
  type         = "saas"
  state        = "complete"
  description  = "Edge cloud platform to serve and secure applications closer to the users."
  url          = "https://manage.fastly.com/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Fastly"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://manage.fastly.com/saml/consume"
    audience          = "https://api.fastly.com/saml/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_fastly_manage_fastly_com,
    citrixspa_routing_domain.rd_fastly_customer_fqdn,
  ]
}
