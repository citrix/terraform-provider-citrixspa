# Dell Boomi — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_dell_boomi_platform_boomi_com" {
  fqdn         = "platform.boomi.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Dell Boomi"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_dell_boomi_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Dell Boomi"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_dell_boomi" {
  name         = "Dell Boomi"
  type         = "saas"
  state        = "complete"
  description  = "Integration tool to connect cloud and on-premises applications and data."
  url          = "https://platform.boomi.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAUvSURBVGhD7ZhrTJtVHMbPB6PT7LOaLDEmxJh4iQpxflA/avgykYVFVAi3AIrjJpfBhsMNmWXBjDEuCjqz4UjGbVuAUqCUgnQBLOPSAd06VmB0WSnlUjpuBepzdg5QkOEYlzdN3l9KOc+/57TnOff3ELuTIxoQGtGA0IgGhEY0IDSiAaERDQiNaEBoRANC8z8GmocscXV39//Z/kKaiiQryfH6Z079/VZe2zdSXaXOzDMJymMNHFP2k6N1JEpGImUkpXH/+faDJT2HSns/vthJzSAeXU1iar6+quUFBGIdA1V9o7TqkbI9aaps9f3FxUX+wWqKe02v5aqpvZian1WDPLrrrDXgc01L6/RDffntJxoht8xTe09fR5HXc9U8tLusMvDhhU5U5Z38G1w/McGVt1Hw+TQV17vIioEvr2gxsg8UdXO9Sc62DsHDvrMtXO8W3AAGDH7+vd833faOxMrv4ku+k93helfgBkhiHZZIlt4Kr2a1woPROsv1zkMNYKXH4JHe2YZ1fdq2gOX13a315KagBtD8e09v2/zzKO5GJ4zP2LjeYUjj4DiJqv617T4PbBksrOjPBIWe6x2GRNf0ocGY+EtjLNWaintMTYMTLMI4pRokPzVaZueZLNAYL900Lkvkv3RzeMgywyTAdNqXuf5yNDAwkJ2dnZWVVVFRwUNboLy8nLj9cQOVgxibtpGIKnJMQWJrSGwtC4KPLnaS8Ko389r049ONA+OIkMNV8NzxYJJloAtAlOxCl5FJ8EmhBtszF6spKirCr05OThYUFMAGjz4tHR0d5DmJCmc1iInZeXJEjsZLVQ2+nNGMzfiN39QvnmlGVdKu39OPTZO4WnJYipwkQY5Kd5sePvoSOznZgKPH5R4Tk+BInR6WJpe6yJGSkpLubr7VREREsERSUlJ4eHhOTg6TFoslMjIyLCxMLpdDmkympqam1NTU4OBgyMzMzKCgoPZ2WueysjKCM+bnxT0Q3EA8LQNoGjU7Xu97TXvuHwM6gU30t/PayImGjQ2caRmCAXjm2gGpVBoYGJiSkuLm5ma1WhHx8vIaGxtDQqFQoJZIuLu706x2e3JyMqo+MjLi6urKIi4uLkYj7WoPDw+8w+f6BkanbHRg4PyM149KpLFPI07T8XR0IfJgabFnBqr7RpkE3MD4OgbQAxqNhqV9fHzwHh8fzyQICAhQq9WO0yM0NNRsNqPVmYR5lggJCcF7XFzcf4bQiYb3z7dj/KCKX13VHirrpcGUxlfOtdLEUYXBMkvPqicbEHkpo/nKLdOzkiaUQhoT9/vaPnzVxkOoq6uLpVkrenp6Mgl8fX0nJiYwopjs7++XSCTDw8O5ubksAj8ssWJgeRKbMYnDKmktY2rw+mzpUETHD6Z1Qh2qyCLk20fZ6JSo+qXlHnWLboGMqPq0kLbuBpO4tLQUQzk9Pd3f35+1tFKp9PPzKywshB80NiL4NDExMT8/H6MLUqfTIUIL2+3e3t6OCfgh0bUryyjabMq2YJ1b23JdRusXZb1cLGVjOfG0MDXH5cO5+dn5BWTYYBkFKGKzrd3m9PpV+8bMzIzBYOBiQ5x/I6N/Tn2UwN92HubmhDjM0X/bfJye43rn4QYqdNv2QHNYkAcagFUfA+nA5ad8pMxgj5SPX3x2iBUDwLkf6hk4+aAqznqtwpCxi60o2R6JKkttWHCui61lklZfLeKAdLCkx8sprhYdYZe7Hzjj5a6zIBoQGtGA0IgGhEY0IDSiAaERDQiNaEBoRAPCYrf/C2WWrJbytOiUAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Dell Boomi"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://platform.boomi.com/sso/<Customer_ID>/saml"
    audience          = "https://platform.boomi.com/sso/<Customer_ID>/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_dell_boomi_platform_boomi_com,
    citrixspa_routing_domain.rd_dell_boomi_customer_fqdn,
  ]
}
