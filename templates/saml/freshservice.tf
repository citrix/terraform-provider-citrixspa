# Freshservice — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_freshservice_customer_domain_freshservice_com" {
  fqdn         = "<customer-domain>.freshservice.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Freshservice"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_freshservice_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Freshservice"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_freshservice" {
  name         = "Freshservice"
  type         = "saas"
  state        = "complete"
  description  = "IT help desk tool to simplify IT operations."
  url          = "https://<customer-domain>.freshservice.com/login/sso"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Freshservice"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.freshservice.com/login/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_freshservice_customer_domain_freshservice_com,
    citrixspa_routing_domain.rd_freshservice_customer_fqdn,
  ]
}
