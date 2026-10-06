# KnowledgeOwl — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_knowledgeowl_app_knowledgeowl_com" {
  fqdn         = "app.knowledgeowl.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "KnowledgeOwl"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_knowledgeowl_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "KnowledgeOwl"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_knowledgeowl" {
  name         = "KnowledgeOwl"
  type         = "saas"
  state        = "complete"
  description  = "Knowledge base and authoring tool."
  url          = "https://app.knowledgeowl.com/app/switch-project"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "KnowledgeOwl"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.knowledgeowl.com/kb/map-saml/id/<customer_id>"
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
        name  = "SSO ID"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_knowledgeowl_app_knowledgeowl_com,
    citrixspa_routing_domain.rd_knowledgeowl_customer_fqdn,
  ]
}
