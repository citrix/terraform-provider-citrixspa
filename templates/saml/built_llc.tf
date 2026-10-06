# Built LLC — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_built_llc_flow_built_io" {
  fqdn         = "flow.built.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Built LLC"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_built_llc_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Built LLC"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_built_llc" {
  name         = "Built LLC"
  type         = "saas"
  state        = "complete"
  description  = "Connect 200 cloud apps, automate business applications and workflows by creating integrations, and boost productivity using Built.io's award-winning iPaaS"
  url          = "https://flow.built.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Built LLC"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://flow.built.io/enterprise/v1/login/sso?org=<customer_domain>"
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
    citrixspa_routing_domain.rd_built_llc_flow_built_io,
    citrixspa_routing_domain.rd_built_llc_customer_fqdn,
  ]
}
