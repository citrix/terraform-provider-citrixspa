# xMatters — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_xmatters_company_id_xmatters_com" {
  fqdn         = "<company-Id>.xmatters.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "xMatters"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_xmatters_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "xMatters"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_xmatters" {
  name         = "xMatters"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration platform with an alerting software that integrates with other tools creating seamless process and effective communication."
  url          = "https://<company-Id>.xmatters.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "xMatters"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-Id>.xmatters.com/sp/SSO.saml2"
    audience          = "https://<company-Id>.xmatters.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_xmatters_company_id_xmatters_com,
    citrixspa_routing_domain.rd_xmatters_customer_fqdn,
  ]
}
