# Box — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_box_customer_domain_app_box_com" {
  fqdn         = "<customer-domain>.app.box.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Box"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_box_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Box"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_box" {
  name         = "Box"
  type         = "saas"
  state        = "complete"
  description  = "Content management and file sharing tool to manage, share and access your content."
  url          = "https://<customer-domain>.app.box.com/folder/0"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Box"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://sso.services.box.net/sp/ACS.saml2"
    audience          = "box.net"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_box_customer_domain_app_box_com,
    citrixspa_routing_domain.rd_box_customer_fqdn,
  ]
}
