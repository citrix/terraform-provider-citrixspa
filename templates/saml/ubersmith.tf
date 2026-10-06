# Ubersmith — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ubersmith_customer_domain_trial_ubersmith_com" {
  fqdn         = "<customer-domain>.trial.ubersmith.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ubersmith"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ubersmith_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Ubersmith"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ubersmith" {
  name         = "Ubersmith"
  type         = "saas"
  state        = "complete"
  description  = "Business management software for usage-based billing, quoting, order management, infrastructure management, and help desk ticketing solutions."
  url          = "https://<customer-domain>.trial.ubersmith.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Ubersmith"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.trial.ubersmith.com/"
    audience          = "https://<customer-domain>.trial.ubersmith.com/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
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
    citrixspa_routing_domain.rd_ubersmith_customer_domain_trial_ubersmith_com,
    citrixspa_routing_domain.rd_ubersmith_customer_fqdn,
  ]
}
