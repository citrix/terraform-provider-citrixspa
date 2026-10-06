# Federated Directory — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_federated_directory_www_federated_directory" {
  fqdn         = "www.federated.directory"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Federated Directory"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_federated_directory_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Federated Directory"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_federated_directory" {
  name         = "Federated Directory"
  type         = "saas"
  state        = "complete"
  description  = "Cross-company contact directory tool to search through the corporate address books of different companies."
  url          = "https://www.federated.directory/of/<Customer_Domain>/myaccount"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Federated Directory"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.federated.directory/v2/Login/Saml2/<directory ID>/Acs"
    audience          = "federated.directory/<directory ID>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_federated_directory_www_federated_directory,
    citrixspa_routing_domain.rd_federated_directory_customer_fqdn,
  ]
}
