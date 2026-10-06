# Hoshinplan — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_hoshinplan_en_hoshinplan_com" {
  fqdn         = "en.hoshinplan.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Hoshinplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_hoshinplan_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Hoshinplan"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_hoshinplan" {
  name         = "Hoshinplan"
  type         = "saas"
  state        = "complete"
  description  = "Tool to visualize your strategic plans and track statuses in one canvas."
  url          = "https://en.hoshinplan.com/companies/<your-orgid>-<your-organization>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Hoshinplan"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.hoshinplan.com/auth/saml_<copy from Hoshinplan metadata AssertationConsumerService>/callback"
    audience          = "https://www.hoshinplan.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "email"
        value  = "ns_user_email"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_hoshinplan_en_hoshinplan_com,
    citrixspa_routing_domain.rd_hoshinplan_customer_fqdn,
  ]
}
