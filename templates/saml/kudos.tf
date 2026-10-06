# Kudos — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_kudos_your_organization_kudosnow_com" {
  fqdn         = "<your-organization>.kudosnow.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Kudos"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_kudos_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Kudos"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_kudos" {
  name         = "Kudos"
  type         = "saas"
  state        = "complete"
  description  = "Retail, job, project and fulfilment process systems."
  url          = "https://<your-organization>.kudosnow.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Kudos"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.kudosnow.com/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_kudos_your_organization_kudosnow_com,
    citrixspa_routing_domain.rd_kudos_customer_fqdn,
  ]
}
