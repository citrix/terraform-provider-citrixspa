# Declaree — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_declaree_customer_domain_declaree_com" {
  fqdn         = "<customer-domain>.declaree.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Declaree"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_declaree_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Declaree"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_declaree" {
  name         = "Declaree"
  type         = "saas"
  state        = "complete"
  description  = "Travel and expense management tool for business travel."
  url          = "https://<customer-domain>.declaree.com/declaree/expenses"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAQTaVRYdFhNTDpjb20uYWRvYmUueG1wAAAAAAA8eDp4bXBtZXRhIHhtbG5zOng9ImFkb2JlOm5zOm1ldGEvIiB4OnhtcHRrPSJYTVAgQ29yZSA1LjQuMCI+CiAgIDxyZGY6UkRGIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyI+CiAgICAgIDxyZGY6RGVzY3JpcHRpb24gcmRmOmFib3V0PSIiCiAgICAgICAgICAgIHhtbG5zOnhtcE1NPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvbW0vIgogICAgICAgICAgICB4bWxuczpzdFJlZj0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL3NUeXBlL1Jlc291cmNlUmVmIyIKICAgICAgICAgICAgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIgogICAgICAgICAgICB4bWxuczp0aWZmPSJodHRwOi8vbnMuYWRvYmUuY29tL3RpZmYvMS4wLyI+CiAgICAgICAgIDx4bXBNTTpEZXJpdmVkRnJvbSByZGY6cGFyc2VUeXBlPSJSZXNvdXJjZSI+CiAgICAgICAgICAgIDxzdFJlZjppbnN0YW5jZUlEPnhtcC5paWQ6RjI0QzVBNkIwREI1MTFFNUFBMUFCOUU3RDVGM0I0NTA8L3N0UmVmOmluc3RhbmNlSUQ+CiAgICAgICAgICAgIDxzdFJlZjpkb2N1bWVudElEPnhtcC5kaWQ6RjI0QzVBNkMwREI1MTFFNUFBMUFCOUU3RDVGM0I0NTA8L3N0UmVmOmRvY3VtZW50SUQ+CiAgICAgICAgIDwveG1wTU06RGVyaXZlZEZyb20+CiAgICAgICAgIDx4bXBNTTpEb2N1bWVudElEPnhtcC5kaWQ6MDEyOUUwRjUwREI2MTFFNUFBMUFCOUU3RDVGM0I0NTA8L3htcE1NOkRvY3VtZW50SUQ+CiAgICAgICAgIDx4bXBNTTpJbnN0YW5jZUlEPnhtcC5paWQ6MDEyOUUwRjQwREI2MTFFNUFBMUFCOUU3RDVGM0I0NTA8L3htcE1NOkluc3RhbmNlSUQ+CiAgICAgICAgIDx4bXA6Q3JlYXRvclRvb2w+QWRvYmUgUGhvdG9zaG9wIENTNiAoTWFjaW50b3NoKTwveG1wOkNyZWF0b3JUb29sPgogICAgICAgICA8dGlmZjpPcmllbnRhdGlvbj4xPC90aWZmOk9yaWVudGF0aW9uPgogICAgICA8L3JkZjpEZXNjcmlwdGlvbj4KICAgPC9yZGY6UkRGPgo8L3g6eG1wbWV0YT4KHzNo7AAAA9lJREFUeF7tms9r1EAUx99kd6tWrfaHtqXqoS344+pNFKEg9SR4aQ8qeNCTKGLF6qEH0ZOK2B8oiL9AESttFTyKJ8GTf4AFFREqgq1VutWyTTK+lxkx+6NudjabbJn5wNs3mWQzed+ZN5nZlmUOAgeNsaTXFiOA9NpiBJBeW4wA0muLEUB6bTECSK8tRgDpI4U7aHa2xbUli3w3SMGzuhaUPkFHojJRA3zmEzDqDiaqoiJ6AWYBkhPzYNXXyhpBZj8DthILEY/JWFKAzU3Lko+YUsBMgtJrixFAem0xAkivLaEKwF20BbRfPpvPNsA6cGjp93/y7lPMsG0VQlkIeY2n0ZKo6J5eYC3tWEh558BCjb0lnoD/TkOi5xyw+kZZI8h040JoFRbwUgre6joObOPmAGLh0rGxDdzRiyjwlL+pQJQtgLeOx+ATF0Yhsa9HVCqQJcAP1PLOO7Dat4qTRXAeXAbn8QCwNbKiBErUKxsveOz91EteVvAFWaAhVRx76CQ49wcAFIIn1AWgcYM9lXq6AIz2NTFAwbsTI5gCXiIooSwATTzWsUvAalfImuCEsey3h8+AOz4CrEE9eEJtDsBvcNzPpF7zvMbd6Rlw7/UB/PyGCuUPDZ6eheT5MWDNzbJGkDcHjLwFa8dOcTIHe6gPg78OLHseVUJJAMp9VtcKqUdfZI2A6hd3oyTr8WCptMDtcOrJZ2BbcIb3EVQAe/A0DvvBUIIn1FKAXntt+TO0++K2Nxkx3OozzIxCBmQJtWbtoRPgPgsveELtSWjMJPO7mFl4Qu2ORfGCn7jp5XyYqD8uL5Q5Aacjt7Rlmz18CnM+/OCJCvVXEQqKl4sQ076BOT82HOqw9xOPAMWg2HF569zqB/d5uDmfS3UKgJOoc/UQOONXKjLs/VSlAAz3UXzyFb5qZUUFiUWAQDOA3ExWmlgEYLRFrhLUn4QVeuUF6Vu6KuDrMgLUBKDnX1wUZR/cRQECaMACChUFagLQt6anRNmHtbcX4DtqkEHDfUEhA/Li8qpASQDa//P3H+TRP1hDPSTvvgHWug1YUwdap7AN0qjcsm6J9IkH5Z/E+Bzuac4+hET3YVmD0AovQHDUYO5V/t1glKg3txoXK9eOZKsXsGerp//LEMD79RW3tnZva1XldKmUNeBYDQ7n9FewD6TA/Ti5LIUI5+8CDn7M4s06O8DqOgqwaTvmMy7oXfpfmAC3r2sCu3+XJ2jU+RGKAH/htDTAVyCQIHTXEoJha+lDlKMkVAGWIxG/dKoPI4D02mIEkF5bjADSa4sRQHptMQJIry1GAOk1BeAPTfM5G/zPR4oAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Declaree"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.declaree.com/saml/sp/acs"
    audience          = "https://saml.declaree.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_declaree_customer_domain_declaree_com,
    citrixspa_routing_domain.rd_declaree_customer_fqdn,
  ]
}
