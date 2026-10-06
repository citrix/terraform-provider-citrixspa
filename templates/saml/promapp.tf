# promapp — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_promapp_freetrial_promapp_com" {
  fqdn         = "freetrial.promapp.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "promapp"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_promapp_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "promapp"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_promapp" {
  name         = "promapp"
  type         = "saas"
  state        = "complete"
  description  = "Business process management BPM tool."
  url          = "https://freetrial.promapp.com/<your-organization>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "promapp"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://freetrial.promapp.com/<your-organization>/saml/authenticate"
    audience          = "https://freetrial.promapp.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_promapp_freetrial_promapp_com,
    citrixspa_routing_domain.rd_promapp_customer_fqdn,
  ]
}
