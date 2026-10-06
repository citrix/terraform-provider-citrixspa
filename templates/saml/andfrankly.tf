# Andfrankly — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_andfrankly_www_andfrankly_com" {
  fqdn         = "www.andfrankly.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Andfrankly"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_andfrankly_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Andfrankly"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_andfrankly" {
  name         = "Andfrankly"
  type         = "saas"
  state        = "complete"
  description  = "Engagement tool to drive change in the workplace."
  url          = "https://www.andfrankly.com/insight"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Andfrankly"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://andfrankly.com/saml/simplesaml/www/module.php/saml/sp/saml2-acs.php/<COMPANY_IDENTIFIER>"
    audience          = "https://andfrankly.com/saml/simplesaml/www/module.php/saml/sp/metadata.php/<COMPANY_IDENTIFIER>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_andfrankly_www_andfrankly_com,
    citrixspa_routing_domain.rd_andfrankly_customer_fqdn,
  ]
}
