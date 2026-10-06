# Microsoft OneDrive — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_onedrive_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft OneDrive"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_onedrive_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft OneDrive"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_onedrive" {
  name         = "Microsoft OneDrive"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://login.microsoftonline.com/login.srf?wa=wsignin1%2E0&rver=6%2E1%2E6206%2E0&wreply=https%3A%2F%2F<tenant>-my.sharepoint.com%2F&whr=<federated domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft OneDrive"
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
    citrixspa_routing_domain.rd_microsoft_onedrive_login_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_onedrive_customer_fqdn,
  ]
}
