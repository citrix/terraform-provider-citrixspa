# Humanity — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_humanity_customer_domain_humanity_com" {
  fqdn         = "<customer-domain>.humanity.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Humanity"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_humanity_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Humanity"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_humanity" {
  name         = "Humanity"
  type         = "saas"
  state        = "complete"
  description  = "Online employee scheduling software to manages shifts, schedules, payroll, and time clocking."
  url          = "https://<customer-domain>.humanity.com/app/dashboard/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Humanity"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.humanity.com/includes/saml/consume.php"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_humanity_customer_domain_humanity_com,
    citrixspa_routing_domain.rd_humanity_customer_fqdn,
  ]
}
