# Mango Apps — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_mango_apps_customer_domain_mangoapps_com" {
  fqdn         = "<customer-domain>.mangoapps.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mango Apps"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_mango_apps_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Mango Apps"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_mango_apps" {
  name         = "Mango Apps"
  type         = "saas"
  state        = "complete"
  description  = "Best Intranet and team collaboration software making work simple"
  url          = "https://<customer-domain>.mangoapps.com/ce/pulse/user/overview/index"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAIAAAC1nk4lAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAATESURBVGhD7ZTbbxRVHMf76D9ADA+oTyQ8mOCLkmgTYkRMIEEIiFYaosY3UpEHuSVqAgQiUWmqVoJIbIOUULClpKWktqUXd5e2e+tuuzN7mZ3Zy1x3Lju3nUt3/G272tJoI2hGTeaTzeR7zvnNOZ+ZPWcanP8hnrRbeNJu4Um7hSftFp60W3jSbuFJu4Un7RaetFt40m7hSbvFWtIEN84ZjmbZ9fZ/hj+VzmtOjBz0Yx2E6lBatd7rOKat2LZVbzhO1V5wqsuj7rBaWjIsVJRZw0EK/TLyXv/U9vvoOT8xhrMTaLEHKfRU9fjG/uQr97NL9Q2tgQlaWcqPxAX/azfiR+qNR2S1dFKsGWSpOzPTjdL8fivRNBR+Zyh5zJf4OIJdpKQ4jDb+jD03mITQiQlP9yEBTitVrCNhEpMNeOshTj0zx4wtPklc1Dsz/LuB/ASjRAX9RJRa2myGrd2cO9YW2AnZsiuUjE4S3+elGDRJOeHLdWaFGciUgg5nvhL0IuSVPCRNqoZoOGPIqWDo9dTsPiXxtoUekG69Gv9m/4wkzXB6Ua0t+fy99NEoRarmtvvZU3EmwuvrexHoX9+bAOmGG3HVtNb1JoKcumMMFyrWxoFkRja2jmCSYT/RPQeVw5lWTs36ch2p0oRqSmfHX7AWjPap3ZyKnx7bbNr65WBTVgh+O70HitsCO+C6kmXpim2zuhPB24UFBw025uNv6sgBLfoW07F1pr0lrDozjETIGlQ+O5AsqObxKLVtFDuf4OBN3yTELlxs6IqBdJMvDzV7fTk/rez9hZhg1S1DmQ9DZEquQH9LsBiXnVbfS3eQ09dmD/04e8iw9XupL2AoRg9MEp396BnIydLkdOFG1+wHUEOIIehZybJ0tVpNiTKtyDndYeeb6fgbRrJZDuyjzm8aQfGIoD2gxaU/d8Pt2ntt6IhcyvCfxJiONL9poLZb1vXMOwvVneM45Jdrm766oQ8ZIWVoXkqXPptnITx5O1VSiavR9yED58a3aJb89YNdkLtmW/JSvNW/HXJv4mSw+BPKjUA+PvRMrXQFD20PcMLLWlp2qEJ3ObbdQpuUvkbf3c+ny84UI87z9QN3NkbD9QLClk27v1CmNfN4hGr256/j4tUMfy0rwuh3aZ7VzV3j+OU0D5ukJy+diFAvDmOjjJETRik5sTiTg3CjCDfWFTt8a+7oVL4Leq6EDvbMn5wkrkDuQz69Hjsco+8u1i6z+iAClKqnFEcuXOTDe/z+tokSfFJMoWKGuTIcU2vhr37g4HQ+dQf5MsFuHkyxleWv5CrgnHXPfbSUdUv+IXxwKa/BH0j/jrH4+5uI1kI9/XOsJb026QxGUXQ0OisIQiaDFQrFfAEuRZ7nsSxeMSo0zRSLJEWS5XJZlpV8vpBKp2EI7vIHpkiKhh7IWSKHIElZlnEiFwpHeF6IxedIkkqm0jBbMBRBULS+5G88vrSiKKqqZXFC1TReEMCY4zhYGNYDV5phBEGEh2AYtlhzp7M4DjW1IRhjWLiCNExCEDl4bI4rMSybzmRM0xRFCcNwmJOiaeiHew3TrK+6yONL/4t40m7hSbuFJ+0WnrRbeNJu4Um7hSftFp60W3jSbuFJu4Un7RaetDs4zq+xaUT5yG3cnQAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Mango Apps"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<customer-domain>.mangoapps.com/saml/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_mango_apps_customer_domain_mangoapps_com,
    citrixspa_routing_domain.rd_mango_apps_customer_fqdn,
  ]
}
