# Cisco Umbrella — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cisco_umbrella_dashboard_umbrella_com" {
  fqdn         = "dashboard.umbrella.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cisco Umbrella"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cisco_umbrella_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Cisco Umbrella"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cisco_umbrella" {
  name         = "Cisco Umbrella"
  type         = "saas"
  state        = "complete"
  description  = "Cloud security platform to provide the first line of defense against threats on the internet."
  url          = "https://dashboard.umbrella.com/o/<customer_id>/#/overview"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Cisco Umbrella"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.umbrella.com/sso"
    audience          = "https://login.umbrella.com/sso"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_cisco_umbrella_dashboard_umbrella_com,
    citrixspa_routing_domain.rd_cisco_umbrella_customer_fqdn,
  ]
}
