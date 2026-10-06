# ConnectWise Control — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_connectwise_control_customer_domain_screenconnect_com" {
  fqdn         = "<Customer-Domain>.screenconnect.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ConnectWise Control"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_connectwise_control_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ConnectWise Control"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_connectwise_control" {
  name         = "ConnectWise Control"
  type         = "saas"
  state        = "complete"
  description  = "Business management tool to provide remote support and access."
  url          = "https://<Customer-Domain>.screenconnect.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAK8SURBVHhe7Zm9axRBGMaf/fTuUngxVilT+BHEQm38+AcEEcVeEO1EwRSmMCCBNHYW/gk2FmIIUVIY0C6iNnZiYSEEVIgJmrsj++XM7mtEvYs3l5k9lnd+d+893LvHzs5zM+/uzjpPTx7LwBiXlC3WAFK2WANI2WINIGWLNYCULdYAUrZYA0jZYsaALEOWpiKSckO0q4p+A2TnM9F5oWlaYoi+56YrmqB9PUAewNbaN1x8/5Ey5fF2dgar84/h1RuU+T9GpsAgQ1EHWRSJtulLn9giSMoWMzVgYwNnX78D8iFpYDqIYuc3GgjqdUoUvJmZxuriQr6tX4wsihaFcE1oShm9xD++48idWUxev0mZgkEMUJ4CsnPFOb53iB8hHB3Fnn1jemNs/3Z4Cp3cCSUD8s7HEeLWJqLNckO2mcUxHYk+lKZA0m5h/MIlHL87R5lyeXV7Cl+eLyEVteXQ1DQOXrlGWwqMTwFZz+S5dliYaNteB5Cyhb0BSkUwbokieO48Tszdo0zBpxfLWLl6GeHeJmV2x9b6Ok4/fITxU2coU7By6wa+vlweXhHshef78IMAfhjqicCH6we0d7PomQKOI94uHFdfwKF9G8YWQVK2WANI2WINIGVLZQ2Qp91/EKdjVSppgOx8ksSQl7BRFOchSdNUbFMzoZIGeCMj+PDgPuYPT2Dx6IE8nkxO4PPSM7i1Gv2qP/QYIFeK5NOgv5bGBg+58ED77oL8lz3R0bDZ/B3iPsQVl9FDGQGJGIJxp4O43dYTnXZ+w7MTsqPdQhUtd4Nl8etu0Kv9uRy+GypZA3RiDSBli5IBssY4QTkLFd0w0bZSEdx+KlR8yXOlIau8fLmDVfteKI4AeQAuxDHAFR+lhhx94kNn5yVGHo5WCVsESdliDSBlizWAlC3WAFK2WANI2WINIGULcwOAn8vVZ3040AMHAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "ConnectWise Control"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-Domain>.screenconnect.com/__Authentication/<Customer_ID>/Login"
    audience          = "https://<Customer-Domain>.screenconnect.com/__Authentication/<Customer_ID>/Login"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Role"
        value = "ns_user_email"
      },
      {
        name  = "FirstName"
        value = "ns_user_email"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_connectwise_control_customer_domain_screenconnect_com,
    citrixspa_routing_domain.rd_connectwise_control_customer_fqdn,
  ]
}
