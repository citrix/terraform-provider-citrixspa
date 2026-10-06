# G Suite — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_g_suite_apps_google_com" {
  fqdn         = "apps.google.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "G Suite"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_g_suite_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "G Suite"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_g_suite" {
  name         = "G Suite"
  type         = "saas"
  state        = "complete"
  description  = "A set of intelligent apps to connect the people in your company, no matter where in the world they are."
  url          = "https://apps.google.com/user/hub"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAIAAAC1nk4lAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAQKSURBVGhD7ZVXSyRBFEb3bwrmnNeEOeeAOYcHxRwwZxHEhIrig4qKqGAWIxgx9R6tZhhmplV8KFao8zBU3e6a+ure71b/0X4hSrQslGhZKNGyUKJloUTLQomWhRItCyVaFkq0LJRoWXxL9OHh4crKytra2vPzsx6y4vT0dHFx8ezsTJ8bc39/f3t7q09+xBeix8bGQkNDg4OD/34QFRU1OzurPzOjoaHB398/MDCQdyIiIs7Pz/UHVtzd3WVmZsbHx3/neEZ8JjonJ8fDwyMoKCghISErKys2NhZlXl5eFrqvrq5cXFyqq6upxsTEBIdMT0/Xn1lBrTo6Ompqah4eHg4ODmJiYpaWlvRn38ZQdE9Pj4+PD2nr7Ozc29t7fHzc2Nhoamrq6+uzMAkiKAXviOny8jI+eXp64jVzGzAlzQzEIwb4jS0458vLy9vbGxF+j4+PceP7AmNsi+Z/U1JSKHdjY6MeMmZ8fNzb25vzWDiVeFJSkj7RtPn5+YyMDAZkobu7e3BwMC0tjdMmJyezF+dBa0lJSWJiIoUtKira3t4WC62xLXprays6OjosLIzsisj6+vrU1BTGAAarq6siLoiLi0M3xqioqDDlqaWlxc7OToxheHgYFzEoLi5G08zMDCahBxDKgXFLZGQkO7a2trIwJCQkNTX15uZGrLXAtmjcyV+Eh4dfX1+LSFlZmaOjIxanoJ6envn5+SJuAmuWl5djJ2dn5/b2diL8Ojg4iKcwOjrKcgb8VUFBAU44Ojry9fXFSwRxo5+fHwf7eFebnJx0d3cnKKYW2BbN35E8WpDFIkJimpub29raqGlAQABtJ+IWcG/k5eUh/eLioqury1z0yMiIuWiaZHd3F9GirUk2bqSDgS2AZLPdx1JLDBuRXCIuOzubA+ghTWMbqsYFwlUoIvQQKaQ7xRRwFFJoR/JkLrq+vt7NzY2BSfTOzg5vzs3NEeTSxN80Jc7EihgajK5OQ9EnJyeurq4km86ora2dnp5mVzzDrUci6VT9PU0jN7zGZmLa29tLoSk6AycnJ+HLy8tLXqPijIVoTEzQ3t6eChCkpLiuv7///S80jTuEZG9uboqpBYaigZzxNaFXyAdWBipIpk36BOSDOO9wV1Ac9s7NzaUCJEwsp+04FYkUjciUL4A4DNc/a2kePpOchLX0IgP24htknhpzPhMN3EQ0R11dXWFhYVVV1cDAAGXVn5nBN1y8w1WAEU2bLSwsEOG6wE54t7KykuDQ0BB2ItOM9/f3uQHpYLLOlCufI5WWlrKR0dUBX4g2Ifb4HD4Zr6+v+uSnUCJ9ZMx3Rf9XKNGyUKJloUTLQomWhRItCyVaFkq0LJRoWSjRslCiZaFEy0KJlsUvFK1p/wDA563lYCgIbQAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "G Suite"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.google.com/a/<customer_domain>/acs"
    audience          = "google.com/a/<customer domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_g_suite_apps_google_com,
    citrixspa_routing_domain.rd_g_suite_customer_fqdn,
  ]
}
