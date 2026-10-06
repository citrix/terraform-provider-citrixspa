# Sigma Computing — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_sigma_computing_app_sigmacomputing_com" {
  fqdn         = "app.sigmacomputing.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Sigma Computing"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_sigma_computing_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Sigma Computing"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_sigma_computing" {
  name         = "Sigma Computing"
  type         = "saas"
  state        = "complete"
  description  = "Analytics tool to explore, analyze, and visualize data."
  url          = "https://app.sigmacomputing.com/<Customer-domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Sigma Computing"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.sigmacomputing.com/api/v2/saml2/assert"
    audience          = "https://api.sigmacomputing.com/api/v2/saml2/metadata.xml"
    relay_state       = "https://app.sigmacomputing.com/<Customer-domain>/finish-login"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_sigma_computing_app_sigmacomputing_com,
    citrixspa_routing_domain.rd_sigma_computing_customer_fqdn,
  ]
}
