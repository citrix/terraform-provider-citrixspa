# Coggle — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_coggle_coggle_it" {
  fqdn         = "coggle.it"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Coggle"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_coggle_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Coggle"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_coggle" {
  name         = "Coggle"
  type         = "saas"
  state        = "complete"
  description  = "Simple Collaborative Mind Maps and Flow Charts. The clear way to share complex information. Coggle is a collaborative mind-mapping tool that helps you make sense of complex things. Create unlimited mind..."
  url          = "https://coggle.it/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Coggle"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://coggle.it/auth/saml/callback"
    audience          = "https://coggle.it"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_coggle_coggle_it,
    citrixspa_routing_domain.rd_coggle_customer_fqdn,
  ]
}
