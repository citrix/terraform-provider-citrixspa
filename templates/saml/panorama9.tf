# Panorama9 — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_panorama9_dashboard_panorama9_com" {
  fqdn         = "dashboard.panorama9.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Panorama9"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_panorama9_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Panorama9"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_panorama9" {
  name         = "Panorama9"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based IT management platform for enterprise network monitoring."
  url          = "https://dashboard.panorama9.com/saml/access/<customer_id>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Panorama9"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://dashboard.panorama9.com/saml/consume/<customer_id>"
    audience          = "https://dashboard.panorama9.com/saml/access/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "https://dashboard.panorama9.com/saml/consume/"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_panorama9_dashboard_panorama9_com,
    citrixspa_routing_domain.rd_panorama9_customer_fqdn,
  ]
}
