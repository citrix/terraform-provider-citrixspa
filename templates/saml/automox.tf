# Automox — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_automox_console_automox_com" {
  fqdn         = "console.automox.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Automox"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_automox_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Automox"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_automox" {
  name         = "Automox"
  type         = "saas"
  state        = "complete"
  description  = "Patch management tool to track, control, and manage the patching process."
  url          = "https://console.automox.com/dashboard?o=<customer_ID>"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAAZdEVYdFNvZnR3YXJlAEFkb2JlIEltYWdlUmVhZHlxyWU8AAADKGlUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxMTEgNzkuMTU4MzI1LCAyMDE1LzA5LzEwLTAxOjEwOjIwICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtbG5zOnhtcE1NPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvbW0vIiB4bWxuczpzdFJlZj0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL3NUeXBlL1Jlc291cmNlUmVmIyIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNSAoTWFjaW50b3NoKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpDN0Q1MEJCQ0U5Q0MxMUU1OUVGMUMyMUUxMTZBQ0FCOSIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDpDN0Q1MEJCREU5Q0MxMUU1OUVGMUMyMUUxMTZBQ0FCOSI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOkM3RDUwQkJBRTlDQzExRTU5RUYxQzIxRTExNkFDQUI5IiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOkM3RDUwQkJCRTlDQzExRTU5RUYxQzIxRTExNkFDQUI5Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpS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"

  using_template   = true
  template_name    = "Automox"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://console.automox.com/saml/acs?o=<customer_ID>"
    audience          = "https://console.automox.com/saml/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "first_name"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "last_name"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_automox_console_automox_com,
    citrixspa_routing_domain.rd_automox_customer_fqdn,
  ]
}
