# JFrog Artifactory — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_jfrog_artifactory_customer_domain_jfrog_io" {
  fqdn         = "<customer-domain>.jfrog.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "JFrog Artifactory"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_jfrog_artifactory_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "JFrog Artifactory"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_jfrog_artifactory" {
  name         = "JFrog Artifactory"
  type         = "saas"
  state        = "complete"
  description  = "Artifactory - Universal Artifact Repository Manager - Jfrog."
  url          = "https://<customer-domain>.jfrog.io/<customer-domain>/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "JFrog Artifactory"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.jfrog.io/<customer-domain>/webapp/saml/loginResponse"
    audience          = "https://<customer-domain>.jfrog.io/<customer-domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_jfrog_artifactory_customer_domain_jfrog_io,
    citrixspa_routing_domain.rd_jfrog_artifactory_customer_fqdn,
  ]
}
