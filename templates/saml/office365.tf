# Office365 — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_office365_login_microsoftonline_com" {
  fqdn         = "login.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_office_com" {
  fqdn         = "*.office.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_office365_com" {
  fqdn         = "*.office365.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_sharepoint_com" {
  fqdn         = "*.sharepoint.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_live_com" {
  fqdn         = "*.live.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_onenote_com" {
  fqdn         = "*.onenote.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_microsoft_com" {
  fqdn         = "*.microsoft.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_powerbi_com" {
  fqdn         = "*.powerbi.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_dynamics_com" {
  fqdn         = "*.dynamics.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_microsoftstream_com" {
  fqdn         = "*.microsoftstream.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_powerapps_com" {
  fqdn         = "*.powerapps.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_yammer_com" {
  fqdn         = "*.yammer.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_windowsazure_com" {
  fqdn         = "*.windowsazure.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_msauth_net" {
  fqdn         = "*.msauth.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_msauthimages_net" {
  fqdn         = "*.msauthimages.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_microsoftonline_com" {
  fqdn         = "*.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_windows_net" {
  fqdn         = "*.windows.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_microsoftonline_p_com" {
  fqdn         = "*.microsoftonline-p.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_akamaihd_net" {
  fqdn         = "*.akamaihd.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_sharepointonline_com" {
  fqdn         = "*.sharepointonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_live_net" {
  fqdn         = "*.live.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_office_net" {
  fqdn         = "*.office.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_office365_msftauth_net" {
  fqdn         = "*.msftauth.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Office365"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_office365" {
  name         = "Office365"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://login.microsoftonline.com/login.srf"
  related_urls = ["<Customer FQDN>", "*.office.com", "*.office365.com", "*.sharepoint.com", "*.live.com", "*.onenote.com", "*.microsoft.com", "*.powerbi.com", "*.dynamics.com", "*.microsoftstream.com", "*.powerapps.com", "*.yammer.com", "*.windowsazure.com", "*.msauth.net", "*.msauthimages.net", "*.microsoftonline.com", "*.windows.net", "*.microsoftonline-p.com", "*.akamaihd.net", "*.sharepointonline.com", "*.live.net", "*.office.net", "*.msftauth.net"]
  icon         = "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"

  using_template   = true
  template_name    = "Office365"
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
    saml_type         = "SP_IDP"
    sp_initiated_only = false

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
    citrixspa_routing_domain.rd_office365_login_microsoftonline_com,
    citrixspa_routing_domain.rd_office365_customer_fqdn,
    citrixspa_routing_domain.rd_office365_office_com,
    citrixspa_routing_domain.rd_office365_office365_com,
    citrixspa_routing_domain.rd_office365_sharepoint_com,
    citrixspa_routing_domain.rd_office365_live_com,
    citrixspa_routing_domain.rd_office365_onenote_com,
    citrixspa_routing_domain.rd_office365_microsoft_com,
    citrixspa_routing_domain.rd_office365_powerbi_com,
    citrixspa_routing_domain.rd_office365_dynamics_com,
    citrixspa_routing_domain.rd_office365_microsoftstream_com,
    citrixspa_routing_domain.rd_office365_powerapps_com,
    citrixspa_routing_domain.rd_office365_yammer_com,
    citrixspa_routing_domain.rd_office365_windowsazure_com,
    citrixspa_routing_domain.rd_office365_msauth_net,
    citrixspa_routing_domain.rd_office365_msauthimages_net,
    citrixspa_routing_domain.rd_office365_microsoftonline_com,
    citrixspa_routing_domain.rd_office365_windows_net,
    citrixspa_routing_domain.rd_office365_microsoftonline_p_com,
    citrixspa_routing_domain.rd_office365_akamaihd_net,
    citrixspa_routing_domain.rd_office365_sharepointonline_com,
    citrixspa_routing_domain.rd_office365_live_net,
    citrixspa_routing_domain.rd_office365_office_net,
    citrixspa_routing_domain.rd_office365_msftauth_net,
  ]
}
