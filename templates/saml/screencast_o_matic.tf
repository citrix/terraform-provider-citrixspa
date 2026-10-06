# screencast-o-matic — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_screencast_o_matic_screencast_o_matic_com" {
  fqdn         = "screencast-o-matic.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "screencast-o-matic"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_screencast_o_matic_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "screencast-o-matic"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_screencast_o_matic" {
  name         = "screencast-o-matic"
  type         = "saas"
  state        = "complete"
  description  = "Tool to screencast and edit video."
  url          = "https://screencast-o-matic.com/<account>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "screencast-o-matic"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://screencast-o-matic.com/saml/sp"
    audience          = "https://screencast-o-matic.com/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_screencast_o_matic_screencast_o_matic_com,
    citrixspa_routing_domain.rd_screencast_o_matic_customer_fqdn,
  ]
}
