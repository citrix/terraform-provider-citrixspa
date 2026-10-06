# Docusign — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_docusign_account_d_docusign_com" {
  fqdn         = "account-d.docusign.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Docusign"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_docusign_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Docusign"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_docusign" {
  name         = "Docusign"
  type         = "saas"
  state        = "complete"
  description  = "Online signature tool for different documents, such as insurance, medical, and real estate."
  url          = "https://account-d.docusign.com/organizations/<your-org-id>"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Docusign"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://account-d.docusign.com/organizations/<your-org-id>/saml2/login"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "givenname"
        value = "ns_user_name"
      },
      {
        name  = "User.LastName"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
      {
        name  = "User.email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_docusign_account_d_docusign_com,
    citrixspa_routing_domain.rd_docusign_customer_fqdn,
  ]
}
