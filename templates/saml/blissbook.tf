# Blissbook — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_blissbook_your_organization_blissbook_com" {
  fqdn         = "<your-organization>.blissbook.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Blissbook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_blissbook_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Blissbook"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_blissbook" {
  name         = "Blissbook"
  type         = "saas"
  state        = "complete"
  description  = "Policy management tool to create employee handbooks."
  url          = "https://<your-organization>.blissbook.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Blissbook"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.blissbook.com/auth/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_blissbook_your_organization_blissbook_com,
    citrixspa_routing_domain.rd_blissbook_customer_fqdn,
  ]
}
