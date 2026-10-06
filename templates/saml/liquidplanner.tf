# LiquidPlanner — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_liquidplanner_app_liquidplanner_com" {
  fqdn         = "app.liquidplanner.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "LiquidPlanner"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_liquidplanner_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "LiquidPlanner"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_liquidplanner" {
  name         = "LiquidPlanner"
  type         = "saas"
  state        = "complete"
  description  = "Online project management software for your business."
  url          = "https://app.liquidplanner.com/login"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "LiquidPlanner"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.liquidplanner.com/<org-value>/auth/saml/callback"
    audience          = "https://app.liquidplanner.com/<org-value>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_liquidplanner_app_liquidplanner_com,
    citrixspa_routing_domain.rd_liquidplanner_customer_fqdn,
  ]
}
