# Reamaze — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_reamaze_customer_domain_reamaze_com" {
  fqdn         = "<Customer-domain>.reamaze.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Reamaze"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_reamaze_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Reamaze"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_reamaze" {
  name         = "Reamaze"
  type         = "saas"
  state        = "complete"
  description  = "Customer support software to support, engage, and convert customers with chat, social, SMS, FAQ, and email on a single platform."
  url          = "https://<Customer-domain>.reamaze.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Reamaze"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.reamaze.com/admin/auth/saml_consume"
    audience          = "https://<Customer-domain>.reamaze.com/admin/auth/saml_init"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_reamaze_customer_domain_reamaze_com,
    citrixspa_routing_domain.rd_reamaze_customer_fqdn,
  ]
}
