# Spotinst — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_spotinst_console_spotinst_com" {
  fqdn         = "console.spotinst.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Spotinst"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_spotinst_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Spotinst"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_spotinst" {
  name         = "Spotinst"
  type         = "saas"
  state        = "complete"
  description  = "SaaS optimization platform that helps companies purchase and manage cloud infrastructure capacity."
  url          = "https://console.spotinst.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAYAAADDPmHLAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAAEE0AABBNAWeMAeAAABOSSURBVHhe7ZwJeJTVucffQHayEEIIhN2wxwWLStVypSyCLQW8KNa2VGu1aJentX262LpU62316dXbVq0WblXq3oq1vdqKgFisAoILi8gaIIGQAAnZ98B9f2e+L5kZJjNDCBD8zu95BjLzbWf5n/e85z3nfDFHFbF4lm7O/xaPYgXgcawAPI4VgMexAvA4VgAexwrA41gBeBwrAI9jBeBxrAA8jhWAx7EC8DhWAB7HCsDjWAF4HCsAj2MF4HGsADyOFYDHsQLwOFYAHscKwONYAXgcKwCPYwXgcawAPI4VgMexAvA4VgAexwrA41gBeBwrAI9jBeBxrAA8zil5R1BxVaPsrmiUyoYW8z0tobsM6ZkgfVPizHfL6eOkCeBfuyvlxc2HZdXeGvM9Vm1Nt5gY8/cRfWTzEX24HJVLBqbKVWMyZMLgVHPMcmrpdAEs3Vkh967cL/Vaw8lx3SS+e4zEaMXzGPdByMD9rbHlqNQ2HZEkVcgdl+XI5LPSfCdZTgmdKoB5L+XLlkP10jOxu3TvFqOt/KgRQmOztnh9DIIAKjxWjyOORK14/m7Rcw/Xt0heVqIsuvIsc57l5NMpAqjSvn3aU9skPtZXoVRmuVZmTmqcXHtOL2Pe+6fGO2f72FfZKCv3VMlzG8tkf3VTq2gQTJNahdfnjZAe8d2dsy0nixMWABU24fEtkq4VSEuuaWyRBBXBr6cOkPP79XDOCs97RTXyw6WFxi/ASmA5cBjfvmG0xKmVsJw8TlgAk57cYvpzKqq8vlkuHZgiD0wb5Bw9Pr73WoGs3lut1iDWWAF8xmVfHekctZwMTkgAty/fK/9SM46prlCTf8XwNLnzsv7O0Y5x54p9siy/0gwVsSafHZom5/dNlvf316qP0Cyxqoq+2rWcl52kz+vpXGXpKB0WwF7twz//zDYzlseTPysjQf44a6hz9MS47q/5UlDRaLoSuhi6BLoXtzPQx5nfahqPyOW5afKLSf2N72E5fjosgFte2W08fkz/4boWWfuNMc6RzuGCBR9JZlJs63CxSSu8hdiBqqC7/uPWd4Oq4XBdsxHBzJEZvh8tUdOhZtOoNfFOYbUZxjGk++aFfZwjncf8cX2kWls4XQufvKwkmT4sXSZrl9BPu4ADNc1Sp8PLBE0DVuiuFUXy0JoS52pLtHTIAry4uUz+Z1WJpMR3MxXxwc15zpHOZcwjG+W30wfJ1Nx055dA/vudYnl6Q6n06aGWQr+XaFp+Obm/fM76BlHTIQH8YEmBccowxyN6JcpjXxjiHDn15B9ukGv+skMy6C70OyJYp90RPoMlMh3qAuj7KWCGauMHRDfWP1ngfD53Va4c1IqHdB09/PzNfeZvS2Q6JIDi6iahgRG4YVbvdDNMrdDcvAwdMRzV0UCMrNhVJY+uPWCGkZbwHLcA/rG93AzPsP9M76R0kXDt9y/u60w3xxjf5NmNpXLxHz+Wu601CMtxCeCHrxfKHW/sk1Q1szo2aw3ZdgUQZe/kWBMsYmhISLmfjg6W5lfK5EVbzFDScixRC+DHSwvlrYIqyeoRJ0e04g/WNsuozCQZkNZ1FnUsUGf0V5MHyMTBqXJI00eACgtF1U98cqvvJEsAUY0CXt9ZIT9ZtleHW3HG8SMO8PisoTI8M9E5o2tCRHGnjhIQAfGKs/skyu8/f/pGLF2RqCzAnWr2Ma+s5KnXyn/rhtFdvvKBdQUEkNzFKasKa+Tjg3XOUQtEFABOHwNsIvGlalYXnsYxf0f4g6YXq4V4WXPw6LoDzhELRBTAazsqzHItHL6RvRNljLaoM42bL8gyXQCh638XVDu/WiCiANYX15mVOnVqRmeMODNDrNOGpZsJq1rNQ0PzUdld3uAcsUQUAIs8gGVeQ9IDl3WdKeC8fvW8TPliXi+5fmxvUUNwXKzYVamjiibn2yeLDk8He4HtpfVy5Qs7pIc6kMxMfmd8H+1OOn/m83RiBRCGK5/fLhUNLeo7dDOBJBbBvDc/7xO1+OSTk5MI3P/v/ZL3yCYZ/rsN6ghWOb+Gp0y7PxafAAtTiDYigq7MO+rkjnp4o5lKv++tIufX9vGMAIhispAkMznWjAaiYfbIDCmrI6J4RLuAFsnSa5l46sqs1Hzi87CaKppl9Z4RQFFVU5vzF6UTeOvFfeUb47LMnMLEIamyZF7XX6FMoCuO9ZOO5YqEJwSwR4d9ZlFplIXiz3x1+p6Zkyv3Thrg/NK1+dis1XC+RMFxCYD59QYWAZwgFdq38ukomGNp3WkYme1lvgUsLnj1pxrKjmhkRyHPka6nq+IcV+isl4xExFHAog8PyUsfH5Y9FY2tJvSIXjG4Z7xcMjBF5ozOkFztF+9YsVdWFdSY7WHAXfnrz3NzzWQMK3dZ808kzq0LHjxrZE+5a2L4vQRsI1v43kGzB4H9gyhc82ri+6ThurGZcnafZOdsH0T+CGMXarpXFVarQ8d1vlVMwzMTzMoh8kGEc3DPBPnep7OdK0VmPrfdFKa7mxkQ/pTcdLntM/2cX0RuW1Yoa4tqW30K8kze/nbtcPOs9/fXyG9Xl8j6kjpTdjwvNaGbzDs3U24aF344WdvUIg+8U2L2SDDlrgMRcz1O6dl9kuQzg1Lk2nMyTRm8uq3cLKP/sLjWiJ1RC3GbrB6xMjg9weSRskZEj80IDOW3KwAyPPWprVpg2mLiu5kMaNkZyCSJIWHTc9PkHjWPP1paKKu1oM1iEQeO/3LyAH34Ubn1tUIzoURhuQrl0TVaUXx962ujzW/B3PnGXnl5a7lkJPqudcUDJIcpX1YNX5CTLAtntu1L2KSFPvfFHfrMOHOdvwVw5waA6/O0QJlKdpnyp63muDsCACaULh+WLnf8R47zi8j3l+yRdUV1AS2NPN83ZYC8satSXtlWIRlJ3QP6ZCqmSitiiFbM81fnmt+CYZRyyyt7zDpHVjgB5c1fZIMl8iyBI80XawNgZMPKaJboU/kuPIvKB/4lGrr+lsAFvO3awhnPbjeFxp4/bkTLw7NkQoUNGeSZOYIkx5yyMveQtnJaHqFjPlz7+7UHzCISrqvS65inJ6pWZyqe1TvdTUHPeWGHuY8/N/19lyzNV+89Jd4UBPfmeoIypfqsqgZffD9bM7+ltF5mPLPNuZKXULCp5Kh5TrDEqUzu4fu0mDC3P3j7pbXa3ak43LxgDfz0YMjRdJEXQszueeTz528Wyco91WZlUqWm0ZfnZtOofOfEyu6KBlmgVi0YJtzm/98es7GW1k1Zk49sLd8ELQMaDIKiUbr7JhEgZROUDSNu/3z6us5AQloA1vx/958FZsiEKRyZmSiP+pkOd0MoCZw+LE1um+BrFZgtTPzdWgBm1ZCCeMj8jZ/qrZ8sx0Iclf9auV9bSLmKJNacR0E+dMVgo2h4dkOp/GZNifRS0ZFEKvwnan6vGtPLHIc3d1eadQpsI6OCaH2zR/WUH13qM9PsWiZo881X95jVwxQYv90zKUfG908xBUTuaTRUij/l+rynN5bJnz8qM/lESFM1r7c7eXXh92X5FXLf28WS6gy7sB600Jt0BPG1sb1NI6Hyb11SIJsP1ps0cQ6Wa81NgRtq7tJu8s3dVeYaKmy+3mPeeb2doyKLWZKv3Qot++HPDVbL18Pci3Q8t6lM/rS+tDW9XxiRLt8en23EgQlgiz5DRH9CWoAPi2taTQ8X/9gpUBcyQISMBDK54pIc110uzPEVLFBxmLuFM4fIty7K9useYuRnakrP65tsBAYUHhlweXB1sWlNwFbz+6cMDKh8mDgkTZ6cPdS0Gp6Vqq3imQ1t90CEVDoV7L6eokXPo7KxPIgLkQdXPvTUY6OzEltNaHtQUeP69TDdClAZ5J/t7WyYcS0keWcxSrMe4xwEy60JN/uzUbsu0kx+uI9/5QNpRji0arfFcy/G/FgF7g38y7OpK67ppfkMrnwIKQDnHoYELb2nN5Y639p4cNpAk8HgN3rQP7lQJjghqDQU37mojzFxQKbf3eubquX1MnQ/ZAwLQgbae3PIqN5JWlFJpjDoUiiEv6rTGg5EEA3BJrU9/PPMn7wLgaVzobhUnTdXLOQxOLLoCpW86GF1MgOnr9kse6s6rN/Vlj0oPfAZfskwBH8PRUgBnJudbPodQEWsCcAz/qd61awHBKaGv64mPdK7fcIVdp567hQep1DZjc6939uPZ+1LGv3wVHU0wzFFxeFaEnyCdUW+9xKdLsIVPH6BqyvNsjHL/rDewhUII5Wb1R+gO+YdCi43aLnTveQEvXSjI4QUAJXaS71X+hFw19T9QvvtTy3YLF9avFOWqCg6AyrMZ7B8LaKyoVmKqtqGnFiAoRH2HrA3wW2tOJSFXThef2x7cDLqwPJ2fCbyjRVg5LShpNb4MeP+8JHZlbUjqNs4EUIKAAh7MrQ4WNPk8141nThbeKMl1b4xPcOlOsaJJwD7/f3LBPXT99E6gGPJ2mrCgXfsiojrgr36Mwl8ksVzc43vRLCMrNBIGBIiBqzjVX/ZaaxCZxC2ZF/QcSqrf8/vl2xUyeIQTDb9NQnC4bj8qbahV0fA7LttgNaBtWGY47YUjrmWqD14CZXbkriO4emZzCjtBlbfOMaMerJTtMHpiILRC10LTh3D3jX7quWuTtj0ErGk8NQfnDbIzIM/dMUgY27xQoF+mj6alz11BDxs19Th9fIdb7l/Wnxr0Ilxc6QlXBx36xyfY9AZunIpmFmjMuTZObnywfwxOrTta2wc1hip4x9EcnajIaQA1qg3Tuj1+U2B3v+FOnbGIlyTl+EbWyqYp4IwFRTT2r6PZXl+ZevQkArP7eXr6y/UUQMZBYIcS3dWmr/bg90/rtNI98H1/rjWxMVNERatpLprLfUi/kHZ+zt9NBDEwPuS6CJoKPxGnsOlv/2SbyOkAN4urDLLp3npAk5ZMAzJmhyvG7NEdxAKEhAq+uTywKri1okZzPyMEb73AOCEcl+6GCwA3Q/r8kKx9VCdmQJFR1gRglEUlj/+9U+a3BAv4/+vvJQvs3SE88QHh4zz6Y9rVVz8Q8P+0GX5085pIQl+xpMfHpKF7x+Uny4Pbd4v6t/DCABR8z9+gYsbA3AJ/BaakAJg7E9fTCCmov7Y/vfhdw+Y4aEJ9GjfxFx5KKg8HLIvL84PUCrODSMJWivnkHDOu35slnMG3nC2CQAB6eA1ckTl/EEU17+8y7QKWgSh5nnnBgZOwJ34AdK9YJ0vBLuzrN6EXAlqPaaCJ/7uQrRzZ1lb14KBOaAOMdFGfyiDTQd8c/DAv1xLcCoS3JtnNDrWDoji4Wy7i3H9odH9Y3uFsbp0dWzLo/xcUuLa8olj/Pet5eY+NIp7VxaF9KVChoIfebdEXvjosIkGonpmz9y9Ae/uqzEPxRmh8pmN+/XlA50rKaRm8/IoKgWoXDx7Ile0DJJLIsmkG/HiRZG/mT5IJg0NHO/f+LddsvlQvYnwAXFwMsF9uAdpYFxNDIEuifNe/fIIc64/zGayJMy1VMQMiNHr4020kP9JH6KbM6aX2VL2QXGtER7RTReuw/+hwN5Xn+h3a0rMNnQCVW7oG/BrEC9iWfX10SYK53KPOm7LCfU6yqJyOHecdltPaPc6W61RteaFWmEiabD6M4irXO+1VsveiF2vY4s+oWD/OAzh9EmLtpoFLDQIRGLSq/dCJ2PVn2OjjD+OvgOhNVKgFDKVjsKpeN4KQgtCFCgcZ8u/8oOh8gnxUkFcR4HyIRPkn35+X1WT/GxCzjGVD/+rBXLZ4BTNbNuOX0wek1KENal8xEXLzM2ID1n58J+jM8y5zMkjOPpO7sNkFddjAagEd98DacpKjjMCIw/uhzgFaadcoKCiwaTDDcG6HwRK5SHy4JiEHjYf91yewbOKnPMQIuVChZEmynyNlv02HfvzbNLL6iZWJwcH4Zj5vII9EHod96aeTBhY00Ial+yskI0ltc7ZPkIK4GptBVePyTCVxZo4N/bM2JQWzszf3Z/NMW/mCAcFhblaft1IM3+NaOjPiS2wu5h9e8vmjZBrzg6M8fvzqykDZfE1uTJe+z7CxjyfiSEqnXS5U7lPzA7/fuHl140yrYyWQ0shDRTwBBUYQ66ZI9NbHVIKH+G393GNZqTzjBjMmW3wW6hzEQ3cflmOCRfznfKizCl78sp0Lq345S8Ok1vaeTHX/VMHylfOzTTnkkfySpnlZiTKK18aIedkB66biGpZOIokEe70cDj8uwAyhohenDvMOerrHwHldxTMJn6Kf/93PFCw9JH4OV0drBbWmLRiRY8HfC2tgnaddIjqjrQMKjRS5QNmz59gz5mKP5HKB/rljlY+kJczofKBWT4mlo638oGp9nCVD1FZgGjYdbhBHtehFH0VJhprwa21y5LhOr5nOfUPLunrnG3pKpxYU/TjIR05LM2vaK18wBPFw96qoli0/pCwZMzSteg0AdCzM67HUcNhcT8Mt+qafI5OtPPwllNHp3UBrEjF42wvWoY4WBhC/2vpOnSaACxnJp3WBVjOTKwAPI4VgMexAvA4VgAexwrA41gBeBwrAI9jBeBxrAA8jhWAx7EC8DhWAB7HCsDjWAF4HCsAj2MF4HGsADyOFYDHsQLwOFYAHscKwONYAXgcKwCPYwXgcawAPI4VgMexAvA4VgAexwrA41gBeBwrAI9jBeBxrAA8jhWAx7EC8DhWAJ5G5P8B5raZfX26B2gAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Spotinst"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://console.spotinst.com/auth/saml"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false

    custom_attributes = [
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_spotinst_console_spotinst_com,
    citrixspa_routing_domain.rd_spotinst_customer_fqdn,
  ]
}
