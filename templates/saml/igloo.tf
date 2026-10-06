# Igloo — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_igloo_customer_domain_igloocommunities_com" {
  fqdn         = "<customer-domain>.igloocommunities.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Igloo"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_igloo_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Igloo"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_igloo" {
  name         = "Igloo"
  type         = "saas"
  state        = "complete"
  description  = "Digital workplace and intranet solution provider to solve IT challenges across your organization."
  url          = "https://<customer-domain>.igloocommunities.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAVkSURBVGhD7ZfZT1tHFMZHqP9G1YdWbdO0UaW0fWlfqr6QRQUEARtIsEMLL31pH1pVTSCIpqoUpSlLFHZv2CwKSgphT8pibCdASSAQwIDNTjBLE8xqDO43984lNjbO46jS/Wl0defM8Zn5ZjlzTXz/c2QBvJEF8EYWwBtZAG9kAbyRBfBGFsAbWQBvZAG8kQXwRhbAG7K/v7+89XJle83t2WK2QLx73paFodtTPa0Lz3a9u8wayLpnCxEQx7u3x0xHgGitC0PVU92I6fF6mDUsj5YdNdM9f832Ta4vMZMfpGfFQQpPkdKoMx05zCbhdLverf+FlMUSfSIxJNGnJvatup9GXy4wD4mYrpukJApxHiwMMVMQiPYOomn8opXFvln34/CLOeYRRKK1iGjPEb2C+qPg3aTOtz9gzQJk4N9polcSU0qSrZjZBG7aH5DSaFJxkRban5I+xWppdO7ofeYnoHqkIcYU+HS5RpkpkFtjbaGiqcVo14ebmJ/E0rabOpSfJ5Xf0Cf8IRhdoKpLONZwmfkdJaB+7gkpjSFVaUSn+KAxw+C09ixP6J22E01Z+D2p+pYURhaPdzDv1wlonB+go6fREj5uzqqa6n68Mnl7uvdkSzbRxVN7acydmX+YtwAdbmUqYkZUpd0YaeledrQ/H07v0RFtPNVgSP7i/u/MM6QAasEotfE5o63MJJGHlSk++31f5Y7feQgvgI5GiFYw1sZMEjpHF90YaNUrmcnnO9OZS4wXMKRPWrKZSWLDs0Wb4K+J7VgcgSWEAI3DQrs0qc525omW1xJGQLnTJsRX4ZwwUyAKayF6R486h0W00HOCaTapxeohupcn6C6oUH/W+huqIQQkWAqE0Sjsa68Oa5fLfqm/5tfBWpSrg3V4Zg7cqZx8KLaGEaDEQUSTTjGx9pyZAnG6F9GKAcRZClDFVqHjMV744XGV6BBMRHUa3WDlyXgPIeCrtutQjxX37e+LFpD1tJZODM7TQdEpPmrOElvDCDjVfgPTj33i2z8ywxJtHHr8XNjWlVPdxJCIweWNtYutwZxszhZOfwzeQwg4bysRR/PML8FdHrhLEyU8USpSSXU6fOKFOQNhBFyUmo5Kl5Nul7ACqmgz3WPIFrRqTPmu1yg6BBOB3ulRTsR7CAHIEoIFZyBXtID5zVWry479t7j5AhmAatArjU6b2BpGQM1ML03kJlVkx5/MFMg57FgskT5R6+gSLfRYY3zGFLF6CNvSGFVYof605SqqIQQArCCpDJ2FSiY6qT/ygCGJmfwEDKxOMZMfdL8hmi7h2nAjM0kgL0lZiE6nCD2E+IlJ9X5DBjNJrOB+wNiELGQWJiu0gDacJJa5Fbg1dE6LxWU3OG3vNVyi6mnmjm6af8q8DwRUqE935KR2a1UPyxTWolvSlWl22Vk0vfLtez8XjXd2Lo6UOczHGjKEW4VGq5/rF51FqAAsMlbGpLoyWIsL/u5sH0YoLA6duy//vsY8QwoAJRNm2qt4WRqSqQ+7idWwl/rdYoAJwNZEK15QdIoYcz5r9vkMyPcH0TCFSNOIKUXDPc38JNyeTZrv6dKJN7Hw9QE9qGrjcQMyPwig30IFkaTk69NBe3R2Y/V4U6bw9SKMHk9N7PHGzJmNFeYhEdWVjwgI/aqUREW2/8GaBeY3Vj9suhIQrYxGm15fZh5B0HnBEqEgL0GDJu6N6vSD60KEfk5v7m5v7W57jvjSBNal8brZJ3iyehD47dbuzo7Xc1CEauiA1qWx8NEOgWR4b64f37Cr225m8kP+Q8MbWQBvZAG8kQXwRhbAG1kAb2QBvJEF8EYWwBtZAG9kAbyRBfBGFsAXn+8/HVTd0Wpcr5YAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Igloo"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.igloocommunities.com/saml.digest"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_igloo_customer_domain_igloocommunities_com,
    citrixspa_routing_domain.rd_igloo_customer_fqdn,
  ]
}
