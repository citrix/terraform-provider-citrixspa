# Microsoft Powerpoint — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_powerpoint_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Powerpoint"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_powerpoint_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Powerpoint"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_powerpoint" {
  name         = "Microsoft Powerpoint"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://login.microsoftonline.com/login.srf?wa=wsignin1%2E0&rver=6%2E1%2E6206%2E0&wreply=https%3A%2F%2Fwww.office.com%2Flaunch%2Fpowerpoint%3Fauth%3D2&whr=<federated domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft Powerpoint"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.microsoftonline.com/login.srf"
    audience          = "urn:federation:MicrosoftOnline"
    sign_assertion    = "ASSERTION"
    name_id_source    = "guid_b64"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name        = "IDPEmail"
        value       = "ns_user_email"
        format      = "unspecified"
        prefix_expr = true
      },
      {
        name   = "http://schemas.microsoft.com/ws/2008/06/identity/claims/authenticationmethod"
        value  = "http://schemas.microsoft.com/claims/multipleauthn"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_microsoft_powerpoint_login_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_powerpoint_customer_fqdn,
  ]
}
