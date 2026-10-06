# interactgo — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_interactgo_customer_domain_interactgo_com" {
  fqdn         = "<customer-domain>.interactgo.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "interactgo"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_interactgo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "interactgo"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_interactgo" {
  name         = "interactgo"
  type         = "saas"
  state        = "complete"
  description  = "Tool to measure real-time and historical data on system performance."
  url          = "https://<customer-domain>.interactgo.com/InteractV7/ManageHomepages"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "interactgo"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.interactgo.com/Interact/Login/default.aspx"
    audience          = "http://<customer-domain>.interactgo.com/saml-sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_interactgo_customer_domain_interactgo_com,
    citrixspa_routing_domain.rd_interactgo_customer_fqdn,
  ]
}
