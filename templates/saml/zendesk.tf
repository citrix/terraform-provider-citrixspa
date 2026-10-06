# Zendesk — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zendesk_your_organization_zendesk_com" {
  fqdn         = "<your-organization>.zendesk.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zendesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zendesk_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zendesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zendesk" {
  name         = "Zendesk"
  type         = "saas"
  state        = "complete"
  description  = "Software to request for customer service and to log support tickets."
  url          = "https://<your-organization>.zendesk.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zendesk"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.zendesk.com/access/saml/"
    audience          = "<your-organization>.zendesk.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_zendesk_your_organization_zendesk_com,
    citrixspa_routing_domain.rd_zendesk_customer_fqdn,
  ]
}
