# Velpic — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_velpic_customer_domain_velpic_net" {
  fqdn         = "<customer-domain>.velpic.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Velpic"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_velpic_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Velpic"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_velpic" {
  name         = "Velpic"
  type         = "saas"
  state        = "complete"
  description  = "Learning management system LMS designed to streamline workplace training."
  url          = "https://<customer-domain>.velpic.net/#player"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Velpic"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://auth.velpic.com/saml/v2/<customer_id>/login"
    audience          = "https://auth.velpic.com/saml/v2/<customer_id>/login"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
        value = "ns_user_email"
      },
      {
        name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_velpic_customer_domain_velpic_net,
    citrixspa_routing_domain.rd_velpic_customer_fqdn,
  ]
}
