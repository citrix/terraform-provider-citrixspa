# Microsoft Teams — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_teams_teams_microsoft_com" {
  fqdn         = "teams.microsoft.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_teams_microsoft_com_2" {
  fqdn         = "*.teams.microsoft.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_office_com" {
  fqdn         = "*.office.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_office365_com" {
  fqdn         = "*.office365.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_microsoft_com" {
  fqdn         = "*.microsoft.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_microsoftonline_com" {
  fqdn         = "*.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_sharepoint_com" {
  fqdn         = "*.sharepoint.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_skype_com" {
  fqdn         = "*.skype.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_lync_com" {
  fqdn         = "*.lync.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_msedge_net" {
  fqdn         = "*.msedge.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_akamaized_net" {
  fqdn         = "*.akamaized.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_trafficmanager_net" {
  fqdn         = "*.trafficmanager.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_cloud_microsoft" {
  fqdn         = "*.cloud.microsoft"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_media_azure_net" {
  fqdn         = "*.media.azure.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_teams_streaming_mediaservices_windows_net" {
  fqdn         = "*.streaming.mediaservices.windows.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Teams"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_teams" {
  name         = "Microsoft Teams"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://teams.microsoft.com"
  related_urls = ["<Customer FQDN>", "*.teams.microsoft.com", "*.office.com", "*.office365.com", "*.microsoft.com", "*.microsoftonline.com", "*.sharepoint.com", "*.skype.com", "*.lync.com", "*.msedge.net", "*.akamaized.net", "*.trafficmanager.net", "*.cloud.microsoft", "*.media.azure.net", "*.streaming.mediaservices.windows.net"]
  icon         = "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"

  using_template   = true
  template_name    = "Microsoft Teams"
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
    citrixspa_routing_domain.rd_microsoft_teams_teams_microsoft_com,
    citrixspa_routing_domain.rd_microsoft_teams_customer_fqdn,
    citrixspa_routing_domain.rd_microsoft_teams_teams_microsoft_com_2,
    citrixspa_routing_domain.rd_microsoft_teams_office_com,
    citrixspa_routing_domain.rd_microsoft_teams_office365_com,
    citrixspa_routing_domain.rd_microsoft_teams_microsoft_com,
    citrixspa_routing_domain.rd_microsoft_teams_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_teams_sharepoint_com,
    citrixspa_routing_domain.rd_microsoft_teams_skype_com,
    citrixspa_routing_domain.rd_microsoft_teams_lync_com,
    citrixspa_routing_domain.rd_microsoft_teams_msedge_net,
    citrixspa_routing_domain.rd_microsoft_teams_akamaized_net,
    citrixspa_routing_domain.rd_microsoft_teams_trafficmanager_net,
    citrixspa_routing_domain.rd_microsoft_teams_cloud_microsoft,
    citrixspa_routing_domain.rd_microsoft_teams_media_azure_net,
    citrixspa_routing_domain.rd_microsoft_teams_streaming_mediaservices_windows_net,
  ]
}
