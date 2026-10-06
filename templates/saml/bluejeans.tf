# Bluejeans — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bluejeans_company_domain_bluejeans_com" {
  fqdn         = "<company_domain>.bluejeans.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bluejeans"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bluejeans_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bluejeans"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bluejeans" {
  name         = "Bluejeans"
  type         = "saas"
  state        = "complete"
  description  = "Cloud Video Conferencing"
  url          = "https://<company_domain>.bluejeans.com/enterprise-admin/#onboarding"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Bluejeans"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://bluejeans.com/sso/saml2/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_bluejeans_company_domain_bluejeans_com,
    citrixspa_routing_domain.rd_bluejeans_customer_fqdn,
  ]
}
