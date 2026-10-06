# HelpDocs — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_helpdocs_customer_domain_helpdocs_io" {
  fqdn         = "<customer-domain>.helpdocs.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HelpDocs"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_helpdocs_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "HelpDocs"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_helpdocs" {
  name         = "HelpDocs"
  type         = "saas"
  state        = "complete"
  description  = "knowledge base software to guide your users when they are stuck."
  url          = "https://<customer-domain>.helpdocs.io/app/content"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "HelpDocs"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.helpdocs.io/login/saml2/post"
    audience          = "https://<customer_domain>.helpdocs.io"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_helpdocs_customer_domain_helpdocs_io,
    citrixspa_routing_domain.rd_helpdocs_customer_fqdn,
  ]
}
