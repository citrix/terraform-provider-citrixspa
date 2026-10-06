# CircleHD — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_circlehd_customer_domain_circlehd_com" {
  fqdn         = "<Customer-domain>.circlehd.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CircleHD"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_circlehd_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CircleHD"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_circlehd" {
  name         = "CircleHD"
  type         = "saas"
  state        = "complete"
  description  = "Training, learning, and collaboration tool to share videos and slides within the organization."
  url          = "https://<Customer-domain>.circlehd.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "CircleHD"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.circlehd.com/auth/saml2"
    audience          = "https://<Customer-domain>.circlehd.com/auth/saml2/metadata.xml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_circlehd_customer_domain_circlehd_com,
    citrixspa_routing_domain.rd_circlehd_customer_fqdn,
  ]
}
