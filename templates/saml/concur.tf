# Concur — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_concur_www_concursolutions_com" {
  fqdn         = "www.Concursolutions.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Concur"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_concur_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Concur"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_concur" {
  name         = "Concur"
  type         = "saas"
  state        = "complete"
  description  = "Travel and expense management tool to manage expenses on the go."
  url          = "https://www.Concursolutions.com/SAMLRedirector/ClientSAMLLogin.aspx"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Concur"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.Concursolutions.com/SAMLRedirector/ClientSAMLLogin.aspx"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_concur_www_concursolutions_com,
    citrixspa_routing_domain.rd_concur_customer_fqdn,
  ]
}
