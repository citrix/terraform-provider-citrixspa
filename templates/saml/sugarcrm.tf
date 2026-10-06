# SugarCRM — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_sugarcrm_customer_id_trial_sugarcrm_com" {
  fqdn         = "<Customer-id>.trial.sugarcrm.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SugarCRM"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_sugarcrm_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SugarCRM"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_sugarcrm" {
  name         = "SugarCRM"
  type         = "saas"
  state        = "complete"
  description  = "Rated Customer Relationship Management Software"
  url          = "https://<Customer-id>.trial.sugarcrm.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "SugarCRM"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-id>.trial.sugarcrm.com/index.php?module=Users&action=Authenticate"
    audience          = "php-saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_sugarcrm_customer_id_trial_sugarcrm_com,
    citrixspa_routing_domain.rd_sugarcrm_customer_fqdn,
  ]
}
