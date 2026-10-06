# Marketo — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_marketo_customer_domain_marketo_com" {
  fqdn         = "<customer-domain>.marketo.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Marketo"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_marketo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Marketo"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_marketo" {
  name         = "Marketo"
  type         = "saas"
  state        = "complete"
  description  = "Automation software to help marketing teams master the art and science of digital marketing."
  url          = "https://<customer-domain>.marketo.com/#MM0A1"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Marketo"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.marketo.com/saml/assertion/<Munchkin id>"
    audience          = "http://saml.marketo.com/sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_marketo_customer_domain_marketo_com,
    citrixspa_routing_domain.rd_marketo_customer_fqdn,
  ]
}
