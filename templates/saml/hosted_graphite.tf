# Hosted Graphite — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_hosted_graphite_www_hostedgraphite_com" {
  fqdn         = "www.hostedgraphite.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Hosted Graphite"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_hosted_graphite_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Hosted Graphite"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_hosted_graphite" {
  name         = "Hosted Graphite"
  type         = "saas"
  state        = "complete"
  description  = "Graphite Monitoring, Grafana Dashboards and..."
  url          = "https://www.hostedgraphite.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Hosted Graphite"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.hostedgraphite.com/complete/saml/<Customer-id>/"
    audience          = "https://www.hostedgraphite.com/metadata/<Customer-id>/"
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
    citrixspa_routing_domain.rd_hosted_graphite_www_hostedgraphite_com,
    citrixspa_routing_domain.rd_hosted_graphite_customer_fqdn,
  ]
}
