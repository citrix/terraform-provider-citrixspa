# Podio — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_podio_podio_com" {
  fqdn         = "podio.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Podio"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_podio_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Podio"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_podio" {
  name         = "Podio"
  type         = "saas"
  state        = "complete"
  description  = "Web-based tool to organize team communication, business processes, data and content in project management workspaces."
  url          = "https://podio.com/login?provider=sharefile_limited"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Podio"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer>.sharefile.com/saml/acs"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "mail"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_podio_podio_com,
    citrixspa_routing_domain.rd_podio_customer_fqdn,
  ]
}
