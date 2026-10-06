# Zoom — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zoom_your_organization_zoom_us" {
  fqdn         = "<your-organization>.zoom.us"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zoom"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zoom_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zoom"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zoom" {
  name         = "Zoom"
  type         = "saas"
  state        = "complete"
  description  = "Communication and collaboration software with voice, web collaboration, and video conferencing capabilities."
  url          = "https://<your-organization>.zoom.us/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zoom"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.zoom.us/SAML/SSO"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "First Name"
        value = "ns_user_name"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_zoom_your_organization_zoom_us,
    citrixspa_routing_domain.rd_zoom_customer_fqdn,
  ]
}
