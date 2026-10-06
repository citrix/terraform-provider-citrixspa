# Proto.io — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_proto_io_proto_io" {
  fqdn         = "proto.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Proto.io"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_proto_io_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Proto.io"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_proto_io" {
  name         = "Proto.io"
  type         = "saas"
  state        = "complete"
  description  = "Application prototyping platform to create fully-interactive, high-fidelity prototypes."
  url          = "https://proto.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "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"

  using_template   = true
  template_name    = "Proto.io"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://saml.proto.io/login/callback?code=<your account sub-domain>"
    audience          = "https://saml.proto.io"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name   = "fname"
        value  = "ns_user_name"
        format = "basic"
      },
      {
        name   = "lname"
        value  = "aaa.USER.ATTRIBUTE(\"sn\")"
        format = "basic"
      },
      {
        name   = "email"
        value  = "ns_user_email"
        format = "basic"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_proto_io_proto_io,
    citrixspa_routing_domain.rd_proto_io_customer_fqdn,
  ]
}
