# GitBook — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_gitbook_www_gitbook_com" {
  fqdn         = "www.gitbook.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GitBook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_gitbook_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "GitBook"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_gitbook" {
  name         = "GitBook"
  type         = "saas"
  state        = "complete"
  description  = "Tool to create and maintain your documentation."
  url          = "https://www.gitbook.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "GitBook"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.gitbook.com/saml/<customer_id>"
    audience          = "https://www.gitbook.com/saml/<customer_id>"
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
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_gitbook_www_gitbook_com,
    citrixspa_routing_domain.rd_gitbook_customer_fqdn,
  ]
}
