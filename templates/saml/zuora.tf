# Zuora — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_zuora_apisandbox_zuora_com" {
  fqdn         = "apisandbox.zuora.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zuora"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_zuora_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Zuora"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_zuora" {
  name         = "Zuora"
  type         = "saas"
  state        = "complete"
  description  = "Zuora is unifying order-to-revenue for a dynamic subscription world"
  url          = "https://apisandbox.zuora.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAARQSURBVGhD7ZV7TFNXHMf3//7cf4tvAUXdhkChlIcC0pY+aAWkgDB0hHVq1T1UJkw3DEWdioAKIqAIQ0QCzKEQZbKw6FBTcZVHoQzmFkdkU3noVMro9dueO6MmbU3842hyPrlpzu93f/ec+zmP27e4NxwmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBsmQBtnAkGqlAiNVhz/3LUs7mNFsm5kbNxXmqBe/Slfagf1Amniw0ePSWgyD+oydqIHlBWUnSBJcPDYSYE0obbx/KkfzkWnfh6ZuJbka06fi/8kPXxFWpIu40xLG0m6xJnA23OFc4VycrkFKHwk8Rj4vdAYDDl0e3iWIBIhX2oHmZkC6cjYfbS/PXRsmneER2CU/Vn5vMCo+cGqsfsPcCsr9/CCELW/bKVnsHq2X6RYo0UyLDbNXaTwCFSiGL94Sr1qo61TVzgTuNFj7ug0Xe/qNXb33RoaxpC+EpsAbpnMA1AKUCSRSoJbgG1gi2Wyrd0ww0eCAn1eiXng5sUrHaExqXhwcUQcyvT5pR+ExfpFJi5aGl1V33Tn3qh2y46FS6I9g1UVtY2dpv6Sqjp3kRJhevZ+0rMTXvYMrP1S/35YLCZs3D6LXb39jgTQkCSswXwXHT9F8gQsIGa33WDcV1xh68pfNvzPXeQf/Ptwurd4jlA2dPtvUgkMxh4PkRJ5PnbMSwnkFJRiWaf7iLt6fyMZBwIKXGjAk5g8y/7iSggcPFqN84DZxUYn+cbzbZhvVcoGEj7FT7bSPUBhMHbzsQNcC9Q3XZglkOL6vrmVT3Fcp8mMdxUpk/nYDmaRCMzxR0M+MWEheYI+7wjuHq6ozS87gcaewnKSb269BAF8Hkj4FG+xBrN2+doNPnaACwFjjxnruCg05tnPCOjuG8Co84Ki+Jjjyk+exp7Ge6Ot0W7GMd2UtY/cAlbOSo5sb//gnqLjbkJ57pFKcmvCYsFxxwn+xfAryYC6sz96hqjnB6mudLyCwL3RMUw8TptI+WHrxavVDc04c1X1Z8uqG6xW6wxfCQ7lR5993dB0YWtOPvaMUJ4EgUePJwb/uPWuVzjeQKPdglcpLK9ZuGQ59HA20G12folNoLiCjAJ0GTkoQA9wxhDrtuagjW8GBC5fM/JFDnAmgF7Ip8MrfAXeDCG53lkQ0jdwM1mXgUXATGPnYGpDlq/2WhY3bXHEyOg4nm1pa8fSoQD12DDYVPgDId1u31uIst2HjpKQIEtah0osKfY96n3EmqUxqZgjfMH4Cgc4E0jZ8BX+idZnvnjhi/TFN3tRUPJd3cZtuzfvyC0qr0G4Jl2PRyYn/7M/zU1NTRVX1q7P3JW56wA+rCQJsCaQP9PyMx//z0+XrqZn56VtyiootW1X/JOg7Pc//yJ3HeH6EL/mMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAHaMAG6cNwTZa/CQctoNPoAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Zuora"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://apisandbox.zuora.com/apps/saml/SSO/alias/defaultAlias"
    audience          = "apisandbox.zuora.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_zuora_apisandbox_zuora_com,
    citrixspa_routing_domain.rd_zuora_customer_fqdn,
  ]
}
