# Ziflow — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ziflow_customer_domain_ziflow_io" {
  fqdn         = "<customer-domain>.ziflow.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ziflow"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ziflow_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ziflow"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ziflow" {
  name         = "Ziflow"
  type         = "saas"
  state        = "complete"
  description  = "Online Proofing for Agencies and Brands. Enterprise online proofing for agencies and brands which simplifies the review and approval process of creative content. Click through and get started with..."
  url          = "https://<customer-domain>.ziflow.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Ziflow"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.ziflow.io/"
    audience          = "https://<customer-domain>.ziflow.io/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_ziflow_customer_domain_ziflow_io,
    citrixspa_routing_domain.rd_ziflow_customer_fqdn,
  ]
}
