# AppFollow — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_appfollow_watch_appfollow_io" {
  fqdn         = "watch.appfollow.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "AppFollow"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_appfollow_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "AppFollow"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_appfollow" {
  name         = "AppFollow"
  type         = "saas"
  state        = "complete"
  description  = "AppFollow is a product management tool for accelerating global app growth and increasing customer loyalty."
  url          = "https://watch.appfollow.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "AppFollow"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sso.appfollow.io/acs"
    audience          = "https://sso.appfollow.io/metadata/"
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
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_appfollow_watch_appfollow_io,
    citrixspa_routing_domain.rd_appfollow_customer_fqdn,
  ]
}
