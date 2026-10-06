# Monday — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_monday_customer_domain_monday_com" {
  fqdn         = "<Customer-domain>.monday.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Monday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_monday_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Monday"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_monday" {
  name         = "Monday"
  type         = "saas"
  state        = "complete"
  description  = "Team management software to plan, track, and collaborate all your work in one tool."
  url          = "https://<Customer-domain>.monday.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Monday"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.monday.com/saml/saml_callback"
    audience          = "https://<Customer-domain>.monday.com/saml/saml_callback"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_monday_customer_domain_monday_com,
    citrixspa_routing_domain.rd_monday_customer_fqdn,
  ]
}
