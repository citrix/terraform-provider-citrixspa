# Keepersecurity — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_keepersecurity_keepersecurity_com" {
  fqdn         = "keepersecurity.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Keepersecurity"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_keepersecurity_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Keepersecurity"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_keepersecurity" {
  name         = "Keepersecurity"
  type         = "saas"
  state        = "complete"
  description  = "One of the best ways to keep your passwords protected and consistently safe from intruders is through the use of a secure online password manager."
  url          = "https://keepersecurity.com/en_US/console/#enterprise"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Keepersecurity"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://keepersecurity.<customer_domain>.com:8443/sso-connect/saml/sso"
    audience          = "https://keepersecurity.<customer_domain>.com:8443/sso-connect"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "transient"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_keepersecurity_keepersecurity_com,
    citrixspa_routing_domain.rd_keepersecurity_customer_fqdn,
  ]
}
