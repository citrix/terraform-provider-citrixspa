# TalentLMS — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_talentlms_customer_domain_talentlms_com" {
  fqdn         = "<Customer-domain>.talentlms.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "TalentLMS"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_talentlms_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "TalentLMS"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_talentlms" {
  name         = "TalentLMS"
  type         = "saas"
  state        = "complete"
  description  = "Learning management system LMS to facilitate online seminars, courses, and other training programs."
  url          = "https://<Customer-domain>.talentlms.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAQ6SURBVGhD7ZRtbFNVGMdPjFFjDIkY/EQMRj/4gRhDgi8fNEaUN0WnMJDIREA06JQYZ4iGlwDuTaHbyjAZKtCVvbmNVV3XjTG2rrTbWAbrXm1r19GxrYV262jvbXtP76nP6b2bJH7Y1A8nJOeXpfs/zzltz2/n2UXJuxwuwBouwBouwBouwBouwBouwBouwBouwBouwBouwBouwBouwBousGAIIYeqBzJOdh2sGpBlonb/N/MIGLrHm/t8avGfsDkDF+w+49UJ9EbZ8XqH5Y9bGqMDbSiDprpjFjngxlfOEnFGrReGKrAuz3L+yg0l38nreRb0mk4t5mPRDoMsy2qRTPqmoyi9anW2OaO4696t1Wp3FnB4MKN2S2GHWoOAs1X4EEnGA2q9MKhAtzu4eKdhTY65qc8XjuK9Z69t1XY22SdhyRsQ0Hp9amdSZ/ZsLugoanBCPtHo0hidwXAcDne4ZhA6WpMLvV2Refrq6C0htT2J3qn0zuZ/ghNUNc8w/Mgug9IBYvkrcZsGXy7B5hMkNBE/vSVhN8DsxfXbJdMRZQ82F8d/2ojbipSSCgzdmFmyy5Be0GG/Htpd0r2pwAYBrSkdGJtx+yOKQJa+9/HPjK7J8PIvG/dX9p9spMdd9mm91uSEP3Nhg7OmcwxtrIQpH58SYX+nK/jy4Vb6DfPx6O5f/aGokmM5z+B2rWQ6KnyEot8+JdUfgCB+vQS3HIeAeypxV6mYieSAR/ziPtltgbeoI/REplFnHlXyV/reVUfb4H6rbd65G7h/W82KfRdeOnTpsU/q4bqgg9bqb4sYwiaNdc/PPbSzuhTPjhCIwX4lAwlHi5pSyCM2EgkouaT5T5RWHonRj6ICLfkkFhbeR0ROQAeGKtH/O136/lmpKZsI08JOFCt8UfYN0zfPCSz9+LdTF9203lCmbXRBeGBbbXPvZErgHJQPbT9fZKJ9hZAgwXGV/OZ3lqxzdgholS4q0W/1haIwP6lFFdxaFPthrZLJbb+4bzEJ31RKoG3w5tI9qVOCgKWYCEHhPfVgcFzZ202X8lfgSxqlmeiphD7uqYCs7jv4ywB6VZdrGF6WaXxhfwv8LPqg7uEddd6ggF45c6R2yH59Gjasy2mHG0gvsEkJGT3/o/Le9bntKK0CE/Lk3gb0VrnVGciuHfqmvE9ZnUMescaOPRc7tjJ+5l21dQcwsfAaz15OJ8TniKSpB4tsRrK7HUI092nx83uw9RTMklSXRW/GSUdU3Qe4/eHBMfoIq7J5x4OiGE90OOktX/NMj/jDEODQsGRz0CY8yJVVAP7vIcMqZKsjMBWJj/gjF/v/3cO3ddBPfwlB2WMlCQleU+2kPGojcfokIMKU7LlMQ2gCd+nIjPr5fwvcpXAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nAB1nABtiSTfwFQSMcJyXzAywAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "TalentLMS"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<Customer-domain>.talentlms.com/simplesaml/module.php/saml/sp/saml2-acs.php/<Customer_doamin>.talentlms.com"
    audience          = "http://<Customer-domain>.talentlms.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "unspecified"
    saml_type         = "SP"
    sp_initiated_only = true

    custom_attributes = [
      {
        name  = "Username"
        value = "ns_user_name"
      },
      {
        name  = "Email"
        value = "ns_user_email"
      },
      {
        name  = "LastName"
        value = "aaa.user.attribute(\"sn\")"
      },
      {
        name  = "FirstName"
        value = "aaa.user.attribute(\"givenName\")"
      },
    ]
  }

  depends_on = [
    citrixspa_routing_domain.rd_talentlms_customer_domain_talentlms_com,
    citrixspa_routing_domain.rd_talentlms_customer_fqdn,
  ]
}
