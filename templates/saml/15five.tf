# 15five — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_15five_customer_domain_15five_com" {
  fqdn         = "<customer-domain>.15five.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "15five"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_15five_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "15five"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app__15five" {
  name         = "15five"
  type         = "saas"
  state        = "complete"
  description  = "Continuous performance management tool to coach employees."
  url          = "https://<customer-domain>.15five.com/?next=/account/profile/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "15five"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.15five.com/saml2/acs/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "mail"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_15five_customer_domain_15five_com,
    citrixspa_routing_domain.rd_15five_customer_fqdn,
  ]
}
