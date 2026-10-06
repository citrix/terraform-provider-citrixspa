# Buddy Punch — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_buddy_punch_app_buddypunch_com" {
  fqdn         = "app.buddypunch.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Buddy Punch"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_buddy_punch_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Buddy Punch"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_buddy_punch" {
  name         = "Buddy Punch"
  type         = "saas"
  state        = "complete"
  description  = "Time management tool to monitor employee attendance."
  url          = "https://app.buddypunch.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Buddy Punch"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.buddypunch.com/Saml2/Acs"
    audience          = "https://app.buddypunch.com/Saml2"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_buddy_punch_app_buddypunch_com,
    citrixspa_routing_domain.rd_buddy_punch_customer_fqdn,
  ]
}
