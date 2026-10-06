# Twentythree — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_twentythree_customer_domain_videomarketingplatform_co" {
  fqdn         = "<customer-domain>.videomarketingplatform.co"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Twentythree"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_twentythree_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Twentythree"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_twentythree" {
  name         = "Twentythree"
  type         = "saas"
  state        = "complete"
  description  = "Video marketing platform to integrate and add videos to the marketing stack."
  url          = "http://<customer-domain>.videomarketingplatform.co/manage/videos"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Twentythree"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.videomarketingplatform.co/saml/login"
    audience          = "http://www.23video.com/saml/trust/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_twentythree_customer_domain_videomarketingplatform_co,
    citrixspa_routing_domain.rd_twentythree_customer_fqdn,
  ]
}
