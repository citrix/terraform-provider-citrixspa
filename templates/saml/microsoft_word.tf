# Microsoft Word — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_word_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Word"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_word_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Word"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_word" {
  name         = "Microsoft Word"
  type         = "saas"
  state        = "complete"
  description  = "Microsoft Word cloud-based subscription service."
  url          = "https://login.microsoftonline.com/login.srf?wa=wsignin1%2E0&rver=6%2E1%2E6206%2E0&wreply=https%3A%2F%2Fwww.office.com%2Flaunch%2FWord%3Fauth%3D2&whr=<federated domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft Word"
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
    citrixspa_routing_domain.rd_microsoft_word_login_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_word_customer_fqdn,
  ]
}
