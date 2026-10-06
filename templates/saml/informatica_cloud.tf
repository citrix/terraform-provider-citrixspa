# Informatica Cloud — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_informatica_cloud_apse1_dm_ap_informaticacloud_com" {
  fqdn         = "apse1.dm-ap.informaticacloud.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Informatica Cloud"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_informatica_cloud_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Informatica Cloud"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_informatica_cloud" {
  name         = "Informatica Cloud"
  type         = "saas"
  state        = "complete"
  description  = "Informatica Cloud delivers enterprise-class software-as-a-service SaaS integration applications and a powerful platform for developing"
  url          = "https://apse1.dm-ap.informaticacloud.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Informatica Cloud"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://dm-ap.informaticacloud.com/identity-service/slo/<Customer-id>"
    audience          = "https://<Customer-id>.dm-ap.informaticacloud.com"
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
    citrixspa_routing_domain.rd_informatica_cloud_apse1_dm_ap_informaticacloud_com,
    citrixspa_routing_domain.rd_informatica_cloud_customer_fqdn,
  ]
}
