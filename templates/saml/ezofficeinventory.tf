# EZOfficeInventory — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_ezofficeinventory_company_name_ezofficeinventory_com" {
  fqdn         = "<company-name>.ezofficeinventory.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EZOfficeInventory"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_ezofficeinventory_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EZOfficeInventory"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_ezofficeinventory" {
  name         = "EZOfficeInventory"
  type         = "saas"
  state        = "complete"
  description  = "Inventory management tool to track all your assets and equipments."
  url          = "https://<company-name>.ezofficeinventory.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "EZOfficeInventory"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-name>.ezofficeinventory.com/users/auth/saml/callback"
    audience          = "https://www.ezofficeinventory.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_ezofficeinventory_company_name_ezofficeinventory_com,
    citrixspa_routing_domain.rd_ezofficeinventory_customer_fqdn,
  ]
}
