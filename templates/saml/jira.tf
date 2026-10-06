# Jira — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_jira_your_organization_atlassian_net" {
  fqdn         = "<your-organization>.atlassian.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Jira"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_jira_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Jira"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_jira" {
  name         = "Jira"
  type         = "saas"
  state        = "complete"
  description  = "Tool to plan, track, and manage your issues and projects."
  url          = "https://<your-organization>.atlassian.net"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Jira"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://auth.atlassian.com/login/callback?connection=saml-<your-org-id>"
    audience          = "https://auth.atlassian.com/saml/<your-org-id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "givenname"
        value = "aaa.USER.ATTRIBUTE(\"givenName\")"
      },
      {
        name  = "surname"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_jira_your_organization_atlassian_net,
    citrixspa_routing_domain.rd_jira_customer_fqdn,
  ]
}
