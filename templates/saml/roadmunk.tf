# Roadmunk — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_roadmunk_app_roadmunk_com" {
  fqdn         = "app.roadmunk.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Roadmunk"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_roadmunk_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Roadmunk"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_roadmunk" {
  name         = "Roadmunk"
  type         = "saas"
  state        = "complete"
  description  = "Product roadmap software and roadmap tool to create product roadmaps."
  url          = "https://app.roadmunk.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Roadmunk"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.roadmunk.com/login/saml/<Customer-domain>/consume"
    audience          = "https://app.roadmunk.com/login/saml/<Customer-domain>/metadata.xml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_roadmunk_app_roadmunk_com,
    citrixspa_routing_domain.rd_roadmunk_customer_fqdn,
  ]
}
