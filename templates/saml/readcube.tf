# ReadCube — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_readcube_app_readcube_com" {
  fqdn         = "app.readcube.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ReadCube"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_readcube_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "ReadCube"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_readcube" {
  name         = "ReadCube"
  type         = "saas"
  state        = "complete"
  description  = "ReadCube Software for Researchers, Libraries, and Publishers."
  url          = "https://app.readcube.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAA+cSURBVHhe3VsJlBxVFb1VvW/T08NMMlkMSggiQ8IOMRwimDELmzk5LEcWZXMBNCgKJlGCSHYhIIsaVAQPIgQNMUMyJIQlwUlMEAlGQnLALCDMZDJbT+9LVfne7+qZ7umlqmcmnKP3TKWr/v9d9d99y3/vV0fSCPiEoKka/rRkE2KhBL66/FK91RgP7I3i5cMpqNBwarUVP5nogV2W9N6h4RMjYP1DW7D16TfBj7ty4UyccVGD3lMaj74Xw+P746iySnBYaLLUllCBSFrDSxf4UWWTMwOHgKNOwLbndmH9w1tgtVthIyn4c97zN+m9xbHmwwTuI61bScseFlzK17ZCltST0rCJSPAPkYSjRsDelv14btFGJONpuLwOSCRMb0cY3/zllfj0pNH6qHxs70jh7t0RoWGfTYI8QPBc9JFwPpFgHzwJw07A4f2d+MOPX0DHh93w+F2QLZnJpZMKasb4cetvviKuc3EwouCHu8L0qaLaLsFSRvBcDAcJw0ZAPJwgwdfjvR0H4Q24YbGR7ergRwQPhzB/3TdQPdKnt2Z8ef7bYbR0pFFDgtsGEdiyJGwkEqoHQcKwELBm+WbsWPM2PCS41W4p8NlkPIXxZ4zDNUsu1luApXuiWP1BQgjO8+7/Dn1qCn0wgeamNhQShkTAa7/fiY2rWuBw22F32QoEZ/Dte9vDWPa379KVjCf2x/Do+3EKbhJcBQGOJp/qBtQEnVO7vYY+rHRuPEWFntOTrJyEQRHw9uZ9WLNsM63rKpwc4Mr4bDKaxIWXT0Li2vOwYFuPIMRLy1rR76hJIqCHZkXMiGnRmmevzVwfJRIqIiCVSOPBa36PHvJnt98JWTZ4CN86Esfu+76OA50J+EmZpSM73SvRnjnNjhkkCd06CQETJJi2lVBHBAu/+DAFu6QIcobCE9RIAuqcyfgopiFQdlmjdiVKnyRg7hhxTs9JdlAXxQWRCpUHryABiiszXgsSEUSeAUwRoKRVLJ39ayE4BzkzUBUVEiU9t916NrqCCWH6ZZEO0T9FBDzKJJgi4JEbnha+nl3TzSAajGP2/Onw0vl9p7rRFqdMvigJJFA6rJ+WEO4okmAo0XtvHKLkpoPSWI7G5qCkFQRG+3HS1PHiesYoB5ae4i1BAl0rEfo0EGqIJHRxEVEEhgSsvudFeKpd+pU5hLuiuGrRhfpVBjNH2YuQQI9P9WY+S2k/F0MgYeaW4iSUJWDnX3YjRhleJaafSqZx3OmfwugJI/SWfuSRwA1amoJFXPSZxjCTUFaydStfhdvn1K9KgUTRFcqaDXdGcfXi/oxvIPpJkKGlgtRCUzCj/VwMgQR2h2BOTChJwMZf/RUWqyyquFJgS9aiTqhhF7S4TdQDk+dMMnQZQcKJYbSm7MRdhcJnkUcCWZJJEjyUhC2hNDyLogSwJl95ciccHrveUgjhxkkbnDethfeex2D5wt9QVzMOc+Z9KTPAADNbb8ZS5+/Qph2Tuddg0EdCJ03I2BJYrriiUf3RL3ZRAp65uxnuKifdv8wNYw7YZ7VAHtlNFuCE5ezdOP2Rg3qnAdq3UobYi1m2V7HE/qthIqG8OySpYGqLa2ist+OHJ7n11iIE8KbFWxvfLbvsaQrl8r4IrGe+S9meDVYaW11dh7//ew02v79CH1UGe5YDtio68WOW9SUsHlYS8t0hTYK3U9Ad57aIbbRFkzx6TwYFBDy14AX4atyltc9+H3HCdvEWyEk3vLUuuHzkyxRX3LYA3ml7oTwJh1ZT5KeiR+T2DD8uJBKGzxLYHdJQNUkkQA6qOp84x4fH6ah1FIibT8AH/2rFh3Twvl0xcHob6UrAdUIXfGcdoRUiQ1TupN1UwpYl4b1fAlbOD3MxfJagwYJQIoqo5MY9E6vQNNWPk6tLW3MeAX9cuF5sagyESmYU6YkJYb/zi6/j+Lv2iejPWi+GfhJ+prfo2PcQhWJaVqVCTWQtgUloHQQJPD6qOdBB97netgFbpUZMrzeuW/pmsmvTXvL/iFj6suCoyYkQL2+z75iGBU3fgHbCXuw/+E9aUmz6qOLIkNCUbwn/eT5DQElU7g48JKFZaXwNGi078YbretxgW0POT1r/kJ5ngL79gLunPQK72ybKXG7ibSwufRtvnEzH58Vgxqod/EJDo9LWXFUYTXahof4SNB5/B/DSeYCjTvfXcghiQ/pL+FHyW6iXOksM15AiwbtQhXPkf2E5keaTeD+B3YuUmKa1fuyXgQnfEqNLQaj71Sd3CKHZxHnTgzc8Tv7CBCzb9r084XeTRhNUtpoVntFnCR/8FgicRfPmpcoI/e5QzBIUCnCdmh9+KYxnHXfhF86FJDxXlLyy6BbMq4GzPnNeBsIC5k95AG7K3iLdUUw4+1gqZC6iyF5oqg+1XACn1UdE9buJWUSp4muwjUJjhOp+/r6REQgE0ZxuxJ3JW1EjhegrZJmwwYU47rY/jqnW10hQFpqDXM4NOTil6DnTNusNpSGtW/mKtuWpv2PMZ0cIwUceV6t35aPl0GN466PVcBREcPOIQkWDYkFjUifAFAkRtGuj8Vz6AoQ1F06UD+HL1mZqd+hHkZskqcY48bvAmNI1SRZS04Ovap85bSwaph6vNxXHytenwGuvE24yFERJiw2KjMaUbkWmbsfJDeUOIuSx+5UQnJHNBqcaB0CGqU3R5n334EDXNtgsle0LlIIgQSUSEpVYgkkkKBE6/X6g5gy9oTwMnTlCUXzP4WZYZaOy2DwofcI7sorNdj2RMFSBSagpWgSOMy08w5CAF95dICK5oelzfwXuIUiwEAk2IoEJMDbE8uDvk7Iw6V69oTgSWw+g4/I/6FcGBLSF9uDj3t2wyGWSHs4bEglovSFooTC0FGnB5Du+PhIcNPmhksD1RR3lGe4xekM+0ge7ceTiJ9B57bNcIemtBjHgt29chrQSJxmL5NJZwUloypX7tU/nkpNKaT8tT9xuAhFaHSZbajE5+BGtaFStVRpoWQR+qTLtFZpX/h6GGk2hZ+46JF5+H3KtBypluyN23ArLiMxqVtICWPO9ibZC4WlymqpA6+qG1tMrHi4RGewi4rBYMsQEqc/EyxOGmxKbllQrme8iIE6CVGoJSgwYd0WB8MGFL+Fww0qk3voI8igf5UYqXLMb+oRnlJxhhGprq5RzQ9YKTYwF0zpJeEUhYTOCDwQTUgkJfA9LOoxgzUTgtOVEQpt5EjjpISvFZ+fqDUB41Q60TvgZYuv2CMElZ8aFtWAc1Q/k5wYlZ1ftGkuuksgITocWDpP5dAofz2q8HCohgb0wTT7sZpnrzgVOWWaeBH6pMuk2cRprehdtpzyI8MPbhLnLHnvfPDV+pfft/rQ+i5Izq/NMQM2YU5CK9UI90gEtFu8X3ED4LMySoGgpHBs4m/IMfakdORU41YAEbrZQgjQmgGT7DLRPXYWeOzdAclFB56cYlBOI+ddpHPh8t1OQHICSs0r/dRsuu/0ILKE4FBv5NgtgUvBcGJHA2ueK8eITyf9zMYJIKGUJMl0HklDanei4fTI656yiZ6Rh4Z2sIu8w2PSrftKoX+WjYBVQDhxAZN48KIcOweqrJoYseOr7YSgWDValcgKy4N8SSA5HweqQogA2/pipmHHCj/SWAWjfCuyal6ns+PHeNAVfG3pWTUD8jVoy9SpI1tIVJgc+XpZHbr9Fb8lHHwFaJIIwCZ7evh1SDSU+tkzgkOnefDz1/RCRQKvUMJLAj+Zg+73zWvQRJdD+OvBPIqE+gNDj4xFeOxayj+KT208WVnq7i6G0h1Hz5BVwTDlWb8mHICC6YgXizzwDmQW3U+QfYOr9JAyvJSSSvThz7FWYPO4GvbcMUlvRMfN5pD+uh+RN0hRJcFsNdZTONbSkAku9F7VN1+kthZCTL25EYvVqyPV0Y5pYMT9XSfN8XHO/FxYiIm0d4JMVIBsT1GAPmbXDnPCEzq91In1kHGQvrUK8IWOrptYywpN1cdJT/ehsvaU4ZDUcouqy/O98GP0k+ETwHSoJcSWM8/YZ/1yWEX5sJ5L/2E9mT6sE/1xGaN4AFBSd0yfAOo6JKg3ZedllwvS1uPFb2nwSpEGToEoaXIoLn1u6HZF7yxcvjNDy12iO2VKcn1n+uUL7XTFUP2T8g2yxZvibmiC53RWSQO5Adc9gSIh7NJy/1olkfRWSzc2ILBqwBOYguOBFSD5jC82FFknCc8OZ5NLlAySjb9EcHAlkCanKLIGD6DGtFhz3jg0KzU8OUCKzYX1REtRQAtGn36Z5ld+CzwUnPVo8jaq7pukt5ZGXNWRI8JgngQRgS7CaJIGmhhhpf9qfXeIzuxMkB2qQXF9IQs931kE+JvP2ySy0Xkp65l+gXxmjIG3yN62rzBKIhKuFOxiTwBr/zB4baj+2iO/lguNQLgnp9zuR2LqfzJgeYhJamgo0SoU91w1xR6hidzBhCaz9uCvj+xwDikGQsHYtlHd2oOeOTaKgIfXrveXBpq+2hVG7/nq9xRyKEsAYbndIUX41absdrgjlASWfSvIGvAjfv0Hs4Eg5vzgvBy2lQGntRW3z9bCMrGzbvsxUKncHNvFiJLD2Fbqe0uxEwlnCQihD5OTIMqYO8b2jyJSNNS+03hUVRI38x1zYThqp95hHWQIYwhI85i2hgAT6i1Ohf84mKlF50IAn8prNew1aKATvT++C9dqf03mirOWLdZ6CHY/zL56BES03i0pwMDD1XoARvOQSaFFi22m8PS5qB8oWn50bRme9gvpDVlz5iDcv8ovHEqlaby+cN94I1803i/a2hgcgVdG6X2RjVZAVS9F3EvDedi58c8/VewYP0wQwgpdcSiRETJEgUZruiEqIeVXh92wFfcInk9C6umCfdSHci+7NNiO08nVEnniTUl6qSQaA6321IwrXFRNRvSL/R5hDQUUEMIKXEglUOpshgc1fHCwhHVqaavnublgmToJ3+TKK8vnvIVvHr4BMQSx33ecAp3ZGYT/nUwhQYSMHhuftVBYVE8CoiAQCBzgtGIRcVwfP4sWwnnyy3tOPnh+sR/zlf0PWsz5NoaDYE6Og6EeAcnpbQ+UBzgwGRQDDjDvwrTnAcfXnXrAA9unT9Z58cCQ/fPrDkEf76Ev0R8ENNhnVy2bBOeMEfdTRwaAJYJSyBHHLWEz0uW65Bc7rSm9IMDqvfgbpfUfE9pUWTaLqzvPhueksvffoYkgEMHh1UMm8OV8QaxcHOPJzO5XZHtK6EZK7PkbHRb8T/7nC/bUz4C+xeXm0MGQCGJElS5F6ebPIFWxTpsC9cCFF8v7/H1gOvYteQYq0H1g1p8//P0kMCwH/yzDMBP+/AfwXSCFtjzQ/ZQkAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "ReadCube"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://connect.liblynx.com/saml/module.php/saml/sp/saml2-acs.php/dsrsi"
    audience          = "https://sp.dsrsi.com/entity"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true
  }

  depends_on = [
    citrixspa_routing_domain.rd_readcube_app_readcube_com,
    citrixspa_routing_domain.rd_readcube_customer_fqdn,
  ]
}
