# 4me — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_4me_company_domain_4me_qa" {
  fqdn         = "<company-domain>.4me.qa"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "4me"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_4me_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "4me"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app__4me" {
  name         = "4me"
  type         = "saas"
  state        = "complete"
  description  = "Service management tool for collaboration between internal, external, and outsourced teams."
  url          = "https://<company-domain>.4me.qa/inbox"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "4me"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.4me.qa/access/saml/consume"
    audience          = "https://<company-domain>.4me.qa"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_4me_company_domain_4me_qa,
    citrixspa_routing_domain.rd_4me_customer_fqdn,
  ]
}
