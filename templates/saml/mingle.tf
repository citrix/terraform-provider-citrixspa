# Mingle — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_mingle_customer_domain_mingle_thoughtworks_com" {
  fqdn         = "<customer-domain>.mingle.thoughtworks.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mingle"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_mingle_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mingle"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_mingle" {
  name         = "Mingle"
  type         = "saas"
  state        = "complete"
  description  = "Agile project management and collaboration tool to provide a combined workplace for the entire team."
  url          = "https://<customer-domain>.mingle.thoughtworks.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAIAAAC1nk4lAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAPNSURBVGhD7ZVLKG1RGMclQ0yIjvezPIooj4SkZIDymkhihJCQEKHIALfk5pFMJDlIiIkI5REpeU4oj5DX0ICSuP+9v2V1uPecs6O7pNZvsPt/37fWXv/12GtbvP5ApGlRSNOikKZFIU2LQpoWhTQtCmlaFNK0KKRpUUjTovjJpp+enlpaWtrb23+ZpLm5+fb2lrpo4eXlhSnNTE1Nubi4sMAIzHRcXJyFNpydnamLWTY3N9Eey8Fibej1evRigRFYOTo6WrVkHicnJ+piloeHh7m5ORZoZnR0FKOwwAisHBMTg6axsbEU/pO8vDy04XvX1NS0vr6emprq5uZWUVGBzODgoIeHR2Bg4OrqKkIcpMLCQoiGhgZksrOz3d3dU1JSlM5v1NTUeHp6BgUFnZycDA8PT09Pz8zMcNOPj49paWnolZCQcHd3R0nwznRYWBh0UlKSra0tOlCJk56ebmjay8sLYXFx8dDQEAR2AE/okpISiOPj48PDQwi01Ol0EFlZWSMjI66urpQEVlZWlpaWfX199fX1SGLQsrKy+fl5anB+fg4RHx8/MTGRkZEBvb29TR3fmY6MjFRSKtgmKnGoJzft7e2NaZDe2dlB6ebmhkKsXFFR0dXVFZIIMR84phJAcm9vD8tPVeLy8hJhXV0dN405VFdXUxV0dnby9kZNT05OUonzwTTWrLW1lTQNSRqgZU5ODuZASXt7++7ubioBJJeXl/GetrY2llLB1lVWVnLTeAYHBwcEBPj5+fn7+0Mjg1OklNT2nzSNG5D02dkZSqQBDmJubq6h6Y6ODioBJNfW1uzs7Hp6elhKBf4+mO7q6sJsf6v09/fjm8HqKCW1vSbTH870V0wvLCzg/OAAsJQK8rW1tYam8d+gEri4uMCqkTZqemVlhUoc+ECem3ZwcGhsbCR9enqKEmmATxmHmJ9pGxsbw+GRHB8fJxESErK7uzs7OwsNYJo0qhsbGxDl5eUHBwdjY2PQBQUF6gveTNM9HR4erqRMwn8umZmZAwMDpOHPx8eHNMDY+L/ikvL19UWYnJyMUakEIiIilpaWSCcmJuKduEOwwFFRUaWlpVtbWxBUPTo6wgWKBtbW1r29vZQE70yHhoYqKZM4OjpSly+C7X5+fmaBCl6OQ8wCkzDT+Nlilvf399CLi4tVVVW4O/8GH8r19TV1+SLYDbjEZUxhfn4+QtJm0druf4BrUdm7N/AzYgVzfKdpgB3G9be/v89ibXyz6c8hTYtCmhaFNC0KaVoU0rQopGlRSNOikKZFIU2LQpoWhTQthtfXP3ec67uDTKQDAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Mingle"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://profile.thoughtworks.com/saml/consume?RelayState=<customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_mingle_customer_domain_mingle_thoughtworks_com,
    citrixspa_routing_domain.rd_mingle_customer_fqdn,
  ]
}
