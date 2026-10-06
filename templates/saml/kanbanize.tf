# Kanbanize — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_kanbanize_customer_domain_kanbanize_com" {
  fqdn         = "<Customer-domain>.kanbanize.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Kanbanize"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_kanbanize_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Kanbanize"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_kanbanize" {
  name         = "Kanbanize"
  type         = "saas"
  state        = "complete"
  description  = "Online portfolio Kanban software for lean management."
  url          = "https://<Customer-domain>.kanbanize.com/ctrl_login/finish_saml_login"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAABcjSURBVHhe7ZoJkGVVecf/79639OtlunumexZmhgEcCMMStiigEnGLIC5VSUqNCRpKE7eAFWMlGmNVEqMJMRGNxq1KUYllVArjlCWmAqKUIoSwKDBOAGU2htl6mZ7X/ba75Pc/972ZcQRFaKIJntfnnXvPOfec7/t/67mvSzlFT+IS9donbfklAL32SVt+CUCvfdKWXwLQa5/w8uhjrWceWZ+4ssh5QFo0gW6wLZW8g/I8U0pbzsvcF1OK8kgMHtkXU4/s+5GFHnNZJAAyapfaV6hKr3XJ1c1LbJRwXVZsTA7S3gegT4JbDx7J3JEkenxxlHfRAMjVgawBNfNU+xY2aXpuszKYHqmPanzgDI3Vh5Djkt78Q+XHN3ePNSk/DAZrQL943MwfCdJjK4sCgOVvkrZNf0033PMpNbLdykup0rSilMGFVkVD9QFAOFGrlh2vySXHaM3SkzRRXQFrh0kSJUJZgvW4jUrdHpu/4AC47Ji+Qf9x1x8rq3Sx+dXIsE7NlWL/7byjTpqonSXK0roWmrHSZFCD1eVauXS9JseO03Hjq/Srq47hKTM3Sq2GdQ8Vk9kn1cwfDoD7HxsgjxuA/tZfu+NNqP5tSvJxpXGbtqxuFtMK5uvKslTzEfdcd5FsglC7aax2K1WbttzJlSUlXXZipg3xhDT2KyrVT1N92QmsvspbHSqsmRjMOFElw7HiW0xE0J5ixsOWw/WoXx43ALZWL3z1TS+Dqd1qJ6NKKzPqdsbVYjRNYdDagCZkOSaRRFpAyl2ITvMYbamoE6JErPk019Gdh3T5xJ1qpNMgVFEL0MqDx6paO1mdpWdqZPSpigePV6e6FK9TCe7WOhPBegQKhUEdwRK3gU0AKgUP7FqUxw1A3/433vxqzeb3wtiEstIBtbuDaADjLN/COaa0CcwmaQmGI4iPGSujGWgKACRBO3Ltiyr6s/hmnV/dzJwBVRcAND6gtJkyFzYTS3tAHUCJJk9XvPZ8DdROVTR+DP2HRR9zZekEreAGDVQEpYQhf/pl0XzAzd//gO6Z/TTMrUTqMJMRGYgICQy2y1WcYQajNTQgDvGiw7Z5WgUs3+f4B9SadZJkRLX0gN43cZuWZncra9dVzSoAnQJoGwBSlZI62oEf6SSsa6a7aiRjGp44VaVlz1J15RmqYDql8UmewwdpmKcj4IR/9i20oCiLBsCeudv15e+9FmmtgOkSzJVpIRrDbJbwAcEEajAYoRE2AYqBQokTxpoWEsBEnSXaM9DVJe3d+q3Br6vWWmAtwmeEj2C9nLASsWbOdQbABiYtoRmtsgbmAQeT2492NYhA0dB6Da04WwPLjlO64hTV1z8vyD5wbAFwszgAsEJeSvTZb2MGCbY7MKduayUE2hmqMAETTJvBdJOdO0gh4Vr4gQ73bZiwluRZVdPxEi1pPaj3j9+lo5s/QC2QXYYJdQCIuYEJVnNulWUDMIF2JcAHOCVAydG6KuYSNdtCSYK/aXdxvOUJ5SdfosnnIKhhTJW1Cp+xCKUE6utXPV9Jtw1VdZWiBCKRjKWDzbvtu6oc5oPLQv0zj1kaTpMNRqmskXRW+/Ix3dBYG2zWAGdmjjXsUyx5p9e27Sxom+dg48xLYcpzDHdimy+TacSYEYnYUHVG0S2Xa8c/nq94+3XkGYX/WoRiJSpp/doLIBJP3xoB+XnCXUGMR7ngHqLC2cD3POW+CNOgz87U9pqopjIgZtW6vrIwqfkKOUW3g1+xdvghSCa8Is5Qy5Y8ahbjW0pdpN6tUavKiSApmmAn6xp1yCsaIxoeWqrJ7m7d98n3oBd7FgkAULeXnais1NjAWsXE5ygegh0kAQiuhPgwL7eEggYUsGG1zOvXMt8kTdkoB6e2dnZi3dtabvNnsoGEQZiRowGVhSxyIwcImAaaUHKCwXXOmM3NgEVhboQDxC+0AacyrtGZ3Zr95tWLBICZi6aho6sT1z1PnfYCAu9Jy/RRndlb8r4u7k23JW/GuWdu0e/sERPCc1SqVd1yYEkwJ2EiZlApzMO0TcItCIYatMog2TRYCGXzYt4kgGRyOjEdCWsRfUaSpiqb/3NxAMhRYWXDXFR0ypoX4g3W4PxmAgGB6Z7qF8lQ0Wcig0ZzYxo9kNhPIOEkIktA+gm2e0u2mkmD9JM0oe7hGX8xP/gOA4AJhH3CGHoSQPLaXhdg6S8BQpnQY6ccdTCLasR+848NALYLmweEKcEvR2Txpbb2zt2mqAKxnPxs24XzK1QxOEBLnD47qMAw/fbFBDeIrTAfCeIXkpiTZTyonU2RHE2GjNImZFXnppA6WpDBXWIwnB9wb2kX0cTj9IfqkEkH3DtEVzpzwa/kS9c9ljBoSbYgGmdjhgh/+5P7dMfWL2nTttu0pzml+uAo8R9+Aq2EOLQjQSRm0ifEIg/AAWKjLbw+JBLGCJE2A0BpI+kO2tQukSjhED+VXa+TultglnAIE/b0wQmaUaRpqZsuh0VHk4howi3Pkv5wnwQHZF+AkQLIcKepbe1hTb7jQ/T+hBKEXKx9WCHgkdh0S9t0z4Mf11U3X6LP3PAmbX7oOkJOW6Mj49isN7SjM0NoA7fuMQChwqSTFxPl0Oa2CI3+eE9Cl59BG5rBVy9HqgrAWLU5aaPmlrZ1jz5/rOf0RYCRWeq+59pz3cY8SzfXJGLk6JV1Z2lozbPY+RGKNSYrzUKMJVeUTrpP22eu1zfu+VN9/rpLdMvWKznY3K/qKLk9dkcGRPUBx64NQqz2IGgfEYTQK/AatMPRABUMBLr4tZlDVwCPtuSkqVzWAxymKnk1hLuMuSkP28s7oriNzSR7gTxLsabXg27T4fvAC11pQiKEme7ujGrpxX9OyGWPn2gCHinNaV/zJm3dfaN27vmuDnT2Ki5blWvqljnJBaAIdYSfHACaEATYqDAS43GiIypPSDQxMJVh5y3mtyAsxbGFCgmtjPWC5xrkeVSV/hYETqEKb1m4U5e2bkUDLG+QozUYZj4yyFZ9a4xbO0BLDNHnnBeUDoSU3BnkUDvRnsa8hl73Dxo5+2LGrHs9Pk1gzhm70BNpy3yim7Zeqxu//zzdtOkybQOAVr5XlQFWLw1JFdsgf94MSTlCJVDCCtBX2GnwxqicITbOQQq91hT3x/xlSXk9yA7PBs1hzJmeAQl0eZOuN/KeVGuBgWcdh8fgGOkL1/YR3MdsVmsjgqSl2Xl4fM7rA/Otthf3EYnCIxCP44mQEKnn31x3r8694lq98qp79V+tZ2tgINZgmdBkFQvrk+aaSD9oLw4BPuF1ATD1iwqvGWzaXh0mIT4z8eaoxxh/xRq+9gN+hvsAJFHA/HmS5zmRChuHCEAPTBWmY0ZJbDh4mWED4eTH9yXfs1juuE9S1kxO1rf2n6BVv/dXIk1RVIUO9onC5llXhEXtI04+7SPX64o79ilbsl7tsdP1vmufq2seeIOmxwbJ8CKVO6vJ0ngE7PwSw+of1NI2iP37TBAkEjwx/QFj23WvL7SuHqP0bNdisAcvEhizXThMZ3H25IQJbwogOEivH6qfYR36UbrAcGTkAkjoEmZQHhzW3VNLdem/7NEFn92oJgckjiqsWQgpOFMb0YOctC780De0MxlXfWhQ1bjJ9iA3NKnPbDpFH73uRZoqnaF4sIF9J2rbZDD2kH6aCaTSd0rIGwli816aaj9teq0bQcVtIm49h+sMOzfdSdkzkC6S9nhOf5mrCYe/dhHzTa9BCs4taIETHzs7awW30axKRKNKfVyN7np99tq6/v4zu/R3t35V1dqAavWKqnYW4QziJlrQAhnXb370em0fXo5t27FADQS4DGQtjUc13Tn7Ar1540W6Zsu5SgClCrFmNQLF2FoACJaOJWkJWRMcAoPErRk4tRDymGeTKbSFa98HLfBcz+trTTE3TroarXIYqgyRvBgOe3lLvWDeYBh4h78Q8+orNZ+frmv+bUwf/Uis79w+qjd+/JNavn4F80yzC2t4b+jBlNL8r7+5Sf98e1v5yHLVO6g/aDqbcsiJUaUSNTXRFhPOcWltq55z1n06ffW9ips7tbdSVbfB4ukY2hGT5CBbGPY7wIS1GvkAPUiXNVL6/DrMGVkXW/V7AL/o6Gaku3DTzobCGd5+boGoMjo/r090btNT9m9if5sIdImQaMBJyPKM1eKRsMbMzhN0/XfK2r5zDqtaGt4fLH/pyXr9p98VXrtVEBR6FSDol9JOKHz6B29XNlBnyC8qu6hUMzhFO6+AWFCznJCHk4PgKGurO93S5NBenXfKJm1Yc0C1dJYEo0k464YQlphACPV7QQNg9O1zU+y5Qyg0oCHUBQCczgIArLVyA+C8IVUDMFbNH9A1ra9qoLFH1e4waxJpyhyUohEl3RHNN0e15f5Y37s11b59eHy0swsNWber6bWRLr9zI8/YoJy3uhwBwMfu3J6/7sZpra1HbFDSAgSbQRMVDhYga03wSwjn49a+SorPT7vqdJZw8juAv9ims4+e1lHr5rRucr/SzkxI1xNClN/INQGgFZgkJqMhBscvR7usy8G5ACMnD2COX4gkJD2ttK55CDyz+aCunv5yALLUHRd+WK3OMt1/V1k/uD/V1oeQabJUUZnQTEbkUFzGbPbMz+ldW7+ufJl9FN7d6t4zAG56LVdv+cYP86vu56G8oVrLhBinDkwz6IwGqYdEgpqYK7spCHeqCem0zIOZ1gIgcSAZzfbp6LUHtGL1nJYu63KkXcAk9oNqJeThXtLvBZ0UpQaA+0JjCFsOw1Gb8QHmjqhdb+rSfXv02ge+oy3b69q1c1C7twxoarZFUIhYe5kPj2ilX6uSi3DOaMZdNXZt1x9880qtfvoG+GEEoUchPf/xUnrFFzfnN7J4PjyjAU4wHaSRo0LhgIF3tf3bHCIDAYPF0ZfWeTf90M9c7DskPeR+qG7WQrotA7SAqi5oaKit4dqQakt2aGi4qtog6yIVJzkmuguAfqXdxfgbM2SB7brarDHd+IEuv2OPVt53QN2oipSJ+eQWjhTgGTLHrDTMvVtra6YHZ3bo7He+Vhe+/dUAb3Kgg+Ujvx57mFK6ZOPm/Gs7ACDeDwBtAHAuD4P2QixYst17QwPhpMJhiz6HwD4ADATHyZ+X7AHiwzekJp1gdibOaXC7sRDGIyfwfobEJcOR+Rgs7JaJ7GvSK9rQauhvN28hEvg1Oibo/bAB2zOyCGE1i2xm+OZqRZWdu5U+/3hduvGDIRfPqw6VjAYNeHgAotUTNR2ozDKpSGCc1vgBMxrenweTeITS06oSi0d2jjxnL128u7OZwKjXYNzGU4obqgxjFkO421pZcbWschWnWwFMZ2UVWGOROCYkVjp66kJLtWQOQA6psdcLvtzzvXkMCGlbQ9Pz2jkZB+a9fXgX4U1/SolecuwkXmUK4KucvGqs5zjrbMuAmHjjCz9eFUYcF4I7MaJm+DCmi3zfjFtzjA52bkKs7ti2VZ7ZzLFW0ZcT38MalhTLYQZZOcWXDagSt/TMGbSyTYhDO5z6OqQ63Pm0l5RL4efTFjl+zH47O1N6+7c/j9EVMss4sDm1CIJ5BOm7RGetHNJRrGTvW7Xq+cWhiS5oNZRh4sMWhsIoxPvKyUgAJNwXY5aXIQytCeP64JIBHLcs4QukW3W4hejn76tq9dwcWrXAukywFjDV8uiiDoPYNsFY45jHtgN79dLPvZvz/XAAJSzl6j1+SgGbTBcct1xxq80GeGBU0MIuRMK1Z9H6YyIQbxgKfaGr5wcOzg3f3BbP9N/fF6e7cHXw4+UMSWDeY0i3Q/o9olldtHuvxqCpXUPicGJJBoa4GcJZ29xahO7W3r065w0v02kveboDlvz+NJii5z6KAqBdvfnsY7Uwu4WrJYQ+hxRTZmm6moFwaxp7Y+E2fIf7wJibXm+451N0HrrujfWuwt/BgX5B2me25nXy1DRa6LBI9ofmGIS+VK0QXWv1QlOlZ5+kC997WXgUf1jM7c17NCVy6nrcyKCOXzuoOZIRYa9+FVUorltbvVe0neJV6HaIDGrer73dTGRxgW/ChqyyXfsN9vAJ0j+BJV2HSYfWGuqHuMIaOLSYaEEIXUUS9bskN1VrEtzEfKyl3tA8B8lj/7U2UYGHf//aj2MyDBSKF0gJYPVp+SklGvBZmvLbJ61BhWZxHsWGIZRRTZ9toh/mgjMMle1DS7/pY5rd448UOv0/AAln84WpKcUzu/SG807SMzasAo2E9NlMFwD7FzDFB3R+94DO2D0TAA7xOwBQMF/qoT6E3e+andbLr/2YXZbK7B98VkCoV33/KAr+JbCo151zrJbN7lUaD8Jfz9uiCYXf9+b+ClMRjF1ar8N/rmzqJCn8NgdBDUBsNJuKsdHzJud11V+8XDNf+Et9+DW/ocsuOpXA/ZAq7X5+znc6qjUctP7w+01VunNa4AQYBcTNvCu0cFtm/a3Tu/TrH3mHlp22NgjFLDxaiR9ZSnlKxhB46erDd0zrbV+6V8MjPhj54xhrSXPlpCYEVqaHxMfSR0PoQqmVEI87TnpgIpvapVPWLdernnuWLnnWKZqo9amDUiJNh6yu9vI/ArQNpDvz+B6b17w+vHlK52zfoTbHfyeHFbLLaZ4th1fh6AJnlc5cQ+nvnKsXX/lOMkZoI9w5NwiCL9D8mQo+ztGVXB7PUkcP3/rvd+vKG7eoNLYStWxzBndWBunYbvDznk0a5p+7rfv+0XKhAwTNtlZE2O8zj9OrXnieThsb5ime9T81EN/9o6khLUgt6c1XfEH/9K37SIY4GaJRl+1+QK+5axZ75iSKOvm13jwxbaRFaweCu1hgj/0nLtfFN18JXYRLwmHNc1kRtxDk+LMWHD3w8qhfNcdBlyJdtXmb3rbxVs0ujGhJCiM4KC/ukJaSxDRJc7vdebXn92uYZy488wS99Zx1etpJxxarBknbvq24GAsP00P1jyAhPdK3Nm3ReX/yOWnpsC7Yu1Xv/e/dnBoxHOhwImabtnOz2s9XOKk2Ej0wlOpVt/2rsomh4i0QC8eA5zXtcB9LAYAiGhuEoPQwGJOzQ4M+eecD+uId92nTDs5zc7PB6fkHjOVLxvSMdYO66NSj9KIzT9SwxYMzDSHOOStsh7c09mK+tW8Izqy4zdKmynFd6y5+t8b2zumKH27XMXOp2vGC/CLOUaP/M3pC7j9HdNi7Z59e8OX3a+IFv6Yy2pj65zcUspeQPh4AvE1RfOF/P4uwU7/mUuw3OcXiTapPXWTwIlgeKphCM2qSRXIet3lYi8JKJq5vlDYeOzznFKhsyvE7ruiDn7ha4+/ZqGc0p1hnUO0y6s/hqEMOi6MPfiDiq7Fnr1a894065S2vUIr/MShlfIJ3WVQAXBwBDjIQdJdrt/yFL8MeGHExg+7j26p4xOkjHKZ6M4via+bZlcBktKepbx/1Yk2sXIKqk9X5PSOcczyCDr8cAXh8UPOcp+iir3wgOGT/phhCY9jv8QPQF9HBEnjly9ULF7/Z0fJlrMJ1QMO1V3qX/ef6tdd7WD0EmH+x1Yq6onM3BAfHE8Hr+8pnBk7/JDu5pibqgXm/k/AYeeFBZvtb9JrHVH5MA57YAtPmrlfs/bvfe1DffeqrtWJ4TPvxt1V8sv1AzPF2y+oBveTuq1UaKJ7pReGAo/OOxSiLtMyjLPasFluvOpwOnLxaR3/pPdqRc1zeNaepRkMz8w01XnmeXnrvNcrNvIUP42baR9zFYt7lf1cDCi0+VGwKfMLP6e1MD339DqVT+7XmRc+UxkgCGLfQfQQ+yHRPAxar/HwBOKw0GfRvj7iDYON+txD8D/OroBAA6FP6/0kDwj3VP6eTTYc0wky7eii09AWH16fU14tUfv4AHCZNUxIYPbL4OZf+2CJS/AsFwCOWI59bxLKI1vR/szzJAZD+B/iYyPHzBtm/AAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Kanbanize"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.kanbanize.com/saml/acs"
    audience          = "https://<Customer-domain>.kanbanize.com/"
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
    citrixspa_routing_domain.rd_kanbanize_customer_domain_kanbanize_com,
    citrixspa_routing_domain.rd_kanbanize_customer_fqdn,
  ]
}
