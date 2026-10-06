# Creative Cloud — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_creative_cloud_adobe_com" {
  fqdn         = "adobe.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Creative Cloud"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_creative_cloud_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Creative Cloud"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_creative_cloud" {
  name         = "Creative Cloud"
  type         = "saas"
  state        = "complete"
  description  = "Software and services for creative professionals"
  url          = "https://adobe.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Creative Cloud"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-org-id>/auth/saml20/accauthlinktest"
    audience          = "<get the Entity Id from Metadata>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_creative_cloud_adobe_com,
    citrixspa_routing_domain.rd_creative_cloud_customer_fqdn,
  ]
}
