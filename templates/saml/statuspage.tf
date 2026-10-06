# Statuspage — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_statuspage_manage_statuspage_io" {
  fqdn         = "manage.statuspage.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Statuspage"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_statuspage_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Statuspage"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_statuspage" {
  name         = "Statuspage"
  type         = "saas"
  state        = "complete"
  description  = "Tool to communicate status and incidents."
  url          = "https://manage.statuspage.io/login"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Statuspage"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://manage.statuspage.io/sso/saml/consume"
    audience          = "https://<your-orgid>.statuspage.io/"
    sign_assertion    = "RESPONSE"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_statuspage_manage_statuspage_io,
    citrixspa_routing_domain.rd_statuspage_customer_fqdn,
  ]
}
