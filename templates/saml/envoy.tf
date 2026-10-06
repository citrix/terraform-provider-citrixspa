# Envoy — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_envoy_dashboard_envoy_com" {
  fqdn         = "dashboard.envoy.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Envoy"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_envoy_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Envoy"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_envoy" {
  name         = "Envoy"
  type         = "saas"
  state        = "complete"
  description  = "Visitor management tool to manage people and packages."
  url          = "https://dashboard.envoy.com/entries"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAATXSURBVHhe7ZtdbBRVGIbfM7s7u9t2u22ptnJVKVghChcKKmqRajQqQkSjjTFe+XNroolXJsSQoKWpXhk1xjtj0FgDpLTWqDUkGDAqNcqV7YJIQ+iPXXR3uz8zx+/sfjVBk52ZwswOzj7Jydn9ZjJ75p3z8845Z4UkEGA0zgPLimpA8dtjyA3th5y7QBIKjtYIVXw9Bv3xJxF/9nkO2sexALn33kVu32sQbauAUIijNYZuQS4uIrTlNiQ/PMBBezgSwDhzBuneOyBWrwayWaBUoivUuAYsk0hAzl5Aw569iD39DAetcSRAZs+rKAx/Qp8E9B2PULXrp/ZQrBysFZEwSr/8jKXB16kpRKG1tCD5xQQftMaRAOndO2GmpiGam9HyzTGO+oO/XnkZxfExyFwGraemIDR7/buzUUDVdtJLXNNe+e4jNNUsVZOUVMhCnqPWOBwGub2bjvpNbzDNSq6K6KBfCrwPqAvAeWCpC8C565iZDA1ROXuJTJZUPboHOPMBj+2COfUrtDVrkBw+zFFr0rsegjk9RaYlwhELaJSR0kTL+AS0jg4OVif75iDyH7wPWSygdfIURDTGR6rjeg0wSLDSD98DDY0kgG4vRaNkO7PIH/qMr+Iergugda9F+JbN9IgywNKSvUTnirZW6A8+zFdxD0+agPoBmU6XXaQtlNskTy8cGBrfNgGFug0tmSy/qNhKra2Obv5yqA+DnAeWugCcBxZPRoHS5EmUfpoEYnH6ZuPnDBNaZyf0e7ZzwJqVjgKuC2D+eRF/3HwjRIwK5KBnl/NzSBwYhr79Xo5Ux7fDoDk7C0HuTmtbBZFIQMQbbKZGmOfO8VXcw3UBynNz9OTVkwl1dSPcuw2hrXdVT7dvhd7/FKIOZndXjGoCdlncvVMubNogFx/dwRFrSqlpOX9Dl5xf3y3z42McvfJkhvbLhZt65HzP9dJcynHUGg9HAaoFai3BZwR+GKwLwHlg8VAAshu6zp/9gzcC0L0LPYrSieMoTHyN/NjopenICPJffckne4vrTtA4nUL6/j5o7e2VyU414/NvyCDKhXk07H0D8ede4KAzfD0h8g/UBERT039TI6VkC4kwxyd6h2cCqGlu7doOhNZvQKin59K0bh0id/Yi/uJLfLaHlO2QTS7PCa6VS6MjHL3yXAVOkMjbX7b2Cm8F8CF1ATgPLB4KQHbD7tqgh3gjADtB4+SPZSdY+HysasqPjqCo5hA9wHUnWEqlcPGBPghyglBL34UCH7EgnUbTRx9Dv3sbB6rjWycowiHIXLYyBApBBYtap1iUSqbB/O0sX8U9PJkWzw4NoHj0aHliFJJ3c1XDMCA6r0PTvoGyEHbw7bS4V1wdL0M+pC4A54GlLgDngcWZAMtDmDe7V5yx/O8VtZHbwWZuRwKoRcuyQUmdhuGzHeOl704AeoQsN6W4Woa3hyMfkHvnbeTeGoRoSpQ3MIY3b6Fxt8b/GKEnL2dmSIDjQDgMrYs8yqcH+aA1jgSQJQOLt26svNWptX67vt5tVPWnMpnnz6P50BFENm7iA9Y4EkBh/H62PM1dtqs+eb2VpkFWMIuGgSHEnujnqD0cC6BQc/v5wwdJ8Rn6UuMekTpmLdGMSN99CHV1cdA+KxLg/0TdB3AeWOoCcB5QgL8BXhYmNfR3+RAAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Envoy"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.envoy.com/a/saml/consume"
    audience          = "https://app.envoy.com/a/saml/metadata"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_envoy_dashboard_envoy_com,
    citrixspa_routing_domain.rd_envoy_customer_fqdn,
  ]
}
