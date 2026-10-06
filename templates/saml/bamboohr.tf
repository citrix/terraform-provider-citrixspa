# BambooHR — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bamboohr_company_domain_bamboohr_com" {
  fqdn         = "<company-domain>.bamboohr.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "BambooHR"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bamboohr_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "BambooHR"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bamboohr" {
  name         = "BambooHR"
  type         = "saas"
  state        = "complete"
  description  = "Human resources management tool to manage employee data."
  url          = "https://<company-domain>.bamboohr.com/home"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "BambooHR"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.bamboohr.com/saml/consume.php"
    audience          = "https://<company-domain>.bamboohr.com/saml/consume.php"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_bamboohr_company_domain_bamboohr_com,
    citrixspa_routing_domain.rd_bamboohr_customer_fqdn,
  ]
}
