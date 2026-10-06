# Formstack — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_formstack_customer_domain_formstack_com" {
  fqdn         = "<customer-domain>.formstack.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Formstack"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_formstack_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Formstack"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_formstack" {
  name         = "Formstack"
  type         = "saas"
  state        = "complete"
  description  = "Online Form Builder and Form Creator for Online Forms"
  url          = "https://<customer-domain>.formstack.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Formstack"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.formstack.com/admin/session/auth_provider_hook/343/process"
    audience          = "https://<Customer-domain>.formstack.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "mail"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_formstack_customer_domain_formstack_com,
    citrixspa_routing_domain.rd_formstack_customer_fqdn,
  ]
}
