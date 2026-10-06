# shufflrr — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_shufflrr_your_organization_shufflrr_com" {
  fqdn         = "<your-organization>.shufflrr.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "shufflrr"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_shufflrr_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "shufflrr"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_shufflrr" {
  name         = "shufflrr"
  type         = "saas"
  state        = "complete"
  description  = "Presentation management tool to create, update, share, and broadcast presentations."
  url          = "https://<your-organization>.shufflrr.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "shufflrr"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.shufflrr.com/login/samlassertionconsumerservice"
    audience          = "https://<your-organization>.shufflrr.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "firstName"
        value  = "ns_user_name"
        format = "unspecified"
      },
      {
        name   = "lastName"
        value  = "aaa.USER.ATTRIBUTE(\"sn\")"
        format = "unspecified"
      },
      {
        name   = "fullName"
        value  = "ns_user_name"
        format = "unspecified"
      },
      {
        name   = "email"
        value  = "ns_user_email"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_shufflrr_your_organization_shufflrr_com,
    citrixspa_routing_domain.rd_shufflrr_customer_fqdn,
  ]
}
