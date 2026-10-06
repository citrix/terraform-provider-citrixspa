# WebEx — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_webex_your_organization_my_webex_com" {
  fqdn         = "<your-organization>.my.webex.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WebEx"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_webex_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "WebEx"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_webex" {
  name         = "WebEx"
  type         = "saas"
  state        = "complete"
  description  = "Voice and Video Meetings With Screen Sharings"
  url          = "https://<your-organization>.my.webex.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAABkwAAAZMBjE7KEwAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAA3zSURBVHic7Vp7lFXVef99397n3DsvYBie8hAfVQhUgwo4ohUIUbO0tlmJbVPbxhSXWV3tsmm6ktbENGiWWVETg0YjuFIFsSUgtj6SCAzBR6wSA+E1goOCPIYZBuY9c+feuefsvfvHOfvcOzN3Hne4A3+Eb62zzr3n3vWd/f2+97c3cJ7O03n6QyY6629c/u5YqNQ0eOnx8H3A1w6U7kQ63QT26vH0nS1nczkjC8CqHQ66+Eb4ahGUfx087yr4/mj4PuApBHcf0ffgcy08vxpKvTM6pTa1zTu1C8uX65Fa4sgAsPJAJXz1VSj/dvh+eaDpLGH9LKHt816AxD0fpWkfMaXrDbCGDT1bW/XPHxV6qYUF4Gcf/zm0+hZ8fx6U6itwD+H7sQLfhwiFL/X8bO4GhNcA/o8Tm+/dU6glFwaAF47NQ7f/Y2i1EKqvsJeNcnDj1DJcO6UUs8eXYMaYOMYWO3AEAwBSnsKp9hQ+PtmO6qNN2PlBPd7fcRQdrYlcbzOtvO3HCbfmR3i1pu5Ml35mAKwyDoprvw6lvgelHPg+LACTYoRls8rxpVljMXtccd6stTH4zc5jWP3aXlRtPxw996n1+OnYugksOSkdfCO1/sDPzkSE4QOw7sQ0aHoZvn8VVEbwC4sZ371mPO68vByuKIyBvbPrOO5/6g18fLwFDe7a942TmC8kgSVDSLG+FFjWsHZvTnMZjIa3wo2NC5D2XoZSk6zW41D49pXl+MbccYgVSPBsSnsKy37wn4f/e+djFwvJYEkI7gxH0m52cFvTyj0n8uWb/0pfbLoOyt8M3y8NtK7wx6MF1i+ZiFnlsbzZ5UPaaH3ZP9z02+PN9ZUZCyA4khGTdCju8pLDj+88lg/P/AB4qela+GoLlCqzmr9jWhzP3TgBJZLzYjVcau5sa5lyz0LFTOOEZEhJcMLLlXzEjemFNY/sGnJwHPqqNzRPhzavQOsyaA1ojX+8vBQ/XzLprAkPAGNLR5f/2bzP7GNBYEEQ9mICM2aQkS9X/ktl0VD5DW3lz30SB5uN0HpCILzC3ZeW4CeV48Bnv5jG3Uv/ooKYIBiQDEjOAkJgXqrEDDkzDA2AsjHfh9bzrOZvmRzDyusqzkEjEdCnplw6UXCg9VDz6Pnd/HXlA5VfGgqvwQF4qW0ejLrXCj+9iLD2+gqMQKAfMsXdmCsEQmEJggiCENwtCAJP3vDQDZMH4zUwAMsNA2oVtBbQGqQVVi8ci3FxUTBhhkOChRBMkAKB2Ut7AUyRNYx1Hf3gYLwGBuDKti9C67lW+3ddXITFk+IFE2S41NrV3mk1LUPzZwouIUKXCGLCV259fNGcgXj1D8AGI6DUcmgNGIMSNnjo02MKLsxwqLGjudOaf+ACgIjiALLiAgSzf99AvPoHwGm5CVrPslH/nkvimFx8bk3f0v4TH7UGMSAU1qZBQgaQ8DMz3XHH00um9MerfwCUWWZN34HB0onn3vQt7T6yL5mJ+pkMIEUQEJkJREGVx0wOpH9Xf7xyA/CL1nJo/acWgFsnScAYpLUZIZHyoz3H9jnM1t8JkjJxIOMCBJYGzAAz3dEfr9wAJNVSaO1aAG4YH4cB0JDwc/79bNPx5toLBQNEBEk2FhgIYULtB3eAQAywwJVf/vmSS3Lxyg2AUjdb4eOkcHFJ4PtH2j2caxuoaz7ZYKCmCCYIAZAksAAEMZg4CowcBkamICCSUUty8csNgNaVQfTXuLrcgRNWPe2eRm2HN3LSDYGef3fDQcGZ+l9QKGTYG5C9QguhEAQQLcjFry8Av/ooBqP/yFrAnNFOj59rWtLoVufODn75+00lghFVfswIswAgyEBS6PcizAIwNiVek4tfXwCayy6D1g60ArTG5HjPv3Rrg12nU+fEFZo6mluaEo1zMqkPEMxhHACIQ42H2mc2IEFgYSAkXZSLZw4X0JOt9qFVzrK3KaVQ05IuuICD0ZNbn6kWDNfW/xYEFughuC2MyF7EYMKor/3voj6VXF8AjKnIAKAR66ffPdSWxiftZzcevLrzlxf0rPZCFyCCIAaF320MEGRLZANigmfEhN48+wKgvRKYDADuAMXf/qYUjp2loLhpT9UeT6cvibTPBpIyFSCxCayBg4aIRXDZ3oAZgNROb76yz5u08a3/Q2sMWPsQobqpGwAwvawP74LSD157TEdVn7BR34CYQSI0/6DoibQfZAAbBwBAuL359rUApdPW/6E10v7A23IGwL6mbnzQ1A0zQpFxxye7a5q7mudyVtoL7hxpnaJeIOP/zEFKDKtBSPY7evPuCwDpxigGKIW27qFVf0c6POxqTEGNAAr/uu6+RNTyclj4MIIoHzU9VuAADEFZQTEEyRXydG/efQHw9bHsIHi6a+jRvj7h4//qkmhPF24zt6r6jd2tiearMkIHAhFxptDhLBew6U+GLbE0IMEggeTDS7e29+bfF4BE8VForQIL0DjUlMprwR2exrt1XQXJENpo/Z2ND8YFEwQy2g8aIRNaQlYbHIGReZZpimg/qG/50heAr09LQqlqKAVohT0n+4A2KCkA+5u7set06oyqxhWbfvpeyuuaGY25KBMDiDLRPRJe9AJB2JjAAGFvrnf01wtsty5wqDGBpkT3sASoS/h460RgDfmGhsaOpuY176ydnWl7AwGDhicULAuEoOCxglOYGTKzQjL0bq739AfAr6ECF4DW2FffmbfwljxtsL+5G9sbkuj0hh4bvrzynoNMGBNpngmCTdABhr6e8fngWfZvQUGEMP0B7Jgtud6TG4COjtehddJawWvV9Wec4ppTCm+f6MLvT6fQNQgQz2xb815tS+21okdq66l9a94sbFqkMOJnYoNtmQH64OHrt+bcM8wNwPLFnTBqM7QClELNyVZ80jh8K7BkEGSKt+u6cKC5O2d8ONp4vO7Jqp/OsdOeTMlrwrTWK+0xQNKApa3+OFMFOuF4DGZNf2vqfybo6aetC0BrbNyR16brgKQMcLjdw7bjCew+nUJHaBEprzv1xSf+to0YZWw3OkSw9cXMUdMTBMKsgseOxNlEvm8tBIBvjL82fwCW/0kVtD5g0+HbB+pwsK6tYCAAgAZwIuHj7doE3j+ZxG2P/M2HSa9rVo9GJ2p4bI0fBrjshieq9jjjMmEdYIB1jy5+82T+AIAMlH4oOxg+vaUa3iCl8bCICN954dsH9h89+On2Vg+JTg++p0EIZn5EFNX7grL6fhFEeEFBAPS0hkFg+kIARFCs1UMDvXrgQf87a6uxoOY2KHUBfIXmti7EBWP2hRWFFB/Pb1pR8/6+X8wUItCy1oDnKSSTCp6vYTRAMBCCIZ1w/G07PQDdvkZbhw+jCaPHCDhukDoB8+yji7c9N9C7+3aDPYgM9P98Dcq8BaUZSuP5qmrMmjYGc2aML4jwL2x56sBbv9swUwgCIaupQRD40p6CUgaJlInSmxQEKRkgAwq7vbIygalT4tGoHECjZPdbg71/8K2e364/hqu/MApKXwc/yArv7a3FwjlTUVYy/CMxRms88dJ3P9y++5VZQhDZCW7g05lczsSBr2ePu8LRN0J/j8cELppRDMdlODHYkfhdD9+4+XeDrWOAGJBFqa77ofRuGxATnUk88PhmnDw1vKDY1d3l/fuqvz+0a//WmbaCI7ujk1XIADbthbs8NhCSicCIxQUumlEEN06QbiC8MbTm0UVVLw5lLUMDYPVXUmC+FUqfgNJwlUaypROP/fBXOHS4Pi/hq4/srv+nH97ecvJUzSVsR9oULISjOj+T7ylseTNAheNuAuJxiRnTixCLB2eFAr/H9u6k/OpQ15PfMYe/enYBpb2q0rRfVur5YBP44M2fm4mln7kKZaX9H4hMpru6n3jxe4f31rx1ObNhETYuMhpuUla7G+b0sA+w2ra/EwOjyiSmTy+C4xJcF3BcBjE+dHyz6PtLf90wMgAAKL595TXFXnpLTOnyzFMDt7Qbn13yKcxfcEVq0riKaCe1LdHRueGN1bXbtm+YCvJKJQfVmqSseT4Hkx0R+b4FJQx80dQ3OA02aYKLseNicFwDxyFIlyEFDkLS4keur8rr+OywDrpMu2XFPK3xCoAeR1AUOpMNsbXd0o2JeFFpp9JpR/ldFcwgIYLDDBzu4nIfjSOq3Zlt4MsMO5gJo0dJTJwYQ1ERQToEJ0ZwHIKQtB3G//xABU9BAQCAaZ/9yQWa1EYAlfZZk/P6bzzn6A0sAh9FDy1nbWdljbWFCOf4vXt5C44gjB4lUFHuoqiUgzOBDkG64V3Sf5WQf/fyxW/mN7k5UwAA4NLPPRFLKn0/gH/z0XrqdHxdBUuO21OcQf+OIFYgMPEMEBntR4caIoAIxSWM0hKJUaMkYnGClIDrBJFeOgwhqU069M0fLdn6zJnIUJCzXlNvWXFFg3zxm9ppulNIAguGPc/rBic4IWW4Txfl96CKk4IgHSDmMNwYIR4XKIqJULtBWRtoHdbcNTv0QhHzfY/clJ+/56KCHnaTn599i4yZB1jwfJZBteaEIDiCoyOtjrDHWzNHXWV0GUhhzwADjiQIB5CSlJT0qiv1gytufXN3odY8Iqf9iv7uyvlSYJkr6S8dSaMjEEKhc4EgQ2FlOMKyz6TEYSl4vcu06qkvbDta6LWO7HHHe652pjm6UkpxsyMx35F0hSNpQgRC1iUEQUoYKem4lLRDCmyXkqpW31k4beeis37ec+59c8eLmJjoyuKxJS7iUgJaUlKwajC67Ojr974+vAnseTpP5+k8DYP+H05wQmYEY/KOAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "WebEx"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://idbroker.webex.com/idb/Consumer/metaAlias/<your-org-id>/sp"
    audience          = "https://idbroker.webex.com/<your-org-id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "firstname"
        value = "aaa.USER.ATTRIBUTE(\"givenName\")"
      },
      {
        name  = "lastame"
        value = "aaa.USER.ATTRIBUTE(\"sn\")"
      },
      {
        name  = "email"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_webex_your_organization_my_webex_com,
    citrixspa_routing_domain.rd_webex_customer_fqdn,
  ]
}
