# Bintray — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_bintray_bintray_com" {
  fqdn         = "bintray.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bintray"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_bintray_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Bintray"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_bintray" {
  name         = "Bintray"
  type         = "saas"
  state        = "complete"
  description  = "Software distribution tool to automate software distribution."
  url          = "https://bintray.com/<customer_domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAAZdEVYdFNvZnR3YXJlAEFkb2JlIEltYWdlUmVhZHlxyWU8AAADJmlUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMwNjcgNzkuMTU3NzQ3LCAyMDE1LzAzLzMwLTIzOjQwOjQyICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtbG5zOnhtcE1NPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvbW0vIiB4bWxuczpzdFJlZj0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL3NUeXBlL1Jlc291cmNlUmVmIyIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxNSAoV2luZG93cykiIHhtcE1NOkluc3RhbmNlSUQ9InhtcC5paWQ6MTgzQjVFRTQ1NDZDMTFFNTlFODhDN0VBNERBMUIzQTciIHhtcE1NOkRvY3VtZW50SUQ9InhtcC5kaWQ6MTgzQjVFRTU1NDZDMTFFNTlFODhDN0VBNERBMUIzQTciPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDoxODNCNUVFMjU0NkMxMUU1OUU4OEM3RUE0REExQjNBNyIgc3RSZWY6ZG9jdW1lbnRJRD0ieG1wLmRpZDoxODNCNUVFMzU0NkMxMUU1OUU4OEM3RUE0REExQjNBNyIvPiA8L3JkZjpEZXNjcmlwdGlvbj4gPC9yZGY6UkRG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"

  using_template   = true
  template_name    = "Bintray"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://bintray.com/api/ui/saml/login/<customer_domain>/response"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_bintray_bintray_com,
    citrixspa_routing_domain.rd_bintray_customer_fqdn,
  ]
}
