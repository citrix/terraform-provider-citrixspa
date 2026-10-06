# Contactzilla — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_contactzilla_customer_domain_contactzilla_com" {
  fqdn         = "<customer-domain>.contactzilla.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Contactzilla"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_contactzilla_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Contactzilla"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_contactzilla" {
  name         = "Contactzilla"
  type         = "saas"
  state        = "complete"
  description  = "Contact management tool to access up to date contact information."
  url          = "https://<customer-domain>.contactzilla.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Contactzilla"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://hq.contactzilla.com/saml/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_contactzilla_customer_domain_contactzilla_com,
    citrixspa_routing_domain.rd_contactzilla_customer_fqdn,
  ]
}
