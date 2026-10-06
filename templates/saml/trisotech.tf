# Trisotech — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_trisotech_your_organization_trisotech_com" {
  fqdn         = "<your-organization>.trisotech.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Trisotech"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_trisotech_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Trisotech"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_trisotech" {
  name         = "Trisotech"
  type         = "saas"
  state        = "complete"
  description  = "Tool that allows customers to discover, model, analyze their digital enterprise."
  url          = "https://<your-organization>.trisotech.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Trisotech"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.trisotech.com/sso/saml2/idp/SSOService.php"
    audience          = "https://<your-organization>.trisotech.com"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "emailaddress"
        value  = "ns_user_email"
        format = "unspecified"
      },
      {
        name   = "name"
        value  = "ns_user_name"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_trisotech_your_organization_trisotech_com,
    citrixspa_routing_domain.rd_trisotech_customer_fqdn,
  ]
}
