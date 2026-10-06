# Microsoft Outlook — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_microsoft_outlook_outlook_com" {
  fqdn         = "outlook.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_microsoftonline_com" {
  fqdn         = "*.microsoftonline.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_office_com" {
  fqdn         = "*.office.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_office365_com" {
  fqdn         = "*.office365.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_sharepoint_com" {
  fqdn         = "*.sharepoint.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_protection_outlook_com" {
  fqdn         = "*.protection.outlook.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_msedge_net" {
  fqdn         = "*.msedge.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_akamaized_net" {
  fqdn         = "*.akamaized.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_trafficmanager_net" {
  fqdn         = "*.trafficmanager.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_microsoft_outlook_cloud_microsoft" {
  fqdn         = "*.cloud.microsoft"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Microsoft Outlook"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_microsoft_outlook" {
  name         = "Microsoft Outlook"
  type         = "saas"
  state        = "complete"
  description  = "Cloud-based subscription service by Microsoft."
  url          = "https://outlook.com/owa/<federated domain>"
  related_urls = ["<Customer FQDN>", "*.microsoftonline.com", "*.office.com", "*.office365.com", "*.sharepoint.com", "*.protection.outlook.com", "*.msedge.net", "*.akamaized.net", "*.trafficmanager.net", "*.cloud.microsoft"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAYAAAA6/NlyAAAMoklEQVRoge1ba4wdVR3/nTN3Zu5re/fV3T4pLaUPWh6NqASqAVNMrZKoBPmgGEQ+GEk0xkhAiSExKH4TjAl+kQ8SP6Cm2ATRojEBRIgogkjtAoVC3+3Kvu9jZs7fnMc879y7u3dXxMg/vTsz5zXnd/7P8z9TvEfv0Xv0P01suSZ/3R0POmOztSs4t2oAqX8hUU57Cv+2VVLbHQdjYV2qOVG6h7lhnBrUqj899uObZrKjLwvgj331vh2vlLb9/s26P9o2W8oipywu06S9fN6+XeqrNpu7Yqj12cfuufGR5Fx5ryBD2r17t1Xvv/CeCCyFP8qfECXmpZpQW3mqL80DltJjkxlquiXKr80439+7d6+1rICPHDliNwRb3RUoJWYSzo2oHQQ6LFJyQZJjUdwnGj7x/jkf/WNjY4XkfFMPvVAQBBxEeuF6Ft+MLi9SfCm3Xpc1Go0UU5cMOH5/coJL0VOkF2A+oJ3GMFzP0rIATopqHrdoMUAzk26rXwDQbrQ8HM7qKDJAcyadLl8OoMlxqK1ZSEsG7DgOa41s4K2am+9XO4D70Mby2Va5tJKyzajz2qxawadrGwb6OrXPvqrkNdjjB76Sms6SAUvuMkYCJDIV2fs0VzggIg1g3YGGNoKBidR65oBNaQdjbTxeHpEOAobAz58tcsAzBhEEILlICi3Ll+yoD4VeiGVf0U1Cuop07Y6XpIP+OIBNiwE749Wd0qxY7za9dnBZxMSUXnvMh/AdRnKRmGXCAdYRaFiWjUcyo7fV5U0lyeGfA/jUYsCqQe0SzswCxUX0KTIbs6dOFHnx/ARerjifdHHJ2CKEkAe0kx7nsVgBrt3x0vZewPZKUrOmW9VStVUHlzOwJaN5O9AEAnkRWe52MVhZp5ECDGA0p+4/SgExJubmgLIFblkAEyDO48ln4hjKcLwT0DwTkgf4HSdpsESrBTguUHDAJKsVh+MNXBsA6mKwcsrTJl3TkjcPvVKz6XHhNyENlwRPiRknDVQEhgBfBJwyIp8U32y5cQEp6onDFmfYscrFxkEHFZdjfDbAP041cGzCW/AYLd/n0jXxBLI2f5yyvISJybnKAIZTQJEn1jn3IS0KcJ/L8fVrVuLmDwygv2S11T9/rI7v/e4MDh5uSzS0kQwj6nMeLKsJRh5YqwDGA1Uezjw9eYJPgs3OeDGweXQ5CII2fAsGvGnIwf4vbsCGAadjm13rSnj45g340VPjuOvXp7rG8YEA3jw+BZQcwGWA0wS4FUthTrxcchl4ZSJV3Snykg2KRd7XE+DBsoUDt56Pdf12VPb3kw385tA0JhsBtqx08elLaqi62iTctnsIZ6Z93PfEuY5jimYrP7ORu4vS9/70FIDh7mCFMI8Mdr0+0RPgb107kgL77cdO44dPnkvN7buPn8Ejt56PbSOuer7z2hH84oVJHJ/spNeZGDACG4t0fDU6LqjdvyaBysAlMSaxdjs9r5WWuvq5ywei55/9ZQL3P3GuTVxPTfv4/ENvITDvKBaY0vWOlDWvyHI6a6pzNhhqEYQCS2ZBIHSZkL9AtL19XsAfubAKtxBb9/uf7CymY2eb+MOrs9HztVur8w1vuJpN4mXQZeqSQFWR0Ls2CgQEBeoq67xmyyFKu6Z5AV+yJo6ST055+OfpZtf2f3w9Brx9dJ4IOwso4miOY4UR04Tek8JFCjwFBqgv7wnkB5gen6z1X/qZ5z5xy117Fgx4tC9W87cW4GePJ9pIyag43V6RjS7QAajRUbMoSpdFyGXSgYsEKQIdyARyNxYo0CdOT2x/+sUTj+275e6tWIjR4otM1WfNhJ8X3yGbFjJXludQ4z2z0tUQqKwOhOG6UH5OyPECYfbZAPkyigvA3BIX23btBnB4XsBTjVjx16ywu7ZVbWpxm6lGgKbfxRmnRDmrw4ajCijpZwXYAAoodZXgJYf1M+m2XhNrr7p6tnLNdW6xcdTat28fmxewDBlDkq5p87CDV8+1Ora/amM5un/hRKNjOwUyBBLZlZDrLAE0iBdDhMAS4A1Q8mVZoC01CVhFB307dgqyd1Xq8lV+wCYmJuYX6ccPz6j3hqJ92+5hfO2RE7lt5WLs2RJb5kdfnu4COEBh/BgwNwvm9AEFVycBkgsC7WY0d0lVn3uxZUCF1jiIxTgIUCg6NHzJxVTo7+NC9hBaV4TQqzovYBk47H9xEtdfWlPPX/jgAA6faeKBp8dT7ST3H7rpPLWxkPSvuUD57E5kNWchmAWGAkilPQoA5wnOGjOsuM4VaOa1MNGq6LjUAFXGLPAVQ4a2rMPghmHGuEyayR4ywchUxoGMZV9QpCUjq6s3VzFU0RuGe69bhRsuq+Hg4WnMtQjbR1188uIVKCcs8jcOnFQ63JHkhIWZuPyxwBgtlnFPBhQFioPwfH0lDVTWrVi3Eis3r4Zlc70+kuOMqeGEvDIVlCyMwyGXr3/wKB6++TyMVHWX960vqV+WJHO++egp/PKFye6DRhM2QJhIhJUh0ARgpZ8+4Hu6vfBQGqhh5dZ1cMuO7qrMgQ4xdWhpFi+8Lma39LfjdVz5g9dUXH3jrlqKm+Ecnzgyi+/89jSee6s+/4DKHgWA7wOWzF76iY0EheGT4SaZhWkBnge7aGFoy4WoDJZlrlrrdCS6zKSj9bOsF0xEQy9qP3xu1lcG685HT+HSNUWltwXOcGbGx0snGzg74y9glBCTBONJ86m5JjN5lrHQZESYYtGXlphzwuCm1aitGtT2TQoIM3qqdswSrBRnrfM6C6rrxWI5nKSGJ/Ds0Tk8e7SX3pqKa9aifvKQCg5gyZBQbhctzRnjR0Ha9bDAR3XjRgxu2UywbabACQlGKDAqPRSCZhp0yGG9g1L2YOE6/J8gViph/Yffj5kJH+Njx7UlltxEwmgJAXflEAa274BdKWleCW2QotCMJa/GDDAtKCzUX8RBzX8NsATELY6BC9ahf+s2nHv5dUy++kYUhBRKZay46CKUhwb0qURmvxvhMHqqACaMldRjlS4yIp91S6ffabyMfBVTSG3jro1VV1yGgR3b8fbho2BWAaWREa17KvCI9VRhkMJsQDF5fMO07lOC21qfjUgzEZ1ZKFM7ee/OQwD2v1NgucXAzvxV5bDIhJHSqDh9ZYxefhFqF5ynuRZt6klt5tWOKCAt+WoLKLVAP1NYL0x8LcL2Qtk/WYaMSN+w4MM0IubOnVjFSv7t2XxL6tOG6DbWLxYEEOOHwKt9ALfVj+SXWNIIGU/n9BVhlx00JxvwZk3cnjROTMQ53Yz4Kv4bMQ4jLWncwnA9Ajx57065DgcWwqHNmzczVlyxsfHRL91uFiC+pL7OoRivnI/MSloFsLWbwCv9YG4FKNiK09qusrgrY3D6SyhI4BN1iJaf8q3EKBVgqFhSHciJ0GJpyQnb8wzgxRLnHIXaqtSJQQw8nYQig1gCVkcqtgvuVsCcCpjlag5LlyTivpFxsiwUh6rw6x68qbo6V9ZAjeUKw1HV2GxlmVloY9mUVediiYAtS1gDo5nvKRLHJZmvekKTweQGgdtgtgNmF0HyHly10LugVLdITbhrwR2uKhH3ZxoGsFlIFoaRwhgrpPxw8ouengFbzCJR6W/7HImQEfFo+kwbKKY5rXZIVkFHRcQ6AkV6eCXilluAN9NEUG9FekxGpJXmZvy0EunCEiItqPNrTpAiGTGFQq8Ri2NKvPWCE4uBU3jynwihYzXIgE3lmxkK1aJyZ/5UA8L3jV6nI6x4MSj2770C1ggKxi+aeRr0ic2JeW98ch+nrViclOsKlDLPcTnjDHZ/CUHT12IuKBVmKmlCYte0VMDhwVc2axWmopCea1RGLB9QyqMlLXxKPcIEYFzOHAv2YAXBnIdgrhknEKJoLB6hZ8Dj46f9KiGlw6nTCEqLMwhIG+88A5WoC8uSTiCRAW2TCtIRG7MtBLNNiKavjJkIuSuCRs+ApTM/e+wNr9L0BNkFvhhRbAObMkyU4WR806l/SjJISx0vFcGcQHFcbj2lm5849OwzpV6/ABgZGYHneb77ylMPkB9o4TObdqIwL0zRSQHl1CXLde5KJL6dTvQRIrd/PIaIxiCYsRW7OayqA152sGbqmZ8+/5O7j1er1SV9EV8BsLq2Yefa9Xu/fCV3y+UOOfdM4I8wcGBknllyFsxEXFFfmHAR0Rhxu7hTOG4Uj8hcYOA3z/7pV38+9uT+MXlSBKDeM2DGmEVEMpUp92/Ocv7/iWUkiV8ehr1dKBSmfN8PljRJzrklhHCNLXi3AvZt2256nqdSqMsyyT179rA5+c3Vu4ykzh48eLCDov0/EIB/A7AiiEufY2DrAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Microsoft Outlook"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.microsoftonline.com/login.srf"
    audience          = "urn:federation:MicrosoftOnline"
    sign_assertion    = "ASSERTION"
    name_id_source    = "guid_b64"
    name_id_format    = "persistent"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name        = "IDPEmail"
        value       = "ns_user_email"
        format      = "unspecified"
        prefix_expr = true
      },
      {
        name   = "http://schemas.microsoft.com/ws/2008/06/identity/claims/authenticationmethod"
        value  = "http://schemas.microsoft.com/claims/multipleauthn"
        format = "unspecified"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_microsoft_outlook_outlook_com,
    citrixspa_routing_domain.rd_microsoft_outlook_customer_fqdn,
    citrixspa_routing_domain.rd_microsoft_outlook_microsoftonline_com,
    citrixspa_routing_domain.rd_microsoft_outlook_office_com,
    citrixspa_routing_domain.rd_microsoft_outlook_office365_com,
    citrixspa_routing_domain.rd_microsoft_outlook_sharepoint_com,
    citrixspa_routing_domain.rd_microsoft_outlook_protection_outlook_com,
    citrixspa_routing_domain.rd_microsoft_outlook_msedge_net,
    citrixspa_routing_domain.rd_microsoft_outlook_akamaized_net,
    citrixspa_routing_domain.rd_microsoft_outlook_trafficmanager_net,
    citrixspa_routing_domain.rd_microsoft_outlook_cloud_microsoft,
  ]
}
