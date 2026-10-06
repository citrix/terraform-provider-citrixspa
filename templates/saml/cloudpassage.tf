# CloudPassage — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cloudpassage_portal_cloudpassage_com" {
  fqdn         = "portal.cloudpassage.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CloudPassage"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cloudpassage_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CloudPassage"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cloudpassage" {
  name         = "CloudPassage"
  type         = "saas"
  state        = "complete"
  description  = "Visibility and continuous monitoring tool to reduce cyber risk and maintain compliance."
  url          = "https://portal.cloudpassage.com/halo/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAARHSURBVGhD7ZVtTFNXGIBPXOKWzC0Lc2TLzKI/tmUhDkckqBgRx4BtqFSdTASHFZwJbmOoWzaHg03NEj/Gl1hxbLNWhKIFGawUKJTSln7XWWgpLbXFgnwI5aOlBUq7t9xuWbYf8485bDlPmpv3Pee96Xnec+69yPcfhwjghgjghgjghgjghgjghgjghgjghgjghgjghgjghgjghgjghgjg5v8iUN8izP72zDcFDIttgFVdx2sT55exqKmHpLT8xtWqmpF5X5aq98Qd6xG1udI6Eph7lPgFkjKPhcbvKqvknCwsfSF8y5bkdInq9jOrI6mKh4TBYq+J3QFBcI182Y2Oo2oz+qnlA2kPNfvoQFLNnVUb3w5kPl+P2cqu50GwfM0muHYaTCsj44NCN25OokP6dMiGA5/lQhD2zu7wrXsgECnUcPvza6Pz8hnrE1NgJE3aEy/ohAB2YEmFSD/hRKVNqOjXyKbf0A/NBxUmKEBX2xBTcEBm5A/a0Y8tT1RJIIVbUjoMUIkuNa5rvI1KGgYc7iMaM7rARZebovladKFh2DX7saoXXeTBb5dID7cgxrWqg1/kQfQ3nnsjCq6wIZouf11eAYN+7IRQqkr99Dikd+/ZwAECcNObzBB8fb4kYlsyBND15dWyU7p7iCWkSw2weuHwOIxnyI2wVghQYX2tbdQ955me9ayqU8Yu2OrsjuuWYXSlFWIAlMCzdUGPGnmWI8tQmtSjU0vZEmokuEZ2xTyEGtslITE0aghwOKdbJDIIoKkut/vV6G1/jDtXx+0UKTT7j+ZAOjA4BMs1Wvoitu+lCsbGJ0LjdkIAfX3qZkeypDu/u984Of3nCuY88yADQfOg/bFKMbQZ/h7S17kaaPYeSXe63HhIYVqo9eV1WqERWWpzcoeBGik29KcpjGf0NnRd9GSVZClb/OIt+XlIYW5d4t641ENtUmV1A/+lDbG5+RftE5PBYZthKuQtGvNm7czMbGrWl1+dLVJpdZRtCbOC2oEVETFw5KZdrn3Zx8MT3vd5vTECXRhPA1MUiNF4uqvP5ZlfC6diQYbWroPruW4bqhBdMt2XP5j0lxXWc/pG4Kho7Q6/NlMQ29qpGZ1CDB4UWB1udE2YrjTZnG70c6AjO0R655wn8BbKOVcc9m7SpvfSqAcgM+dUYsYnp4sve71eCF57c/vn331PVcLL6uWohJNFpQn7MytqudD4rfSP1tNSwJNGP/xhCTOoTg0PMafvAVVvcbgghaVnq3qDONJcrfWVOuWKW3JULlSMTsEiQrlqOGxU7wsM/bB0EDusNC2pFI+5ZgVD44+zxSt/Ue4W6ZMk/t1g3YWTJgCNs3obpIv6OzADp668nTswZplyoTJ+8317YOIvLGoBoGvcGS/oiuJrqTfBP1nsAv8KEcANEcANEcANEcANEcANEcANEcANEcANEcANEcANEcANEcANEcCLz/c7soN51cDk824AAAAASUVORK5CYII="

  using_template   = true
  template_name    = "CloudPassage"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://portal.cloudpassage.com/saml/consume/<Customer_ID>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "firstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
      {
        name  = "lastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_cloudpassage_portal_cloudpassage_com,
    citrixspa_routing_domain.rd_cloudpassage_customer_fqdn,
  ]
}
