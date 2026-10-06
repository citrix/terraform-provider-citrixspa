# Mojo Helpdesk — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_mojo_helpdesk_customer_domain_mojohelpdesk_com" {
  fqdn         = "<Customer-domain>.mojohelpdesk.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mojo Helpdesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_mojo_helpdesk_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mojo Helpdesk"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_mojo_helpdesk" {
  name         = "Mojo Helpdesk"
  type         = "saas"
  state        = "complete"
  description  = "Help Desk Software and Ticket Tracking Knowledge..."
  url          = "https://<Customer-domain>.mojohelpdesk.com/tech"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Mojo Helpdesk"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.mojohelpdesk.com/saml/consume"
    audience          = "<Customer-domain>.mojohelpdesk.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_mojo_helpdesk_customer_domain_mojohelpdesk_com,
    citrixspa_routing_domain.rd_mojo_helpdesk_customer_fqdn,
  ]
}
