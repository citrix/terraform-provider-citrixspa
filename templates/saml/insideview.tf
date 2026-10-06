# InsideView — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_insideview_my_insideview_com" {
  fqdn         = "my.insideview.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "InsideView"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_insideview_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "InsideView"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_insideview" {
  name         = "InsideView"
  type         = "saas"
  state        = "complete"
  description  = "Data and intelligence solutions to solve sales, marketing, and other business challenges."
  url          = "https://my.insideview.com/iv/home.do?methodName=showHomePage"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "InsideView"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://my.insideview.com/iv/<customer_domain>/login.iv"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
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
    citrixspa_routing_domain.rd_insideview_my_insideview_com,
    citrixspa_routing_domain.rd_insideview_customer_fqdn,
  ]
}
