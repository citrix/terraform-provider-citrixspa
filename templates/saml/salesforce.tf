# Salesforce — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_salesforce_customer_domain_my_salesforce_com" {
  fqdn         = "<customer-domain>.my.salesforce.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_salesforce_com" {
  fqdn         = "*.salesforce.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_force_com" {
  fqdn         = "*.force.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_lightning_com" {
  fqdn         = "*.lightning.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_visualforce_com" {
  fqdn         = "*.visualforce.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_forceusercontent_com" {
  fqdn         = "*.forceusercontent.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_salesforce_sites_com" {
  fqdn         = "*.salesforce-sites.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_trailhead_com" {
  fqdn         = "*.trailhead.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_salesforce_trailblazer_me" {
  fqdn         = "*.trailblazer.me"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Salesforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_salesforce" {
  name         = "Salesforce"
  type         = "saas"
  state        = "complete"
  description  = "CRM tool to manage customer contact information, integrate social media, and facilitate real-time customer collaboration."
  url          = "https://<customer-domain>.my.salesforce.com/?so=<customer_id>"
  related_urls = ["<Customer FQDN>", "*.salesforce.com", "*.force.com", "*.lightning.com", "*.visualforce.com", "*.forceusercontent.com", "*.salesforce-sites.com", "*.trailhead.com", "*.trailblazer.me"]
  icon         = "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"

  using_template   = true
  template_name    = "Salesforce"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.my.salesforce.com?so=<customer_id>"
    audience          = "https://<customer-domain>.my.salesforce.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "name"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_salesforce_customer_domain_my_salesforce_com,
    citrixspa_routing_domain.rd_salesforce_customer_fqdn,
    citrixspa_routing_domain.rd_salesforce_salesforce_com,
    citrixspa_routing_domain.rd_salesforce_force_com,
    citrixspa_routing_domain.rd_salesforce_lightning_com,
    citrixspa_routing_domain.rd_salesforce_visualforce_com,
    citrixspa_routing_domain.rd_salesforce_forceusercontent_com,
    citrixspa_routing_domain.rd_salesforce_salesforce_sites_com,
    citrixspa_routing_domain.rd_salesforce_trailhead_com,
    citrixspa_routing_domain.rd_salesforce_trailblazer_me,
  ]
}
