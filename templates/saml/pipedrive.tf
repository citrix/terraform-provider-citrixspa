# Pipedrive — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_pipedrive_customer_domain_pipedrive_com" {
  fqdn         = "<customer-domain>.pipedrive.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pipedrive"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_pipedrive_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Pipedrive"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_pipedrive" {
  name         = "Pipedrive"
  type         = "saas"
  state        = "complete"
  description  = "Sales CRM and pipeline management software."
  url          = "https://<customer-domain>.pipedrive.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Pipedrive"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.pipedrive.com/sso/auth/samlp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "email"
        value  = "ns_user_email"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_pipedrive_customer_domain_pipedrive_com,
    citrixspa_routing_domain.rd_pipedrive_customer_fqdn,
  ]
}
