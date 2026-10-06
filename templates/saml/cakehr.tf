# CakeHR — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cakehr_customer_domain_cake_hr" {
  fqdn         = "<customer-domain>.cake.hr"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CakeHR"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cakehr_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CakeHR"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cakehr" {
  name         = "CakeHR"
  type         = "saas"
  state        = "complete"
  description  = "Human resources management tool for attendance and performance management."
  url          = "https://<customer-domain>.cake.hr/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "CakeHR"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.cake.hr/services/saml/consume"
    audience          = "cake.hr"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_cakehr_customer_domain_cake_hr,
    citrixspa_routing_domain.rd_cakehr_customer_fqdn,
  ]
}
