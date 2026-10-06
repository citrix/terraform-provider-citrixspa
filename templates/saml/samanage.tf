# Samanage — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_samanage_customer_domain_samanage_com" {
  fqdn         = "<customer-domain>.samanage.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Samanage"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_samanage_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Samanage"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_samanage" {
  name         = "Samanage"
  type         = "saas"
  state        = "complete"
  description  = "Tool for IT service management."
  url          = "https://<customer-domain>.samanage.com/welcome"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAYAAAA6/NlyAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsIAAA7CARUoSoAAAAASdEVYdFNvZnR3YXJlAEdyZWVuc2hvdF5VCAUAAAWJSURBVGhD7Vl9TJVVGP9d4F4+L8JFBC6kImIgFxeimBJpW5Qof2BShrnFiFZTqw1ctdiqP9Q10ZK2olpzNDIza6BZunBqI1iAQWFACIopX2IXgcun98Ltec57YUQvzLba8uX9be/e5z3POc85v+fr3IEGz52wYxbByfGeNVAJKx0qYaVDJax0qISVDpWw0qESVjpUwkqHSljpUAkrHSphpUMlrHTcAWH+s/VMD0NOnvyMQ04n9zDkxhhTx+We6TH9H+LtNDxiA6xj0rfWGXAh/wzT2BjpNBrAjcac6D1EYwwdfY+MSm7UuUhzSQ0PLWAjO7dJx2t5jRvpeWfeg20502MlvTMtdicd2+G9+RyscycbbIv3Ylu8hr/ZzqjDtp0G+ExassF6GcgT5k0GrNiZvATPxM8ne3YcKr+Ok7U3kLfFhDB/D3RZRpB1rB6dfSM4nBEDM83fe6oJ76WZ0HRjAK8W/YbPMmMwQofemFuGR1cYkZ0YhkC9K66YB7HtUDXuDzUgK3ERLrZaUHbZjJwN4ahs6cEL+VV4KXUpUpcHQU+Eqq/1IuNQjXDgnqeWITU2CA0d/RQDDTILa2EK8sLeTRF0bA3t24CzNR2S02RIyxMmbyXHGvH1jjg8W/gLwud5Yr7BHRVXe7CLDvjOmSvYnxolpmqeLkJ3fjJ8KYo/NJthMnrDh+Sfr/dirpcOIb7ueHB/GV5LCqcgOROxbiHXtvbigdxy9OUlCTun67qQGOlPAdZAk14Me0EKPiz9HXpXZ2yNC8Fbpy/B3G9DLjki4o1zOJ+1GoFz3JBeUIOC9Bi8fqIRC/zcKEALYHy5BB3dg1JGTsHfRxgUYfYeY99jkTD6uOHFo3U4SBHN/rJeZGVn77DQo2cY39V3CTHhlTPIO3tFyPH7ykQGMLwpSkl7SnGkshU3KCMYBk8dLA03RYYwknZ8Q1l0Tcg6Ty2020+isbMflyhbGF6uWoqDVF59Q1bhmF/b+7AxOkCMBXjrME/vJuS0OKOU9jKQJ0y1WHS2BZs/qML3Td3YtioEl3c/hOM5Cfg8MxZ1HRbODQnMfhyuLuQoyaQn1bCrizSJSdk+3YS8J02ob6e1BOsorSPHTiQdpSCTECJlgvX9ZKStNKLlD4oUQUPp+S6dyTJiRf7WaBRWtCJ65ykEersK/blGMz6ijFh3oBxHq9qlOpaB/Cg1jF1bopBP9XKgpJl6x5io0dC5HkJtI+8FekvejF4ZjDVhBiGbIvyQsFiS191rwNolfkJOuS+IyDihd9AGd520JdsKI0cGOA7st9Qfq0N9hZwRf49499B8f71OyAmLffEElZmeIv0wpf72tQtx4eB6fFwmZcXjVNc+1NjOZ6+hSJPNyYGYBGesSHvTIf8F/UQqMtCLNglGKUX5kbwKHPupHcuC9TAFeyOnuAGhfh5o6xkS6VlHkTP3W7GAiNRTQ7k1aIXBQ4er3UPUuEbx/OFarFrkAyeK4pHKNsyhOmdCHGhOXZ6/kNZebLNQqlrwVXUHEsINoubbbg1RhtoRQee5aRnG22daxPwNpgDs/raZmmkn1kfNE/O/uNCGT35slRrWHTctBtcAXyvcsXkhdz2eydcCB4lTl68NTp3xeuGri68WBuvZy7yeU5UbCF9BPM72bDSP50+sddhjiGZD66z08DjXbt9t2I+niRujgG6MzcsDqTnqoMkolvQ2PhyBew9fVY7ymIqZ/yHOh2VMeMrxzS8eGn//JxjfwPHih9Ihheo6iMqg+eYASrhWORCib9AEcZ6ZDzQz4f8bOGM4g/jEHMEZfmBMB3bN3QMmSTeBSFn+VfcPyTLuLsL/AlTCSodKWOlQCSsdKmGlQyWsdKiElQ6VsNKhElY6VMJKh0pY6VAJKx2zjDDwJ4qzIOVSW8PxAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Samanage"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.samanage.com/saml/<customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_samanage_customer_domain_samanage_com,
    citrixspa_routing_domain.rd_samanage_customer_fqdn,
  ]
}
