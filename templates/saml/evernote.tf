# Evernote — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_evernote_www_evernote_com" {
  fqdn         = "www.evernote.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Evernote"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_evernote_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Evernote"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_evernote" {
  name         = "Evernote"
  type         = "saas"
  state        = "complete"
  description  = "It is an app designed for note taking"
  url          = "https://www.evernote.com/business/AccountSettings.action"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Evernote"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.evernote.com/SamlConsumer.action"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_evernote_www_evernote_com,
    citrixspa_routing_domain.rd_evernote_customer_fqdn,
  ]
}
