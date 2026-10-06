# Manuscript — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_manuscript_site_name_manuscript_com" {
  fqdn         = "<site-name>.manuscript.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Manuscript"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_manuscript_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Manuscript"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_manuscript" {
  name         = "Manuscript"
  type         = "saas"
  state        = "complete"
  description  = "A writing tool to help you plan, edit, and share your work."
  url          = "https://<site-name>.manuscript.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABBCAYAAABhNaJ7AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAqxSURBVHhe7VsLcFTVGf7OJSSQkJBAeEkYHkIqFspArdqgQEsLKFRBgVYpUEgBQbTDI1YDlEdDi+FZHAYEwVreKFhA5VEoKAMilALVCX1gS4YCzYMk5EUe5J5+/+5d3IHNbu5mNyHIN5O52XvOvfd83/n///z/2bsq/YKpJyUCV68qjBwFTJqiASj+3Y3QOLRS4fROoElLYOgiMu3Xx9RCvkF9oCAfiGioMTNF48mnDOuiuwNndgH7l2mYFeQaCVSUAi07aKiHu2ldj+TrsZPMuzaBvKtAx04aK95WuC+ubltEdjqwiVadd0UhPAYIkXklHYM8wyNNGIYwt6DJtR4/N23OC7MUBiQAv37V0eJor1vQ2P4rYNmzQHGBQkRT8hbyblSEu2c7Z6fQMKAZ/eSjXQqPfRPY9a7VcMdD4/g2jdk9Fb74GIjkZIbQwiuDerSH1qKMwxDEAiyV5OgyfMXPudlA5wdNzF1qoEMnq+EOw3/PAdtma2SlKzQScxcCwo3jd8w+If/LOXGBRtF0AedpL+AFBnvFUskrlww81w+Y/wpPi7/cIbhRBmygq/7up0B+jkKky9yrgCp2c+iA0FDGhxbAwT1A3y50i81ytjaF0Phkk0ZyLyDtGM29GWc4xGqqIny7gOsc4TIfOcq/BbkabdsBc5cD7b/hcpiaQfoXGuuT6ZoZNOXGCoqPdx+7wH28/rtAJZAbxjRRyMllAtVfIYVLTTlNMdgoLQLemgYsHKFQXChLG8n7zcKGC1SG+oywzVoDxw4pDOoBfLjFagg4NA6uB6b1Bs59BkRzhbJr7p7g0wXMCpK0HuRuUp5MS07nM4lqxjgxbxWTqQedbdXF+dMab04BijjjUYzuMi7H+FzjkE78/1YXUBy7iGTwnG0XuHEDuHxRIyzMxMULQFmJ1eAF8pAmDERl5cDkIRRhkjytelg2EUgdw4mgk0c24TOqEGpKi5nNXuLYI0zk8miSbGXwKIBz1jXSLmvsOWrg80smuvUwkZHpXHJuJgieQM4hVD2aQpw9qfDDdho737ErhMYHazRGxmt8+TkQ1Uw5MlRvdxFhyjlJBdkaXfuaSD2rMXWHgZRTzsmUFN8TPAogRdHUZLnC2VyPNrVsncLGXSaiYyjEZYokzT5mo0FDoEUbhbWpChP6A/88azV4wbkTHDhzjT+yaovltWHMSH1NegUJ5lwCWsebeG23xtBZMm7X2BW+93OgpNDzfTwKIIo1jnZ4lhsUHuhqYNsBA6mrTVqCxjUug77yIZmZaCYmhQXA9OGsLcYD1xnJb0VBnsZvxmikjFV0N5o7fd0Xc5mEohygYYRG0jYTE940ENPqdkrh0U6r9gSPAshzK7tA0Ke/gX1/BUa9qJGTpR2EvOrARqk4mzJynzsDjP4u8P5qucL5t225xkQmM//5u0IMA6j09XpDthXmiTtqPD/bRDLr+7jOHqk4UJn5Cyq/yicUxkw2cPgfwOP0uewM7QyU3maNA28YLtEXeHeVwuiHKcZDwAd/YHRvSnOny/hCCcUuuqYxINHE658ADw0UCj5MxQuqIYAThqEwY7HC1o81OjBoZTDqik96i9ZSW0REMtCGKzRoxGSmkdXfy6xLkpXLoNyDYi8+CgwYXz3iLlRbACcUWsYxUG5SWLJRs5Q2kcfq0dvyIxDS3oSSJhFT0t37Omgs+BMwcp6B+qHVJ+5CgAT4Cj0SDGw8rDBxJmNDoUZhPgOl1WYHJi/KY3wJjzIxbY2JpN8zhrQKHHEXAi6AEwpPjeBSxoDXf6jGVcaH0uuO094h7SRexGW4tFgjcZ7G/N0KnR8JjLl7QpAEEMiAFV6YqfDeaaDLtzUyGR9uMEusDLKaFFzVeGKUiVXM93s+HTziLgRRABcY5CIUZq0EVnzEVaARl05mlBWMDy5qkl1mMpHp/riJ1ceBZ18KPnEXakAAFxTaxSus2scsc4nJT06LyGS9EdfRxMrDJiYvkqWwZoi74JcARw4wGg9kTT7DOmELCo8NMLD2CJDC8nbph8Ac5gGxjgzOPvmdS4HNcxk0uVr4A9sC/OscMH4YkE0z3r+biUgrYOtaq9EWFLo8otDWsZNkn/ie1cCEeODYdhZdXB6nf8dqsAnbApw5CTQnadlqjogA2nQAVqUCzyQAh/dYnYKII+8BSb2BvRS9eXsglElUAyuRukp3sgvbAkhZKnDNmTwyJtaZb899mfXBD4Dzac62QCLtU8YOEt+6gGNguR3JdNo9wZC9CO2HJdkW4NZnyEcZh+zCxLKQKWLVl/gkMHsikJ/r6FItpNPl5j8PLEq0iMumiP1RV4oA3sophGyGtGoD/O0ElzMWOisYoPzBtSySHscAyXiTxSUytjWJ31qhBwABFcAdYaz6ZLP04PvA4K6M1kxlqwJJlNbMBF7qA1zg7EsJXT+UDfbdu0oImgAyYHGPiCjZtwfeXsz40BP4jEtoZdjMPmNpNScZ1eULmPoNeDJIxF0IngBuYMWMxtaG5m8nA1OGAP92C5R7NwJjuIzt2wRENZVs0WqoAdSIAAKZSNkHEJPOYXn7y+HALxgsx/UCNiziUsYlVSzFj0DuBh/1twfUmADuEJ9uLPuEslFJwkJcxKmWufspXGAFIIHSElZ1xfzXBxkZr+MFBV8jYEfZFJGtMK/fTfB5N8rsqxBQAfJzNR79vkb/Z1j2lpmOV20cQvgzO7xG8vucDI3IGBODXtDo3sd0fPPkaRfJ+ZhaEkAem5PNNfsNjVcXApNmKGw/qTAuycT/mJ5KQuTYma3i+GQpzL4sOYWJpLc05u5Q6DdaYfQ8haFT6Tp5VscAICACyN6fRPmEvvLJxVJh8M8MHLjACE8hKso1LcL711TlJH4tSyM6VmPOlgosPnDrbpBCrx/7X/l5QkAEcLxc5biTp9spPD3KwKbjrCJnaJRcpxDM8lxCyLUy45kXgRZxJma9ozF/B3B/N0n7bjcZiQNi7oFCQARwwOeoFJ54zsBmVpPj55jsTiEyNXI5463bayzdY2LeFgPx3d1nPPgInABVBv15mIF1R4BZazRSt9PHNyjEdazeUEQyf/afa0EAFxS+lWCg3QOBGUIdFCCw0KKAad917hoBZPLNG19nAfzEPQGsY9CQdgpYlcLs7Zp1ogq49CWwegZwmcdgI+gCTPkJ6/2twIgEYMkrTHq8ZHEZTIbmjwGShwAn9gKvJwYy5fGMoAsg7/jIDq6UvycPAYPuB95Idv4mwQXZHJkzAnixD3DxPPvGOt8fCAm9CwS4mdTxKC9NtW4PfLqf2SDJLnyZyRCJJw+n2bNmaNlWSLOrdY2dmC7XsNKwPlUdtRIE5Y2QKBZPaX8BrqSDxY/z9wl+5DFfwY5abqgVAQQy3hCS9vZjBluQPKDCvgq1JkCgIS5wLxHyww3uLgH8wD0BrOPXFkaFl1di6x7sraOyLWdEMuMqK7XO1HlUXYBycg4LVzDWrWdSEqmRleFU5G6HcCzKlu8gNQYvoADyq68/H1VYuNxESYlGbq7zFba6Bpn7Ci95gLz9XkxuytQYtgAYt1U5fmZ3Mwj+aLCBU2kK01+TbWqN4mKroa6ACmjHvtjtKCsSohoDk1id7lOI7201ELetAiPHAqdZkXXrYS+g1Dq8JEGdEjRmHAG6D7ZO3ATwfwCmgomPGAIzAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Manuscript"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<site-name>.manuscript.com/auth/SAML2/POST"
    audience          = "https://<site-name>.manuscript.com/saml-sp"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "FogBugzFullName"
        value = "aaa.user.attribute(\"givenName\")"
      },
      {
        name  = "FogBugzEmail"
        value = "ns_user_email"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_manuscript_site_name_manuscript_com,
    citrixspa_routing_domain.rd_manuscript_customer_fqdn,
  ]
}
