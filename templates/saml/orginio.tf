# Orginio — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_orginio_your_organization_orginio_com" {
  fqdn         = "<your-organization>.orginio.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Orginio"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_orginio_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Orginio"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_orginio" {
  name         = "Orginio"
  type         = "saas"
  state        = "complete"
  description  = "Online organizational chart creation tool to visualize the organizational structure."
  url          = "https://<your-organization>.orginio.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Orginio"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.orginio.com/api/auth/getToken/ui"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_orginio_your_organization_orginio_com,
    citrixspa_routing_domain.rd_orginio_customer_fqdn,
  ]
}
