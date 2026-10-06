# Insightly — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_insightly_crm_na1_insightly_com" {
  fqdn         = "crm.na1.insightly.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Insightly"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_insightly_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Insightly"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_insightly" {
  name         = "Insightly"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based customer relationship management CRM and project management tools for small and medium size businesses."
  url          = "https://crm.na1.insightly.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Insightly"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://crm.na1.insightly.com/user/saml?instanceId=<Customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_insightly_crm_na1_insightly_com,
    citrixspa_routing_domain.rd_insightly_customer_fqdn,
  ]
}
