# Zivver — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zivver_app_zivver_com" {
  fqdn         = "app.zivver.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zivver"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zivver_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zivver"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zivver" {
  name         = "Zivver"
  type         = "saas"
  state        = "complete"
  description  = "Tool that allows secure email and file transfer from your familiar email program."
  url          = "https://app.zivver.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Zivver"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.zivver.com/api/sso/saml/consumer/"
    audience          = "https://app.zivver.com/SAML/Zivver"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "https://zivver.com/SAML/Attributes/ZivverAccountKey"
        value  = "ns_user_guid_b64"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_zivver_app_zivver_com,
    citrixspa_routing_domain.rd_zivver_customer_fqdn,
  ]
}
