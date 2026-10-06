# Dropbox — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_dropbox_www_dropbox_com" {
  fqdn         = "www.dropbox.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Dropbox"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_dropbox_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Dropbox"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_dropbox" {
  name         = "Dropbox"
  type         = "saas"
  state        = "complete"
  description  = "Cloud storage tool for secure file sharing and storage."
  url          = "https://www.dropbox.com/sso/<your-org-id>"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAAAMCAYAAADS87vJAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAABmJLR0QAAAAAAAD5Q7t/AAAACXBIWXMAAAsSAAALEgHS3X78AAADu0lEQVRIx93WW4hWVRQH8N/5ZhwzhCQtky5adtHMVCbDCtNKlArBirJIne+rJKHALmQTZjmpdKEke+geM14wsgsFaZYSlhdCS0pDoqv5kEJjkJYy1ny7h7NGv9HRXrSH1ss6e/33WWf/1/7vtU+mzerSKAUNkjWYoynbtR+bkHqqNgd9ZOo1Zp/7n1imlM5XVq9gYkX8G61mY4UqN0rqZU4LrEXmJWVzNWU//8frrcZYNGP10UhYwNiDyEM/VaYqGIp7K8hDZ8ndqlxeEbsQy/Ax1uBd1B2DAnTBG5h5tBIWNGZPSobhI7Rih7JpGKEpW6qsVjILO/EX3pIM9Fq2sCLPSbgaPbEd56IJTwU+Gn1xCsbjxIhfhQcwGb0jdgLGoRtG4X6MCawVe7ALF+A+3BKFabOTUYy810asF25C/xgfh+swon05SmmBYhqjIyumSerSvMMUciQSpsS4GmvxN87Cd/gU22LeJXgunndW+GE4L8af4Be0xHgGqiK2Dd/jt8CWy9U8AFsj1hz+7SjKT9iBHngmsJvbiA9RTEuVUotS2qouTTExVQfxrurSdKX0q1Lao5SaFNOZhynA1IrYwxEbjXXx3BC7en2M5+F4DMKfQXpAYBtDMb2wGXtxdpBvDgV0xYKYfxleD5VcGqp4LLAbcA5+j1yteAQKJqQHZTbKXIMa9FbwgmorTUwjZDYomB2V6yI/21vcnm51ZEsVz51jBx7FSgfkvlAu6a/wWZDqEVhj7Nh2vBmyPT3UtQpf4w8sjvlDMDAIrouCLQrsylDhE/GNzVEcBXyobMlBi9+kbJkqg5U14scKrIxFWm3ogHRL+IL83JZDqjWxoDbbE/6MiljPUMG+GPetwPqE3xs5e3eANUfe7vF9ODX8jlDLpJgzIFSh2qLsS4xXTItlpktW41uZyTK1WClpwHCd9NNqhsZs1UHEq8LfhSvkzWYgno3idZPLrs1WxO69jItj/gA8HkTgjsiboYT12CJXYa38NtiGO7Fb3sT7RI7l+KKC8Pt4R95fhstvkSUYmrWjcVs6UfK8vFO3t7K5mj1kabbPoTYIT8fiOsmb03tBMFdM3hCLFe+MxHT5Fbo7FjQT/eRH4sUgNDjI3IMfgsjaUMi4KMJMfBBKq5ffDN2jYA3y4zgL8/Fq5H0F69sXoJgKWo3VyQxJ7f54ZoVWs8zPjsrPx0HWWS77tp5xETbIr7+5gbcc5t0aB45MR3nb3su070n/YnWpRjFNU0wblNKkY0D6SNYfmxybH6lD7B/10iUqMAVnRgAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Dropbox"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.dropbox.com/saml_login"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_dropbox_www_dropbox_com,
    citrixspa_routing_domain.rd_dropbox_customer_fqdn,
  ]
}
