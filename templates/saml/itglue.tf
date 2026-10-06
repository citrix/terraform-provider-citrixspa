# ITGlue — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_itglue_customer_domain_itglue_com" {
  fqdn         = "<customer-domain>.itglue.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ITGlue"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_itglue_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ITGlue"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_itglue" {
  name         = "ITGlue"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based IT documentation platform to help MSPs standardize documentation, create knowledge bases, manage passwords. and track devices."
  url          = "https://<customer-domain>.itglue.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAMcSURBVGhD7ZVPKGxxFMctKOPPbCVCsrCYyRhWs5idLTKhiYWFlBRNyXYaNYwSqVEUSlM2CBO9LCYbpRhF+fOU8mdBmQWzmEkxzfu655pG3Hd5Xp3e63wW0/ece+5v7qd77+9mJP9xRIAbEeBGBLgRAW5EgBsR4EYEuBEBbkSAGxHgRgS4EQFuRICb/11ge3t7Y2PjxztWV1cfHh4wEI/HDw8Pz8/Pkc/Ozk5PT3++gnx8fKws88Ll5WUwGNzc3Dw5OUG5tbV1c3NDh76DjkCGNm1tbRiYnZ2lUmtYWSZZWlqKXFZWVlhYiBCLxXJzcycnJ+nod9ARoOtbXFw8Ojrq7e1NJBLUt9ls7e3tCAsLCxjIyclB7unp6e/vN5vN6JSXlw8MDHR0dKDf1dWVlZX1cloaBQUFc3NzavENdAQaGhrq6upwQcj4zc/Pp35NTc17AcLlcqHT0tKi1sqJS0tLavEKBPAcNjc3pzQCgQD+C2Fvb89gMOCsoqKiu7s7OqqFjgBWub6+xm9lZeXFxQWC2+1G/zcC3d3d6DQ2Nqq1skg4HFaLVyCAt6uqqmpkZIQ6Y2NjuG8ImF9fX0cYHh5GVg5qoi+AF3R5eZkWdTgctGJtbe2XBHZ3d9UimczLy8OjWFxcDAGsMzo6Sv2JiQmLxYIFKyoqqANw7sHBgVp8hL7A/v4+gslkQqYO/tvr9TY1NaH8gzuA8v7+vqSk5L2A1WrFE4WBdHZ2dmjgQ/QFaNejjFuMEmF8fNzn86H5SYH5+Xm1UMpoNAoB3FJcMZaivt/vr66uxraB+0Cdz6AvMDQ0hFdwZWVlamoKG0soFKKtMBKJYOAzAh6PBx21SBNYW1urr69vbW2l/uDgIHaw29vb9GFd9AW06Ovrw8DMzAyVNA+cTidKu92u1gq4UBojIJCdnT09PY1nCWVmZib1YYXhzs5OZKPRSE1aQYsvuGrx+PiY+j4Q6KjpLfg8X11dUX56eqIAUh/s5+dnCgDPKjZAtdDmLwjwIgLciAA3IsCNCHAjAtyIADciwI0IcCMC3IgANyLAjQjwkkz+AvnCpGXSfBd1AAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "ITGlue"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.itglue.com/saml/consume"
    audience          = "https://<customer-domain>.itglue.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_itglue_customer_domain_itglue_com,
    citrixspa_routing_domain.rd_itglue_customer_fqdn,
  ]
}
