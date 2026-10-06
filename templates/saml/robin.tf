# Robin — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_robin_dashboard_robinpowered_com" {
  fqdn         = "dashboard.robinpowered.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Robin"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_robin_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Robin"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_robin" {
  name         = "Robin"
  type         = "saas"
  state        = "complete"
  description  = "Workplace experience tools to schedule conference meeting rooms and desk bookings."
  url          = "https://dashboard.robinpowered.com/<username>/search/spaces"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Robin"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://dashboard.robinpowered.com/sso/saml/custom"
    audience          = "https://robinpowered.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_robin_dashboard_robinpowered_com,
    citrixspa_routing_domain.rd_robin_customer_fqdn,
  ]
}
