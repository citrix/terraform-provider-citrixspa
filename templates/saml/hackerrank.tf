# HackerRank — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_hackerrank_www_hackerrank_com" {
  fqdn         = "www.hackerrank.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HackerRank"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_hackerrank_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HackerRank"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_hackerrank" {
  name         = "HackerRank"
  type         = "saas"
  state        = "complete"
  description  = "Powering Tech Recruiting with Machine Learning"
  url          = "https://www.hackerrank.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "HackerRank"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.hackerrank.com/x/api/v1/sso/saml/<Customer-id>/acs"
    audience          = "https://www.hackerrank.com/x/api/v1/sso/saml/<Customer-id>/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_hackerrank_www_hackerrank_com,
    citrixspa_routing_domain.rd_hackerrank_customer_fqdn,
  ]
}
