# CONVO — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_convo_app_convo_com" {
  fqdn         = "app.convo.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CONVO"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_convo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CONVO"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_convo" {
  name         = "CONVO"
  type         = "saas"
  state        = "complete"
  description  = "Team communication and collaboration tool for internal conversations."
  url          = "https://app.convo.com/app/login"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "CONVO"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.convo.com/app/saml2/v1/sso/saml_sp/module.php/saml/sp/saml2-acs.php/convo-sp/acc-<ACC-ID>"
    audience          = "<CONVO-APP-ID>"
    relay_state       = "https://app.convo.com/app/saml2/v1/sso/sso_relay_state.php?account_id=acc-<ACC-ID>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_convo_app_convo_com,
    citrixspa_routing_domain.rd_convo_customer_fqdn,
  ]
}
