# Realtime Board — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_realtime_board_realtimeboard_com" {
  fqdn         = "realtimeboard.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Realtime Board"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_realtime_board_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Realtime Board"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_realtime_board" {
  name         = "Realtime Board"
  type         = "saas"
  state        = "complete"
  description  = "Online Whiteboard Platform for Team Collaboration"
  url          = "https://realtimeboard.com/app/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Realtime Board"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://realtimeboard.com/sso/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_realtime_board_realtimeboard_com,
    citrixspa_routing_domain.rd_realtime_board_customer_fqdn,
  ]
}
