# iMeet Central — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_imeet_central_company_domain_imeetcentral_com" {
  fqdn         = "<company-domain>.imeetcentral.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "iMeet Central"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_imeet_central_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "iMeet Central"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_imeet_central" {
  name         = "iMeet Central"
  type         = "saas"
  state        = "complete"
  description  = "Project management software for marketers, creative agencies, and enterprise businesses."
  url          = "https://<company-domain>.imeetcentral.com/home/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "iMeet Central"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.imeetcentral.com/saml2-assertion.php"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Firstname"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Lastname"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_imeet_central_company_domain_imeetcentral_com,
    citrixspa_routing_domain.rd_imeet_central_customer_fqdn,
  ]
}
