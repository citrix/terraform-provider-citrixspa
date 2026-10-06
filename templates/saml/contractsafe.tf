# ContractSafe — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_contractsafe_app_contractsafe_com" {
  fqdn         = "app.contractsafe.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ContractSafe"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_contractsafe_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ContractSafe"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_contractsafe" {
  name         = "ContractSafe"
  type         = "saas"
  state        = "complete"
  description  = "Contract management tool to track, store, and manage contracts."
  url          = "https://app.contractsafe.com/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "ContractSafe"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.contractsafe.com/saml2_auth/<Customer_id>/acs/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_contractsafe_app_contractsafe_com,
    citrixspa_routing_domain.rd_contractsafe_customer_fqdn,
  ]
}
