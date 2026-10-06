# VIDIZMO — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_vidizmo_customer_domain_enterprisetube_com" {
  fqdn         = "<customer-domain>.enterprisetube.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "VIDIZMO"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_vidizmo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "VIDIZMO"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_vidizmo" {
  name         = "VIDIZMO"
  type         = "saas"
  state        = "complete"
  description  = "Enterprise live and on-demand video streaming software."
  url          = "https://<customer-domain>.enterprisetube.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "VIDIZMO"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.enterprisetube.com/Handlers/SignInHandler.ashx"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Primarysid"
        value = "ns_user_email"
      },
      {
        name  = "User.LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "User.Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_vidizmo_customer_domain_enterprisetube_com,
    citrixspa_routing_domain.rd_vidizmo_customer_fqdn,
  ]
}
