# QuestionPro — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_questionpro_www_questionpro_com" {
  fqdn         = "www.questionpro.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "QuestionPro"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_questionpro_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "QuestionPro"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_questionpro" {
  name         = "QuestionPro"
  type         = "saas"
  state        = "complete"
  description  = "Online survey software to create surveys and questionnaires."
  url          = "https://www.questionpro.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "QuestionPro"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.questionpro.com/a/saml2/<Customer-Id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "emailAddress"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_questionpro_www_questionpro_com,
    citrixspa_routing_domain.rd_questionpro_customer_fqdn,
  ]
}
