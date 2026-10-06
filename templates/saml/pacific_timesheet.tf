# Pacific Timesheet — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_pacific_timesheet_customer_domain_pacifictimesheet_com" {
  fqdn         = "<customer-domain>.pacifictimesheet.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pacific Timesheet"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_pacific_timesheet_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pacific Timesheet"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_pacific_timesheet" {
  name         = "Pacific Timesheet"
  type         = "saas"
  state        = "complete"
  description  = "Timesheet software, expense tracking for payroll"
  url          = "https://<customer-domain>.pacifictimesheet.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Pacific Timesheet"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.pacifictimesheet.com/timesheet/home.do"
    audience          = "https://<customer-domain>.pacifictimesheet.com/timesheet/home.do"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_pacific_timesheet_customer_domain_pacifictimesheet_com,
    citrixspa_routing_domain.rd_pacific_timesheet_customer_fqdn,
  ]
}
