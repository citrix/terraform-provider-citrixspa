# SPARKPOST — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_sparkpost_app_sparkpost_com" {
  fqdn         = "app.sparkpost.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SPARKPOST"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_sparkpost_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "SPARKPOST"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_sparkpost" {
  name         = "SPARKPOST"
  type         = "saas"
  state        = "complete"
  description  = "Easily embed email into your app, website or product and effortlessly take control of manipulating templates, generating messages, sending emails, and reporting performance. Improve inbox placement and optimize customer engagement with reliable email deliverability."
  url          = "https://app.sparkpost.com/"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAUiSURBVGhD7ZVrTJtVGMf7TafTGD8YjX6YmmVLMMvGxr0FSrnftnIXBgyYyCDUgbAEEyDInUEDqAQl6ZhcAixMICwwZKCwMYZA58LEga4LjtuEQhBwp7fXf3veofDBMGPydsn7SwPP83+fnvf8z3nOqYB5xuENcA1vgGt4A1zDG+Aa3gDX8Aa4hjfANbwBruENcA1vgGt4A1zDG/hf0fc3sNGuMS8Duptt2vIYNtkdZrYDAy2PJU83JfMyoEmxIieeMyw+YPNdYEYG9MPtJOglEvCC4Q81K+0CMzJAAveS2H0kaC+b7w5zMaA5JyRRb5LotzQ5vobZKVbdBayBtbU1tXqFxv8OKtnoaTAYDBqNBgH+bm7+ubGxSXWKrr8ezaM5/S45+bruGzmJfIN9YGJqanpjY4PGCPB1DELHwcdoICws0tLS9uhRO5FI0tDQBEUqDRYKXY4ds4fo5eWvVP5o+jpTXv7Z4cPWFy58TVOQnZ1nZ+dkZeVAi5OSPqL6/v3vpadnIFhfX7e2ForFHoiXlpYSEpJbWlobG5sxVG5uoamWIcEvG5c/5BUSsEeTZk8CX9R9q4De3d0jl1cODt6ora1LSUmHolBcbGq6lJ9flJdXhACzFVRVfSUUiu/dm5qe/kUm+7i3tw910dGnQ0IiRkfH+vu/hz0bG6HxPQyDeXh5Hffw8KUpKC0tF4vd8Y7r14cwuqOja1bWp9DhqqCgGIGnp5+7O1u/uLjQ3NxKY4A3NlzuZAYbifR5TbqD/lelYXlW11JIpHvQSyjIzMyhlTsYH1eOjSlpLIBFJye3rq6rNKdERcVFRETTuK/vO8xGpVK1t3fa2jqOj9+2tRX19PTSpzCApzQG3t7HYR6Bvb2zXF6RmCiDva0GgIH6euMOb1FWU88UB2ji3mFzE7oeBfEQMGQ1UZbGStsZHr41NHSLxgJssb9/oI2NCG0QHByOrYCKSYSGnpyc/Hlk5Ifw8Gg8hYi1p6sbF5cQFPS+8dsmAxKJZ1tbB+wVF5c6ODhXV9dAd3Z2gxmJxAttQCsBDLS0XGYThkFXrOL2jNunbS1hpScQNwFzdwBBYeH5nJz8ysovtlYBbDNA/z18ONvcfEkqDaHLGRsb7+cXgLbG2fD2PoG9Vipvo39iYuKTk1MiI2Odnd0nJu6iEgZ8faVOTq44M25u3gpFrWk8Bk7gJzb2Q5HIZXHxERVXVtTx8UlVVV/iANTUKHAQIWqSj2irk2kBxbC+gt9jw2+TbG66A1JTz22Ns80AeresrJwmJSVyS0s7BKdOfRAeHkVFCpoqMDCsoKAkOzu3qOi8j49UJkuFvtVCarUarZWRkWkqN56BwkLjurq4eGKLMAPEO3aAortaQ1wFht9n2ByWMsQk4jU2ecLy8jKOPo23Gaio+BzLjMXGB9cFJgc1ICAUO0ArwMTETwcPHuro6GRzhkGfHDhwCFcqttjC4ggV6+oaLSws6T0GkZqZnZ1Dc7q6eiFeWFhQKP6+wbbA3f9YLNAWBGlrUkjYq8RHwKw+IgbmzBnZ0NBNrVa7srJ69mwaIYTW484YGLhBY2MLYXng6dq1/pkZdhnQHnfuTNAYqFQPcBjYxAQGHRkZnZ9fQO/989Ho6DheiQDi/fsqKsIDijGIXq+fm5un4g70Y92aTyQk7m1dfRYrmcBMcO1eudLF5iZWTdDYXH6J/zO8Aa7hDXANb4BreANcwxvgGt4A1/AGuIY3wDW8Aa7hDXANb4BrnnEDDPMX3uYXZ7U9DnUAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "SPARKPOST"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.sparkpost.com/api/v1/users/saml/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_sparkpost_app_sparkpost_com,
    citrixspa_routing_domain.rd_sparkpost_customer_fqdn,
  ]
}
