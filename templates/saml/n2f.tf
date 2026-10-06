# N2F — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_n2f_www_n2f_com" {
  fqdn         = "www.n2f.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "N2F"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_n2f_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "N2F"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_n2f" {
  name         = "N2F"
  type         = "saas"
  state        = "complete"
  description  = "Expense report management tool to manage your business and travel expenses."
  url          = "https://www.n2f.com/app"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "N2F"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.n2f.com/app/Saml2/Acs"
    audience          = "https://www.n2f.com/app/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_n2f_www_n2f_com,
    citrixspa_routing_domain.rd_n2f_customer_fqdn,
  ]
}
