# PagerDuty — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_pagerduty_your_organization_pagerduty_com" {
  fqdn         = "<your-organization>.pagerduty.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "PagerDuty"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_pagerduty_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "PagerDuty"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_pagerduty" {
  name         = "PagerDuty"
  type         = "saas"
  state        = "complete"
  description  = "Digital operations management system."
  url          = "https://<your-organization>.pagerduty.com/sso/saml/consume"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAABmJLR0QA/wD/AP+gvaeTAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH4gQDDDcpq8LYOwAABD1JREFUeNrtm9+PE1UUx793fvTHlm633bbb7ZZmxRKfkMQQMEJU1hQjRERdBRJCgj7hiw8m/gM+mhjf5cHECCoJatAI68puglFZ8U1U0IW6v9rtL5hlu9P5eX0AZ3dsmXaJ3ZTtPU89t/fcmfnMmXPPOdOS1PiBMQCD6EzJCAAGCSEdCYBSCg4dLgwAA8AAMAAMAAPAADAADECnivDg30EOAXGDbeymttA5AJ4KPYY3Bw+DJ3ecuawu4J3JE/hrabozAHh5Nx7xJcET/g4AUQIHwmIAA8AAMAAMQEt3gaDgx2ZfEglPFF7ec8+4q5kGypqE6/IsJpdmmlo74Yki1bURYbEHIud8ilv8qbUHMBx7BvujT2KLPwUv5wZpsO3o1EBeLeNc4Uecyp7HVDVXd15Y7MGh/j3YG9mJpDcGgfAN1wYAQsjaAXgj+QqOJ4fh5sSmbUQiYMATxesbX0C/J4x3r3+EWSVvm9Mj+PHWQ0fwcmyofWPAUGgbjg08v6qL/6/sjezEwf50Xa96qW93+8YAnnAY6t1ek3eXVAmyqTja9blC4Mgy63R4Bz7NjmBWKQAAQmI30uEdNa6cV8rQqeHsXZyAiCvYegA9gh9b/ZttYxdKP+OD6c+RU0r3PgAn4MXo0zieHLYuMO6OIOmNWQDi7gg2eQdsdh/OnMXp3ChkU3U8r3Tvdry96aiVCrcMQBfvgV/osnRKKSakK/hl4Y+GtiOlSzgcfxZBsdu6a37BZ33fLfjg5lyWXjFkXChfxp9NFDTzanltYgABAbfCRSkoFnW5KVuDGqB11lv5mKz0fkopqoa6fhIh0tRmxjJBBoABYAAYAAZgXQIQiWDfCClgUtNSTVqbJbi45nI0H++16fRujtLyfgAAhF0BJDx9jnNcRMCe8OMICMs1hEo1SPqipd82lqCYmpUNbhC82B3ahoJ6E6qpORw/iF3BrbY0uGLIqBjV1gPgCIej8X14NZZuUETxiLqCtkJntprHdHXepmfkOTy6otZ4LbEf+6K7HIshNyfWFEK/Lk6iqN5aGw8IuQL3ZTdS/AlZpbhcUWoSRosTNgCEEMTcvatat2oo+K44AZVq7RsEv5gfxyfZkZrx07lRnMmN3fe6BjVxYuZLnC1cbG1HyBZwaONgQwHoVEdWKeKbwg84lT1ft4IraRLez5xEXi3jucgTiLsjEBqUuBQUVVPFb4s38FX+Ik5mz7W+JbYcuU18Xfgel6XfHefpVEdJXcCkPIO/5azj3JxawnuZj3FmfgwPdyUQFgOOTVHZUDCnFHCtMoWSJrW2H1BPLt26gs9y3/7vj0lGnkNGnmOJEAPAADAADECrpeldgILiamUKkl6xcoDbdz8/yEJS4wduNPuHiX/b2lYRo1cavrhoZ6GUZlaVB6zm11csBjAADAADwAAwAAwAA8AAMAAMQNtXg5lmurvrVDL/AOw1ZvCNklVmAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "PagerDuty"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.pagerduty.com/sso/saml/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_pagerduty_your_organization_pagerduty_com,
    citrixspa_routing_domain.rd_pagerduty_customer_fqdn,
  ]
}
