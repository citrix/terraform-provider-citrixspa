# Sentry — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_sentry_sentry_io" {
  fqdn         = "sentry.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Sentry"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_sentry_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Sentry"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_sentry" {
  name         = "Sentry"
  type         = "saas"
  state        = "complete"
  description  = "Open-source error tracking software."
  url          = "https://sentry.io/<customer_domain>/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Sentry"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sentry.io/saml/acs/<customer_domain>/"
    audience          = "https://sentry.io/saml/metadata/<customer_domain>/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "UserID"
        value = "<user_id>"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_sentry_sentry_io,
    citrixspa_routing_domain.rd_sentry_customer_fqdn,
  ]
}
