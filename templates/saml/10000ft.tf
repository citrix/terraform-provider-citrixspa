# 10000ft — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_10000ft_app_10000ft_com" {
  fqdn         = "app.10000ft.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "10000ft"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_10000ft_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "10000ft"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app__10000ft" {
  name         = "10000ft"
  type         = "saas"
  state        = "complete"
  description  = "10000ft's high-level project management software helps teams get the clarity and insight they need to look ahead, avoid pitfalls, and plan for growth."
  url          = "https://app.10000ft.com/me"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "10000ft"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.10000ft.com/saml/acs"
    audience          = "https://app.10000ft.com/saml/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_10000ft_app_10000ft_com,
    citrixspa_routing_domain.rd_10000ft_customer_fqdn,
  ]
}
