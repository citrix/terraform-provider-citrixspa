# Flatter Files — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_flatter_files_www_flatterfiles_com" {
  fqdn         = "www.flatterfiles.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Flatter Files"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_flatter_files_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Flatter Files"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_flatter_files" {
  name         = "Flatter Files"
  type         = "saas"
  state        = "complete"
  description  = "Digital flat file cabinet for drawings and documents to provide a secure and simple way for providing access to content."
  url          = "https://www.flatterfiles.com/a/app"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Flatter Files"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.flatterfiles.com/saml"
    audience          = "https://www.flatterfiles.com/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"FirstName\")"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"LastName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_flatter_files_www_flatterfiles_com,
    citrixspa_routing_domain.rd_flatter_files_customer_fqdn,
  ]
}
