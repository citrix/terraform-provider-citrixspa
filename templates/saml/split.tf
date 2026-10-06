# Split — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_split_app_split_io" {
  fqdn         = "app.split.io"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Split"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_split_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Split"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_split" {
  name         = "Split"
  type         = "saas"
  state        = "complete"
  description  = "Split powers your product decisions with a unified solution for feature flagging and experimentation."
  url          = "https://app.split.io/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAABFySURBVHhe7VoJjF3Vef7u+tZZPOOFsWvZDDZu4iWBttgmBtxAFqGoTkockChSQamUNKnSpFKDEI0SJ4W2KYpTKlVKm0ghmxJkSqCNIqSGBgFuA6FOITZ4wTYeY4/H9pvlzXt3v/2+c8dgGpt68BsPLf3t9+599557zr//33/uWDkJb2Gyp45vWfp/BUwd37L0JlGA0tCpVFQcL1Rimv0kmKfkIuPRptA5rCRF5pSQ8lpmpbB5y7U9WFPDO02z6gHSfGo5iOEh4TGxXEx6Jdh5Cx4VMDKWwMuooBmkWfUAs7BZnpZPKaitj4vjuYtPfu8lpEEb99+2QqNmjGbVAywJjpDCR/Rzj1FQxl/960ms+Pw+/OOLNdTr9amRM0ezpoBX3M7yKHwZP9jbwOq/2I0vPuGjr+Ki5tmYdGocEPMzc07acQUwZ5FdfYfmo/OMbi5PN9GcUKAspuU1xsLPTmZ4/98fxB9+s4GG1Yf+cpEEnSyDl2sOKmgGqeMKsCiopYye+9SEb84laMYbTk7hHYerejie2fjUtr34yFeexc+Pl+HP6We2z6gwUwsMY7YqhKGZqgEzogAyLYtbFi0uxjMKEvBDi8vdLRt/9+QRXHHXbnzrhT5Ec5bB98u0OJVjv+rqxdnMCX6KOq4AI4QtF0/4I0MswZnZLZa5f9rbwtp79uHOR4DIn48uP0M5baGcteHmHJ+fJvDMy26o4wrI4NJ6BC6MYTo7PAq/ewL48Nf34ZbvNnAg7UG9Rhd3U96PKLNtsIBi3WFYnKILJP95KoBWk6uaSM14niqGM4aBrO4z7m189of7sfGeg/jpiT7Uqj48yWrrK0LguHyujMiqInQs5oiUnsAMwHkniQXGE3mRVlDCnBk6LwXktLZi3iHTGY8p3V+igRb9+o4GBrfswd/+tAG0j6EXTZSUCGnbnMt6qYVanCD0ErhWA5W0jdiqgNkCQ3EXVjiHcMd6lUH5wsz5w3khQWN9flEWlIpL2P5yiD+5/yX8YrQb/d0+nJE9SFpjFM6GX+9HpWceYoZFmhP6UjDbilEPfXqBg6N2hO5oAp97bzc+fsVczpYjofFd+02qADB5wSrzY+F4K8JnHjyCbXss9JaZ1V0KSRfPR/bDnzyKjG6fsCRmjk8l9MOpLkCUlzlmBBPJHFTHG7jl8hy3f3Ap+i3WfybE0PJREhbgkV/Fmh2m8+4FYlrp7keGsXU77VmqYh5aCF0LLaeOOktbePIlhLSqQ4Fc5gmLYRBlPPc9VLp7kKa9WLXIw5Yb+nFZL8tkPsEZa/QYh+HCYGE+sNgkzb4C6IoJeVDvdoqh+55r4q6HjuBo1o0+xQAtHHi8xySoZOaznQ1PHEI2OY6cWF9JTrnC4r92xB7Ar+ID17wDX9jUiyVagNm07dio0Op5HtFraqwmXHQqBPRdnInl4ux86ZwVwIIFnylKC//7iI07fnAYT52ool524DPznZpEQirJpSxrPvFAeGI/8ta4UVpO6yeEwgkVtGrNCqz5jYvBPAh7YgzrB3K8/229mMsnlR9Mo8T8kLMaWAwxoUNRkbVnQQHIxghfa/jjB0bw4E4L/XRhx02QMdZz1vyCSSqA5/KEjC7sUQFBQwqYQJrlCOMQS5YuwhXr34kyscA4f1u5+n96F8PFZe/wwSUWPrBI2Z/X+NE2CWfjnFNCc97OiT8dBZC27Unwe/cN49e6E8Z4jcIWDFL6YgBPczsxoYKsQo9hO3T8MNpjh9Hb5+Oqq69A/7wKc4K8PTPVw0ttMsHf1IKqAoIE8/0Am1d7uLyL4cTJcoaFYlD4QKgy5aqm3HaApoEDQlQZ057vmsVjqzSFXPUlyaf0SOtbGXt7/ouCMWTBUWzc8HZs2nwNaj0VsFjIyc1QCW+Si0XUyOd9zcNWeMSu48vPAHfvaGOE16VfIcac1tdaZtkO0TQUUKJbK7kFzAcVVNOYGF45oejeCgalCA8xL3tjQ/joWh+PfuXd2LiyB81GiCyljRkiapSkPAOg2AEGbobQtpVe2SvFvA7M8UvY2yrjU/+W4YHdjSL7GAVoRX6ENplvpEDzWzfeAE0jBDI8vC/Frd85jFqti8pQvNMlKUxmEhXVQIYmJgNcf0mGv7xhIRZWFARklOz9/ESG7+5t40haQ42CEAGYEBAAUqvsUxaPXhYRVxiLmzUlKCsDIXYfJvGRQQ/vGiDuMClZYIoDOViKtFg11GZPy6akN6iAOpkmijPWZFLiDDYjOI4p+I39uIWMSoSUN9gMw8ubHKAUV8bj+yM8dCDCEc5RJv81dsEuE2RCL0iYP5yMSpMGpkj+pXyRsTIELBnLyxO4eXkNl/b6XFGsUxUMpYSeZIunaQbI9NT1K1Rkfq2pKqCd3Lvv34+tTwzzInG/FRjwE1va2xOqa+Paiz3cubGO99ZbqIy3KVqEVonfDCErJZh4jfD8SSX2RNozTOBXHByO5uDPd9i495cNNCMCKnmSE3CcDDI94UVvzAOq9ACCE4WAilShRzU6bG4o6Mm2hcFaE18mwLn+ki7eS8imyhcZVtg6KnDAwUngO7uaeK5VRdXzUWOOEJAyYkxxpUObxaAe84xJOGa1cClomLiInXFsusjGh5dSwYLPZvtsxkIgx4/2tXDrt4fZ1PTxV2ZiVv1/sQegLS/Fu0UVJIxbC6M03MaBUWy9aRmW1cVYyDsuLeWw9mt7TO7t45kTCb6/s4khpxtz6AmpRcUyKWo7TVBYOEB9gXKsqQM28QN/ZHEZAcOw327ipuVVrJsnhKrkqNAseCrUcXbPmIYCwGw8jo/eNwp7bh095B9scVNWBHmDrKvFtNmpc0OcOcxshESCt14O/PXvLDXCZNo34BDhPQEbHagr/PP+AA8fYkPsdkOlXxjDJFsKJGBFJMBR8jyizannRDFzSJtOtbo8jtvWVLCgJNFLJkd4qhQGup+ZpuEvAQbn1zHYS4zOzq3tKI8TDFF/PktiKZMIcu2Cq+KbbBAt9VZr+OazPVj4+QP4xtMnYWsjhEkxo5WRMg7o+h6V9aGLy7hnQw+W9zMvhITBiTMVOvQEWtpT1SGF7CzlayLZj70X6m6OZ5s17G5K2BItW+xInp5TzkTn7gFqfc1Ojo9/+Nk4vvTjITS8AcYuy5bYpNsWlVxEj+C0SowqESFbYIdlyk8CvBzVsKrawL2bB7B+cZVjW1StlGmjJlyh3WQu8/xEjvuej3EwYtCws5RYinAtYHysQGGFfFxD2GIyyvDJlcC6PiqKhte6hGvFQ2ehc1YA7c7kI0sw2bg2GmTpSw8O4YGnh9GqLabSu+Azy6sZKhxLbsiF+d9RS6sZLI/fNfYFNtqjo7huuYOtNy7A4gqHx4wpxnnuRhzjknHN4+OxIwke2j2BCZ7HJXoc53PVKXIudZXi3uiZo5OgjU+sdnF5H3+wNAaO8IZ85eyOPg0PYOPDWDL7+1Sw5QqbVfCLsQh/dv+LePJQCV6tFyVb8SkrCaAUISGQqyYpZ5k0H46RBduxiyQcxR9cWcUd71mIbo4u7EWmhZIkmR1glGL8y55xPHrYQYvh5DIn2FrEmL8gPdeOYnx8pYP1fRSYeaHYopMije+ckc5dAQVnryG98RH4EP2IcPXOh09gX6sL1UqJThJS77R8KvPKFxSTcl5BZnWLdGp6i8fcMUpFXITjuOP6i3DzZXM4Xm2wAoodJUOrWMHF9tEUX9tJT9TegpK9dGDuFeyFUsBqB2t7pD8KnzKWNLAjHnAGUv8mSxk4RHdjccLXHjuGr/7kMBqlJah4HsHPhBE2NQJLEfonRcg6emGSY9KZy8rgwJl4GUsHYmz53UFct0BRL+uxnFIA9QNPHGMJdbtQpmdpj1GK/O8K+BgVsK6H1ueNmVcA51bzIkxgZURxtiqBjeHUxee2DWPbfzRh9/WwpBEb0FzaJpeyCgBF58xrVFATEWFwzLrnqpIwFYwyum5a0cIXNg/imeMpHnt2DEfrKo2egc5tP+RY4YJXXeC1CrhAHmDKXibrkgGiNLk3q1GR/UkvsAP82LdfwtMnfMxjm5uydEoJquMWXTsmfy5zgxi1iPUz2zcNTslJ0WieRKnUxI2/PYgGGyTtDFW5nGJ6tGyjnBB3nMa5dBFMhcB0PODsd86FyFTRy8emQrjigkwx//ATY8UcD4/+0XJ8Y/Nc3gsRtIn9dY+r0ugsi9ph7CK89aicGJkTwnFbyI7/Eu7I85jjElx7dfh8oMxoiN2Y1o9RCyW8Sh/n4oJFG06rp/RFLXBacvyf6PwUoHVMppYiFNs80N21h2kb9FVMv3llDbtuH8Sn1+cIWk2MJRX6Dns8PmPeCZIs1rdo8iiCQ88hCIk5HCY6jmEu5IdhRnNbOUtfSsxgTE/BdaCRsyxGHAVYXQ+wvEsXivuGIWn7deg8Q+DcKKU3KPq1MXosyPCnDxzAgy946KoySTJ5ZsEEgtEh2GmbglY4MiY4zLFg8Xxcfe0aNBn3Ptl0iRhl9ZQaThQSvD7GRFSvBrh5mY11vWXOwYtUb8JwU6k0unodHVwQBchUCZmxpsIFdgnbTwb4xLeGcOjFA7xOwGI8KGVfr5rNcJhSwDXXrSEsVsos5lEOichxxgri0uU3DTrYNMBnFOqyOmGUXs6mRK3CC4UnmKfPSOcXAudM6gG11aXY8DBEvnYOZ3j3ukuw9reWEWGHmKTlAiFB8cthp1vFZ9IUiojYO+RJgnLQxJW9Ke5dV6HwehEvz5Dg9B1bIaJuU08yvJQfXodmzAM0qcEH+uZB1len8MP9IR45yHPfhkdP8H3ep1PseGoX9ux6kQizBs+lJ9ADLlq8ABuvW42wnaMltw9DrKo1ccPbenBpVVqiQmhldYbqGpVYc0svbrhgyqac7bYs/Cpa+FXquAIkps3SmJgEF5jOjSAZj40EeOj5FvFeLxwK75LRTEmNz4jJEr14YjzC9id3YHjoGJFkDYuWzMeGa1fiWDvDEozjhl8vY22/9gQV53qKrj3FvZyrMLa0zY+Uo2vm7tlpBhTArpBAKKMPkgXsIrb58VMnsJOQOKnUTVn2FJvsB4wCyKFhgUfB6jKbwSNDDTz+k8fR1dOHq99zJT40v4kNy7pg/miOY81wlZoO0AwogI0K/VHASFn/nv8cw87RHkTsfJWKynEBh02jMrWHpzpu6jl7A5cltcWBureocQi3v28x+0eS/obIbG6wOLITVFfYCep4EnyFL8a8OZQriOneXUpeaWRqu+SOaW1pXm4qTK0tDocZPghS4+6ffnuGL0r4dJJPMJ1pA4X9vdkqI4rsFHVeAUYqeUIxtc884MpyzAsR22LVA+0isVBRqML6oiTK4LHO33Zpji2/2YPVNSqQiTBhotSfyxUbIHpAGb5zbHdcAXJwzWqTUfObTRKdm1BXrS/1wNsJ41f7fDabIuIiJHGAaxcmuPtdJVylmm4e5CScQmGkkNFsYlZq6CTbnVeAshqtemrLyrwyI/vmDZCToJqFrOvE9PSEgIjmHV0T2HJlBTctKzHJEQKfIu0Xci6jRs7xSmzpeOq8A9RxBZyJtJnZyxygPwYf9zyMMxEO5C189jLgM6u6sZAxnRDsBLlK3IWlC6IA/a3QBF25zfJYaTeJ2zPcta6Gld3aM5KX+PQHZXaluwtLM64AdXwxm/M4mMRV8yewdUM33jdQYpLUthfxeqZ3giyeLIsek96FphnAAdSqgI6mJQr8m2eOY5Lp/vff2Y0Bv9jmypjgbLNjoV6+eN2h7K4EeUFc8jTquALkxGpE9NccOWv3UDPC3LoPbY2avQoeza6vSMltlqnjCjDTUUpTBejeeqGpC+ZdHW8puQv3iYqXabNLHVeA/n5YCtAensSUxVUIqQFzW+5RKEDV/f+iB0hYkjxA+F4vUiS60J9Rh5Cf6QEo/OzL33kFTEV3kQj1i7KqL5S9XyGzopRhfs0qdT4E/pfRm8AJZ5OA/wI41y/4gIijKgAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Split"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.split.io/internal/api/v1/saml/acs/<customer_id>"
    audience          = "https://api.split.io/internal/api/v1/saml/acs/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_split_app_split_io,
    citrixspa_routing_domain.rd_split_customer_fqdn,
  ]
}
