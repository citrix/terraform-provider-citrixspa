# skyprep — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_skyprep_your_orgname_skyprepapp_com" {
  fqdn         = "<your-Orgname>.skyprepapp.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "skyprep"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_skyprep_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "skyprep"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_skyprep" {
  name         = "skyprep"
  type         = "saas"
  state        = "complete"
  description  = "Learning management system LMS to train customers and employees."
  url          = "https://<your-Orgname>.skyprepapp.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "skyprep"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-Orgname>.skyprepapp.com/saml/consume"
    audience          = "https://<your-Orgname>.skyprepapp.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_skyprep_your_orgname_skyprepapp_com,
    citrixspa_routing_domain.rd_skyprep_customer_fqdn,
  ]
}
