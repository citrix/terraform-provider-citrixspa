# EduBrite — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_edubrite_customer_domain_edubrite_com" {
  fqdn         = "<Customer-domain>.edubrite.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EduBrite"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_edubrite_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "EduBrite"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_edubrite" {
  name         = "EduBrite"
  type         = "saas"
  state        = "complete"
  description  = "Learning management tool to create, deliver, and track training programs."
  url          = "https://<Customer-domain>.edubrite.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "EduBrite"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.edubrite.com/oltpublish/site/samlLoginResponse.do"
    audience          = "https://<Customer-domain>.edubrite.com"
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
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_edubrite_customer_domain_edubrite_com,
    citrixspa_routing_domain.rd_edubrite_customer_fqdn,
  ]
}
