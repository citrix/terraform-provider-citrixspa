# Aha — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_aha_your_organization_aha_io" {
  fqdn         = "<your-organization>.aha.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Aha"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_aha_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Aha"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_aha" {
  name         = "Aha"
  type         = "saas"
  state        = "complete"
  description  = "Product roadmap and marketing planning tool to build products and launch campaigns."
  url          = "https://<your-organization>.aha.io"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Aha"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.aha.io/auth/saml/callback"
    audience          = "https://<your-organization>.aha.io/"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "EmailAddress"
        value = "ns_user_email"
      },
      {
        name  = "FirstName"
        value = "aaa.USER.ATTRIBUTE(\"givenname\")"
      },
      {
        name  = "LastName"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_aha_your_organization_aha_io,
    citrixspa_routing_domain.rd_aha_customer_fqdn,
  ]
}
