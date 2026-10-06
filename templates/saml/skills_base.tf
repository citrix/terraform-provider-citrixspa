# Skills Base — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_skills_base_app_skills_base_com" {
  fqdn         = "app.skills-base.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Skills Base"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_skills_base_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Skills Base"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_skills_base" {
  name         = "Skills Base"
  type         = "saas"
  state        = "complete"
  description  = "Talent management tool to track and document performance and skills of employee."
  url          = "https://app.skills-base.com/org/view"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Skills Base"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://saml.skills-base.com/module.php/saml/sp/saml2-acs.php/<customer_domain>"
    audience          = "skills-base.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "surname"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "givenName"
        value = "aaa.user.attribute(\"givenName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_skills_base_app_skills_base_com,
    citrixspa_routing_domain.rd_skills_base_customer_fqdn,
  ]
}
