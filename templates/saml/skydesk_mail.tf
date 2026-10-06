# SkyDesk Mail — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_skydesk_mail_mail_skydesk_jp" {
  fqdn         = "mail.skydesk.jp"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SkyDesk Mail"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_skydesk_mail_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SkyDesk Mail"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_skydesk_mail" {
  name         = "SkyDesk Mail"
  type         = "saas"
  state        = "complete"
  description  = "Email Hosting Hosted Email for Businesses - SkyDesk Mail"
  url          = "https://mail.skydesk.jp/portal/<your-organization>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "SkyDesk Mail"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://accounts.skydesk.jp/samlresponse/<your-organization>.business.skydesk.jp"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_skydesk_mail_mail_skydesk_jp,
    citrixspa_routing_domain.rd_skydesk_mail_customer_fqdn,
  ]
}
