# Lifesize — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_lifesize_manage_lifesizecloud_com" {
  fqdn         = "manage.lifesizecloud.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Lifesize"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_lifesize_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Lifesize"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_lifesize" {
  name         = "Lifesize"
  type         = "saas"
  state        = "complete"
  description  = "Video conferencing solution."
  url          = "https://manage.lifesizecloud.com/#/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Lifesize"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.lifesizecloud.com/ls/?acs"
    audience          = "https://login.lifesizecloud.com/ls/metadata/"
    relay_state       = "https://webapp.lifesizecloud.com/?ent=<entity_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstname"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastname"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_lifesize_manage_lifesizecloud_com,
    citrixspa_routing_domain.rd_lifesize_customer_fqdn,
  ]
}
