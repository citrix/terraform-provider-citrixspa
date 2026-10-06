# Slack — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_slack_your_organization_slack_com" {
  fqdn         = "<your-organization>.slack.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Slack"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_slack_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Slack"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_slack" {
  name         = "Slack"
  type         = "saas"
  state        = "complete"
  description  = "Collaboration tool to communicate and share information."
  url          = "https://<your-organization>.slack.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Slack"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.slack.com/sso/saml"
    audience          = "https://slack.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "User.Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_slack_your_organization_slack_com,
    citrixspa_routing_domain.rd_slack_customer_fqdn,
  ]
}
