# Workday — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_workday_impl_workday_com" {
  fqdn         = "impl.workday.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_workday_com" {
  fqdn         = "*.workday.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_myworkday_com" {
  fqdn         = "*.myworkday.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_myworkdayjobs_com" {
  fqdn         = "*.myworkdayjobs.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_myworkdaysite_com" {
  fqdn         = "*.myworkdaysite.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_akamaiedge_net" {
  fqdn         = "*.akamaiedge.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_workday_akamaihd_net" {
  fqdn         = "*.akamaihd.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Workday"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_workday" {
  name         = "Workday"
  type         = "saas"
  state        = "complete"
  description  = "Tool for financial management, human resources, and planning."
  url          = "https://impl.workday.com/<customer>/login-saml2.flex"
  related_urls = ["<Customer FQDN>", "*.workday.com", "*.myworkday.com", "*.myworkdayjobs.com", "*.myworkdaysite.com", "*.akamaiedge.net", "*.akamaihd.net"]
  icon         = "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"

  using_template   = true
  template_name    = "Workday"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://impl.workday.com/<customer>/login-saml.flex"
    audience          = "http://www.workday.com/<customer>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_workday_impl_workday_com,
    citrixspa_routing_domain.rd_workday_customer_fqdn,
    citrixspa_routing_domain.rd_workday_workday_com,
    citrixspa_routing_domain.rd_workday_myworkday_com,
    citrixspa_routing_domain.rd_workday_myworkdayjobs_com,
    citrixspa_routing_domain.rd_workday_myworkdaysite_com,
    citrixspa_routing_domain.rd_workday_akamaiedge_net,
    citrixspa_routing_domain.rd_workday_akamaihd_net,
  ]
}
