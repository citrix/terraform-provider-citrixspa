# oomnitza — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_oomnitza_your_subdomain_oomnitza_com" {
  fqdn         = "<your-subdomain>.oomnitza.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "oomnitza"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_oomnitza_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "oomnitza"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_oomnitza" {
  name         = "oomnitza"
  type         = "saas"
  state        = "complete"
  description  = "IT Asset Management platform solution to track and manage assets."
  url          = "https://<your-subdomain>.oomnitza.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "oomnitza"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-subdomain>.oomnitza.com/saml/consume"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_oomnitza_your_subdomain_oomnitza_com,
    citrixspa_routing_domain.rd_oomnitza_customer_fqdn,
  ]
}
