# Remedyforce — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_remedyforce_ap4_salesforce_com" {
  fqdn         = "ap4.salesforce.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Remedyforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_remedyforce_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Remedyforce"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_remedyforce" {
  name         = "Remedyforce"
  type         = "saas"
  state        = "complete"
  description  = "IT service management and help desk system."
  url          = "https://ap4.salesforce.com/home/home.jsp"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Remedyforce"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.my.salesforce.com?so=<your-orgid>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_remedyforce_ap4_salesforce_com,
    citrixspa_routing_domain.rd_remedyforce_customer_fqdn,
  ]
}
