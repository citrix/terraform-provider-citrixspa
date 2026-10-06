# Domo — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_domo_your_organization_domo_com" {
  fqdn         = "<your-organization>.domo.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Domo"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_domo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Domo"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_domo" {
  name         = "Domo"
  type         = "saas"
  state        = "complete"
  description  = "Business Intelligence tools and Data Visualization"
  url          = "https://<your-organization>.domo.com/auth/saml"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Domo"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.domo.com/auth/saml"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_domo_your_organization_domo_com,
    citrixspa_routing_domain.rd_domo_customer_fqdn,
  ]
}
