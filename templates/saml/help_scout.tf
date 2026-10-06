# Help Scout — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_help_scout_secure_helpscout_net" {
  fqdn         = "secure.helpscout.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Help Scout"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_help_scout_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Help Scout"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_help_scout" {
  name         = "Help Scout"
  type         = "saas"
  state        = "complete"
  description  = "Customer service software and knowledge base tool for customer service professionals."
  url          = "https://secure.helpscout.net/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Help Scout"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://helpscout.auth0.com/login/callback?connection=<customer_domain>-<customer_id>-sso-saml"
    audience          = "urn:auth0:helpscout:<customer_domain>-<customer_id>-sso-saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_help_scout_secure_helpscout_net,
    citrixspa_routing_domain.rd_help_scout_customer_fqdn,
  ]
}
