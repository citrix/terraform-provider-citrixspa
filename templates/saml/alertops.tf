# AlertOps — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_alertops_company_domain_alertops_com" {
  fqdn         = "<company-Domain>.alertops.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "AlertOps"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_alertops_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "AlertOps"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_alertops" {
  name         = "AlertOps"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration incidence response tool to manage IT incidents."
  url          = "https://<company-Domain>.alertops.com/MyHome.aspx"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "AlertOps"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-Domain>.alertops.com/login.aspx"
    audience          = "victorops.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_alertops_company_domain_alertops_com,
    citrixspa_routing_domain.rd_alertops_customer_fqdn,
  ]
}
