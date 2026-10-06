# TimeLive — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_timelive_timelive_livetecs_com" {
  fqdn         = "timelive.livetecs.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "TimeLive"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_timelive_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "TimeLive"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_timelive" {
  name         = "TimeLive"
  type         = "saas"
  state        = "complete"
  description  = "Tool to provide timesheets and track time."
  url          = "https://timelive.livetecs.com/Employee/Default.aspx"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAATkSURBVGhD7ZRrTFNXHMBvSx9AgRaKSkZC1DijwuYWIsY5XByCIzBlgyVY3SN78N4I2dARk21EATNBdFE3E9zc4pRHaWkpbSm0KBbkVVqhViugQ1hb5VUnlIeOu3+5d5R9YNkHkwPL/eWk+T/OuZzfPfeA4cscSgA1lABqKAHUUAKooQRQQwmghhJADSWAGkoANZQAaigB1PyvBfr77pDREmZRgcMHszfFZ5///gyZL1UWFSg88tWpY0cH7vWS+VLFKXDT8jg0TxtxojWi2Dm2HWsW660FasvWQn23xZEtvLWzqCXqZFtkcWvkyTaITzf8BqtCj2ohhfFaYcuF5sG5p5EUKHph8vwDt+Q3mcdwQcLboS+9PDgwQE7C8QaNJmTDxsM5OYoa+bYtYdGRUWTjb97asxfqv1z4GeKKsvK42DcjX4/YHbELxo7tr54oKoK6U6C5bwwTVHlk1AZ8ofbLqsPelZxS30s410lLltffGg7N1/Kz6thpSm6myiNDCQH8xp3VMVIUnE9r3dOVrDQFPVmeW93j/Js4HpCtpqfIsU9qYDBTFc40Wd76AN+wOoiJ0Qx6PTENuHTxVyaNHhsdDTGGYWw3RpNWS7SA9rY2Ft0N6hDH741jYDQGhkEOQeCqAJ6Xd0pSErRcn5DGPAybi/6unUhjTndA2mAe2f5tMwQmyx9E/afmAZ9MFWxxbGKGqIg6rd6fqWC7EAtKDIxURfqlm0Srf8QBhoEH1RCvX7OWw3bv7uoiWkB5aRlU9sTEQCxITOT7cGPfcMoQwIH4c3kH9gk0ajWYvxK2dWpqimit4vuDxujoKMQuAbHBChvdVdxKpAsFWGnKzvt2og7wMlXsdOXdoQkyx3HYNGx0YvopBEFfasgqju8obME+lunvP4L4+X8VMJlMnkwWHIjD4XwsbI5Fo3uy2FarNTpqNwhMT0/PLcKLjhdCmvv1N0T6XwV0/S4BvyynQN8/BTwzai32KSxJHv+DjiiK9TZsf9U75zqJdN3qNbBdY3c3kQILBYAXNm7y8+HClYA4IzUNXj9UIN4cHOLuxpibgk+Mj8NHFcBfQaTAMxOAE8Bn8ZWf18MNIYoSg+281nVl169d5+Xu0bXICQClly/zOF5+3j4QezLZXI6XSFgJ8f7EfXAauo4OiG8YDD+WlNhsNueCOZ6lwOPJJ2eu9MMhbD5yrc401GAebrwzojQ+FOltsziemJDgz/N9MThELBIJyysar16VSqQLBQDYdNBzgXBH4ZfP5RHF3p4eeOsrffmiysoGteZaY2N9XZ1UIjEajdB1CZS2/459UB2W30Sk4cdbIFUYh4JzG7H3pdf7nDeGAEuqgZbZNk7mUHlPApX+kUmIUy52Yx/KsI8WjESxyDA0Nf7Ig8nisNg02A6GrfDzVSlVEOwMDyceAuQcOgTXgMfxhjtakJdHVnG8Uiikz61ayAGBAFougcGxyQqd9Yp5hEiv3x0r77DYHTPwT7as3fLnLLxEEpVpCFpPnroqCuNDqMzPGR6flt54IOsih1BnhecQLYVcLquulkmr4V3a7XY4itYW8swBh8MBn02NTAY7npkhl8yjVChgLXRhVInF5tu3oegSWKZQAqihBFBDCaCGEkANJYAaSgA1lABqKAHUUAKooQRQQwmghhJADSWAmmUugON/AUGlgaTKQBBlAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "TimeLive"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.livetecs.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_timelive_timelive_livetecs_com,
    citrixspa_routing_domain.rd_timelive_customer_fqdn,
  ]
}
