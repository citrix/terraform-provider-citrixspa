# Microsoft OneNote — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_onenote_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft OneNote"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_onenote_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft OneNote"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_onenote" {
  name         = "Microsoft OneNote"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://login.microsoftonline.com/login.srf?wa=wsignin1%2E0&rver=6%2E1%2E6206%2E0&wreply=https%3A%2F%2Fwww.office.com%2Flaunch%2Fonenote%3Fauth%3D2&whr=<federated domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft OneNote"
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
    citrixspa_routing_domain.rd_microsoft_onenote_login_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_onenote_customer_fqdn,
  ]
}
