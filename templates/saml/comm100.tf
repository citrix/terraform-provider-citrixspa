# Comm100 — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_comm100_hosted_comm100_com" {
  fqdn         = "hosted.comm100.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Comm100"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_comm100_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Comm100"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_comm100" {
  name         = "Comm100"
  type         = "saas"
  state        = "complete"
  description  = "Comm100 Omnichannel Customer Experience Platform Live Chat. Comm100 is an award-winning omnichannel customer experience platform encompassing live chat, ticketing, KB, chatbot, and social media."
  url          = "https://hosted.comm100.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Comm100"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://hosted.comm100.com/"
    audience          = "https://hosted.comm100.com/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_comm100_hosted_comm100_com,
    citrixspa_routing_domain.rd_comm100_customer_fqdn,
  ]
}
