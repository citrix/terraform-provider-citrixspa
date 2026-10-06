# Microsoft Excel — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_excel_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Excel"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_excel_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Excel"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_excel" {
  name         = "Microsoft Excel"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://login.microsoftonline.com/login.srf?wa=wsignin1%2E0&rver=6%2E1%2E6206%2E0&wreply=https%3A%2F%2Fwww.office.com%2Flaunch%2FExcel%3Fauth%3D2&whr=<federated domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft Excel"
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
    citrixspa_routing_domain.rd_microsoft_excel_login_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_excel_customer_fqdn,
  ]
}
