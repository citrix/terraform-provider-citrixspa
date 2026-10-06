# StatusDashboard — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_statusdashboard_www_statusdashboard_com" {
  fqdn         = "www.statusdashboard.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "StatusDashboard"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_statusdashboard_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "StatusDashboard"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_statusdashboard" {
  name         = "StatusDashboard"
  type         = "saas"
  state        = "complete"
  description  = "Public and private, securely hosted status pages"
  url          = "https://www.statusdashboard.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAAZdEVYdFNvZnR3YXJlAEFkb2JlIEltYWdlUmVhZHlxyWU8AAADdmlUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgNS42LWMxNDIgNzkuMTYwOTI0LCAyMDE3LzA3LzEzLTAxOjA2OjM5ICAgICAgICAiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJkZjphYm91dD0iIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtbG5zOnhtcD0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wLyIgeG1wTU06T3JpZ2luYWxEb2N1bWVudElEPSJ4bXAuZGlkOjYzNmU3M2ZjLTAxMGMtYTk0OC04ZmM4LTNkZjA1N2E1MmM3ZSIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDpCOEMyNUUzNDA0QUIxMUU5OUU3NkYwRDFDNjYwM0RGMCIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpCOEMyNUUzMzA0QUIxMUU5OUU3NkYwRDFDNjYwM0RGMCIgeG1wOkNyZWF0b3JUb29sPSJBZG9iZSBQaG90b3Nob3AgQ0MgMjAxOCAoV2luZG93cykiPiA8eG1wTU06RGVyaXZlZEZyb20gc3RSZWY6aW5zdGFuY2VJRD0ieG1wLmlpZDo3ZTM2MDFhZi1jYjRjLThjNGMtYmFmNC04MTRkYzUwYzY2MTEiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6NjM2ZTczZmMtMDEwYy1hOTQ4LThmYzgtM2RmMDU3YTUyYzdlIi8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+D513EwAABDNJREFUeF7tmlts01Ucx7/ter/vQqfd3BibuCDTOV2GREOUIMTFZCEmOGNUwnRIMmOiL5IQE42iPmk0BkR9QeJtiTcgBMZkXHavlg43Wcdcdytd2dZu9ErX+evxPJjQLr7COZ+X79n/f875n/M5v/770CmWCQiMkqewSAE8hUUK4CksUgBPYZECeAqLFMBTWKQAnsIiBfAUFimAp7BIATyFRQrgKSxSAE9hkQJ4CosUwHNFbudfz1b8bfCg5xfs7twPxObTFpCXX4k/nvwcJcZC3uPWJ6uAL0eOo6ntNcB4B9VJDhOAeAhILMK/sx92nY33vLXJ+hFoOruPNm+nFm08vfnIDN58cA/q1jyBXd0fsj6z8QQazjmxs8fN/u6dDbL8L675Bd66GXdwES90u7Gn/xK/kpl3/vTwFvDpsJe3/j+febKPySggmLgO3AjDorHCka6A0Bhatx3AjtLH0TNxHhdm/l3wblr4Q3lWPFvqwAlfAHU/tqFvLoS97sto6h2Ak9oP/NyOo9MzrH3q6jXMxOJ4kTb93uAVJkebo0S9w46KXzvgDUewub0Hw4th/DTpx9YzfYgkk3hrYASNnS6EEkl8NDyGp8//zsZORqLYcKoTXdfmMUHtx2jshcA8G7ur142FG0l2QC3OQbbeTGQUYFEbmYCWyu2Y2v4DnI0nsaFgHapbGwC1CSa1nvU7VFuFoYXreIVEVJqNMNksqCUhydQyzgXmEF1KoaowHyqFAm3+WRwnEUe80+ihSlGwGYB7aFxagC8Wgy8ax8MFuXiJ5O0b8OA+qxlTdM2sykFz+V14e9ADjVKB1kdqmMT6Die6t2zEo6e7EaNnbVqVxzb7wdAoXl27muR04buN1bBp1PxpN5NRgJIWbChYj3c730eH/yJq8tai+JstgC4XSEbwDFVCmjdcf8Gu1aLMaEAvnXB0aYmd8lejk2wOq1qF0XAUcVrcaRJwxOuj/hpsvXMV9rou04ZjODw2hW100q9XlqH+rBN6qoggnXRDsR3jdP/glQnatBKlRh3Ns4zpSByH6FqhXosigxbfktASgx7NfZdgIFGJ5RTiqRTup8MopvsHRsYxF6KKzkLWl+BEeAYlX28CVDpoDAVIxBeBpTgsJgdCO47xXsDhv6fIsApPFRViiB6UpOm8tOkivQ732kzoJzHlJgN+IwHlJiNq8634ftyHNXRtncWEk1cDqKDr621mTNOGL9J7ocyop1NX4RhVzMsVJazEa3KtVA0xtrm+2RCeW+1gkj+hj0Qz9UlLbqdn3G02IEEVWJ1rYev7gmRV0dx1+Zlf2it+DQZiQTzftR9n/C5akB6NpZvxcW0Lv3t7IP9HiKewSAE8hUUK4CksUgBPYZECeAqLFMBTWKQAnsIiBfAUFimAp7BIATyFRQrgKSxSAE9hkQJ4CosUwFNQgH8AFjekVqZpv/kAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "StatusDashboard"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.statusdashboard.com/accounts/login/sso/acs/<Customer-id>/"
    audience          = "www.statusdashboard.com/<Customer-id>/"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_statusdashboard_www_statusdashboard_com,
    citrixspa_routing_domain.rd_statusdashboard_customer_fqdn,
  ]
}
