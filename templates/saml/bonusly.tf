# Bonusly — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bonusly_bonus_ly" {
  fqdn         = "bonus.ly"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bonusly"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bonusly_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bonusly"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bonusly" {
  name         = "Bonusly"
  type         = "saas"
  state        = "complete"
  description  = "Employee recognition and reward management tool to recognize team contributions."
  url          = "https://bonus.ly/bonuses"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Bonusly"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://bonus.ly/saml/<company_domain>-dot-com/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_bonusly_bonus_ly,
    citrixspa_routing_domain.rd_bonusly_customer_fqdn,
  ]
}
