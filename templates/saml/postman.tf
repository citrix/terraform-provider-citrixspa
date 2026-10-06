# Postman — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_postman_customer_domain_postman_co" {
  fqdn         = "<customer-domain>.postman.co"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Postman"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_postman_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Postman"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_postman" {
  name         = "Postman"
  type         = "saas"
  state        = "complete"
  description  = "API Development Environment"
  url          = "https://<customer-domain>.postman.co"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAsqSURBVHhe1VsJeFTVGT0MqyyBkBAskUVIC6ZYxAVEbFHBUhcSoFWoCkiVspQGQliUlqVfsYLginxiEUGRoqQVUUpbtAqlSKEKFm3UIkXCJqgJkBBIgKTnvPvmy8xkZpJZ3sz0fF8y895MZu7577+c/96belUEnERlJbB3F1D4b+DQZ0DRIaDkFHD6JHD2NN/Ar2/UFGjeCkjiT2p7oF1noNPlQOceQP365nMcgjMGOH4Q2LIW2LkROPY5v8UFNGxMMg0AFwnVq8dH3gMfLXAIlfyporEqzwMXLgDnzhrjXXwp0OsWoN+dQFoH+/3RQ3QN8PYaYP1i4EsaoClnsxFJuwmHAw2tksaoKAfKTgAp7YDBOUD/u+03RI7oGOA1kn79aQ6YRC9qbmbaCVygd5wpNc+zxgFDJpvnESAyA2xbDyybSsINgSYkbrl1DKDQUP44zzD56WNA32z7hdARngE0E3OHAAcKgBYpsSPuCxmi5GsmzG8Dc9aFlTBDH/meLcDILoxzZvOWbeJHXtB3awzHmHNGsnLsftt+oe4IzQPyFwHrngRaf4N/GUfi/iAaRUeA7InAsBn2zdpRdwM8MQbYRQsn0eJhJnXHISanvgKu+B4w5XlzrxbUzQALRwIFOyhWku0bCY5SlszLegHTX7RvBEbtfvz4/YlHXnPm/vEHqcqPOWaNvRYEN8ArjzCxbE4c8qo+JUWcYWb+08V0dz6WUVb7M4TGvPsdYC3zVhAEDoFdbwGLfsKER/WVCDFfcQZozJ5hHOt+Zh/7HhXiu68BL883WsS3IomZEuP0lcwLN5l7PvBvgPPnTKmzsn0CsNd4mrdkLvJT5greBZZQEaqPkAF8q5PoFR0FXtwHNKCRfOA/BOYOZbanwEkE8iIgt5fQ8cX7m4DZVIFLdgJ3UpGW2TLZE+IgLhJuflDTANvoUoVUeA3YyCQCFPeXsjWWB3jiPZJ/mE3R7780DdMNw9lBMkz8RbS4FH5suPmgpgGk7SVvE6XWqz1u382+sLHrrwyHUYa8sJiN0Ul6SeolJhR8IS7iJG4+8DbAq48zmTQysZQoUEwXfWFf2Jg6gCXu7/YF8dVhSuLWrAo0QiCFKk7iJo4e8H73G0tNVxdvyI0r2OlJ0JTzsWCb/YKNlXTn8T0567YH/Hoj8Ol7TJasCsHylriJoweqDfDWKv7iH8d79pXxT3DGM68Dcp+hpH3WrAo9Od5+A9GBIfHUP4B7OhktICyeADSrRa9Y3MjR4mpQXQYn8QvPVdBNHFrMqA0ahmY8lbpj5ssswW3N/TUPAds3AF/TzXsPAnKWmPtCMT1gaj/miK7A0f1Aoyb2C0GgpCqOi2lAwhjgeCEwuS+Qkm7djDmUuBTnQ3KAOzwS1aLRnDUOdsoyc537Xba+B4AeN5jSpudfaM2Rs+qnxgeEjPkEwyqtgx0Cm1+h+1A/xwOakRPHOeu/8yYvHV+vfjV5ofdtZpz7PwI+oMyVAUQ8FPKCPuOdNdZTY4AddDGt2sYa55i0yileln4AdL/evkk8wu5T1SzvOXMtyEv/zOvGdPOGzOZydxEPR6yJ684/WU9d1hL08UN85uz6ew2US9tfBCz/xHRvbsjtpes9yRcwXicw6/f8Pj2G440U4qqle36WC/tofUtDh2HJcKGV3eQ2zOzU8Z6Yfw9/cRx5y8218OYLwIMDKXqY7Q+y/EVjJUpcxXnfv2iAz/fE1v3LSoBLMoAF7DY9sfBek52neqzkrPgla/4sYP1JE/Pab4jWTpE4k7sLh9klxar0aTuscyYbEx9N/pu7TBmcxrbVjXl3AHs2A6sZ+8JT1AHNqfaiBXE+/BkNUMSSEIv4t5apegO/yLdv2FgwgsmMg5lOV3dj4jUcIBOdW+7OZ9OjhiaaIk2ci47QACeon51Wf0p41/zAe4aFh9jBaebda3cqiXexoemTxbhfbe5pFVpJUAkzmlAuOVVMA5QURyexBIK1r0fXb5pk37AxnzMvkfPAS+b6JJNcNrX6fQuBuxn3wodbgfxHKXpSzXU04WIipFe6rAbCKVgJrzPwEtXaR1uAp6n0hHnDTJv7oE3+kx3AsIuNy/f/sbl36D98Hz3EsVUpfmb5aUrhnOuqLNeL9pfI7btcDsywSQqTKLdVxzOZCybQtYWtrzIPMMZfoJZPoREE7TpNoTBKSnMuPO0WyIWGaiDMRdQgkmnsK7I54ytnM8b7A8PZ5GjToqKsumtbNRdY8nNgA43lJl98jH0JGzNtwDidm5hY61XNGcpmiJo62qXwAtvasyRraXVmdMneAZS4w6YBvxpKotT/TVuwBBpJaqGYDdFkzrwM5PDJEGsZjc2fCy0oQ213iCokZ5u1NEvZ+rILFPf7qTpHs5c/uBcoYSubyZBwQyLnZ1eb9XynyQvaWU5KpgesmFWF7a/XrZcOB0qE2r6+PhtoSwXYkQaQYYQZ1PYZ1Pg/ymWf3yc2bu+GVpyuvZ0ekM4srSToBEpZ/r7ZA5i1FriRiU7Jz01eWLAJKGQzNJHkY73VLs7pGTRAp++YA0nRhmRvt6tY59nnB8JG9vpHKcWTme2d1CL+oJxE7i504QwpB0QzD0j2dmM8e5ZAX2jbbc3DQJNmDtX5ILD4Mgdk9KABpInT2ptEFQ1Ymp9afnr1wqMXdJpj7BXAp/8EWmnmY0xeEFcduSN343e9bjMuESlEPlP78gHIaxNzMvOAdaiKMx8viGuvW62nxgDaVjrNwUcCke/OZDbNz6GEo/81q85/YTOUQsvHoswFg7jexBacMAZQCLRhFxZuNXCTz1th3/DAsulmNVdL7lr6ioPHe0Ec0zoCqWYFvDr1DppQfQgxFJSym+zO2fUlr82HUaz7WnxsQwPHatGlNohjFrna8D4fMLor5SnrdF3rsQ4rqrbneixgbn+DjQ31f7k0Pz8r1uUtGKT+dKJkBbWHDe/RyTJn6+AFMpkUnnpqN3ltbIzpDiydYmZbkjaRyAviNmicfWHgPcIhkxgjbGJkqWA4RR1/6/1Ud0yeMylnBXVxZ+gR2rGJxRJbqBAncRvqfb645hSNfdQcPw2kizTzN48Cfkj9Pnwme/4rgTm385GCqgGJO9FYRQoNSZzEzQfeOcCNWVlm40DH3T0hAaEcMXaRsaZObaR/C3g2DzjPLF/ILq80BmuMoUJ1X4cn5jE/+cC/AUR0BJuk5LY141ivKZHIpSQnpeRkFN1X759o7q8xKjxXUYv4GZv/qdIbtUytxOZrHr2mBKdYVwenoydKemqnE448f3RCTFwCjC2wr2oLWgePtYwVDPKAeOj5ukDJejATu7gEQPBgHUYV17Mf4zpCmRwPSKD1vNEcnwuC2rNV7nKgK/v6/ycjaK8j81pvgRYAtRtAeGC1kbsnGQ41U2biQGPTwakrOfPTPLbagqBuBhDyngcG3suMeoRflIBW0Ji0z5nFvJXjfRIsGPyXwWB4/03gsftM9o/HqRJ/UJ2X0JnCcL3qZvtm3RC6AYTzbCnnZpsFTRkiXsJHWkTEO14GzF7HCaEOCRHhGcCNrX8Anpth7bBYKzyxMoSIq7GRGh2zEOg72H4hdERmADfyOYg//pYGoNjQaUynen9tuZ1hLyJ1N2i86UciRHQM4MYmZt4NTEA6u6u1AOUIGSVcoaShSWIrxiW/W1F5qmMdMMJ+Q+SIrgHc0OHFv+Wbf57W8Tb1EzKGPEMGUah49RgcgkWWMyvCWrYSaT3X6u3VA80/T7frYr8/enDGAJ4Qmb27gQMfAkf2m1JVTD2hjROdTdDX69/nmyWxt2CP0ZpdW3qG+R8BtdiOLqUB/wNXftT8k7hfWgAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Postman"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://identity.getpostman.com/sso/saml/<customer_id>/callback"
    audience          = "https://identity.getpostman.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_postman_customer_domain_postman_co,
    citrixspa_routing_domain.rd_postman_customer_fqdn,
  ]
}
