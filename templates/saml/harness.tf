# Harness — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_harness_app_harness_io" {
  fqdn         = "app.harness.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Harness"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_harness_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Harness"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_harness" {
  name         = "Harness"
  type         = "saas"
  state        = "complete"
  description  = "Continuous Delivery Simplified CI CD for Devops. Continuous Delivery and Integration for Java,.NET apps in AWS, GCP, Azure, Bare Metal. DevOps teams can quickly, safely, and securely deliver their..."
  url          = "https://app.harness.io/#/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Harness"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.harness.io/api/users/saml-login"
    audience          = "app.harness.io"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_harness_app_harness_io,
    citrixspa_routing_domain.rd_harness_customer_fqdn,
  ]
}
