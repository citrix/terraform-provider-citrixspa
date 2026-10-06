# Stackify — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_stackify_customer_id_stackify_com" {
  fqdn         = "<Customer-id>.stackify.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Stackify"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_stackify_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Stackify"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_stackify" {
  name         = "Stackify"
  type         = "saas"
  state        = "complete"
  description  = "Stackify offers the only solution that fully integrates application performance monitoring with errors and logs"
  url          = "https://<Customer-id>.stackify.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Stackify"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-id>.stackify.com/sso/saml2?clienttoken=<Account-id>"
    audience          = "https://<Customer-id>.stackify.com/sso"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_stackify_customer_id_stackify_com,
    citrixspa_routing_domain.rd_stackify_customer_fqdn,
  ]
}
