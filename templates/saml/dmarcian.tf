# dmarcian — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_dmarcian_dmarcian_eu_com" {
  fqdn         = "dmarcian-eu.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "dmarcian"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_dmarcian_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "dmarcian"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_dmarcian" {
  name         = "dmarcian"
  type         = "saas"
  state        = "complete"
  description  = "Email monitoring tool to filter spam, malware, and phishing."
  url          = "https://dmarcian-eu.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAANoSURBVHhe7ZpdaJtVGMd/SZq8iSsK6ZeFoiEyQtwE21qXiTdVrIIiClPwg64wJ4PuYqLuRkRhsgthXqi9EHvh6heyKWNUFAU/1g2zxHXD6qLQOLroaJo0Sbd8NFmWmBwe0btmul2Mc36Q/J9z8oT35Pee973Ja6s3QGPsktpiBEhqixEgqS1GgKS2GAGS2mIESGqLESCpLUaApLYYAZLaYgRIaosRIKktRoCkthgBktpiBEhqy/8W8EH8oFTXJv9JQKFQIJPJUK/U2R1+VWavHuFwmFqtJqMry2UJOP3LaT57/yAnj87y64kY0a+Pc1NHH9nEsnRcHTwej1RXnpb/Ho/FYiydTfKe9xCF1RJOu4NLtTpn8wvsLb/AK+436XL1kL94ni7LS7qywjpnO39eWOCHh79gz8l9zJ9fwOVwsVRcYqM3wK4NO7jvy8dZf4OfYrXI07ds4WgyzJmVBG6XuzFX4sXBHSQOnWHsmTGmp6epVCrY7Xay2Sz9/f34/X6mpqbo7OxU62zuzNHRUdrb29V4LVoWMPnuJNu2b8N34E4WHovKLGw4cDfvDL/BJ/OHeSv0mpq78cPbWHxqTtVbv9/J7o07iecTfHXuW5wOJ267xTd/HGE0+CTrcDIWeEL15isFNn/+IHOPfqfGfzMxMcH4+DiRSIRcLofNZlOvZDKpfnggEMDn86nemZkZ3G43Q0NDarwWLV8CHR1e8rk8pXqxsdALLK9m+Tkbw2V5yKTStDkc0gkXa1WpmnWN1GqKZ4/t4u3Ne3l98CWeu3U7K5Uc1G2UqmXpBMthkV7NsFzKkCymOFdYRJ2dxlsikSAajTIyMsLw8DDBYJByuazuDf++P1Sr/xy7FVreAZXGwfZP7uf6u7qYc8xTb2z/UrFEaCmI1ePB8l3HA733qt49p/bx8u3Pq/qj+Kfc33sPR1IRfkyfwm5ro1KvsNX/CLmGSIe9jU1dA6q3yYn0T3z8++HGfENo/RJb1j+E/bcqg5vuYHZ2Vm395tlvLjsUCikx3d3deL1e9f14PI5lWfT19anxWlz2IzKRY8dJL6bUIgYGB+i5uVc+uTYxzwhJaosRIKktRoCkthgBktpiBEhqixEgqS1GgKS2GAGS2mIESGqLESCpLUaApLYYAZLaYgRIaosRIKkp8BcWgTYl65BukgAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "dmarcian"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://dmarcian-eu.com/login/<customer_id>/handle/"
    audience          = "https://dmarcian-eu.com/sso/saml/<customer_id>/sp.xml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "Email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_dmarcian_dmarcian_eu_com,
    citrixspa_routing_domain.rd_dmarcian_customer_fqdn,
  ]
}
