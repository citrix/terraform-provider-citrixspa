# LiveChat — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_livechat_my_livechatinc_com" {
  fqdn         = "my.livechatinc.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "LiveChat"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_livechat_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "LiveChat"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_livechat" {
  name         = "LiveChat"
  type         = "saas"
  state        = "complete"
  description  = "Live chat and help desk software for businesses."
  url          = "https://my.livechatinc.com/chats"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAr+SURBVHhe5VtpbBTJFf7m8G184oPT2ECWiOsHoIQoQARIkCiBKFyCLAixEEAIQVAIQhwSAsSh3SSKwEFaIBwJbGAFEVGyQSEyV+BHQBBsrsUS8RrhC2N78DHjOZz6arqH7pnu8T1jZz9UzFRNd/V7X7336lV12eJ0OtvwNYZV+fzaolctoK2t+11bLBblW++g0wRYrVbY7fZeF6wz8Pl88Hg8knCWzsjWYQJiY2Nlx3zAvXv38OTJE5SVleHVq1dwOBzwer1SEJbOCtEe1P5IvjoAqampGD58OIYNG4YJEyZg4sSJ8lo+3+12y+8dQbsE8IExMTF49OgRDh06hFu3bklh2Gaz2WRRle1Jpc1AMgiVbBLf2toq5Zw1axZ27tyJ/Px82aZeGw5hCYiLi5Ofy5cvx7Vr15Ceni7bqGhPj3JXocpBMlwuF+rq6rBgwQIcPXpUtpGIcHKaEkBFa2trMW3aNMlyYmKiaUcdYbqnEU6WxsZGpKWlSWulHiTBDCEEsAOaN4PKuHHjZAesax+osk5iWNS2SEGVRRuMg+WjNTBuPXz4UMYEWoMRdATwRvoSb6Q/VVRUID4+XvnVD7VzBj6yXFBQEHCVSIHKNjc3o7S0VH4OGDAgEKS1aGpqksHxwoULUmYjhFgAlTl37hy2b9+OjIwMXadk8c2bN5g5cyZ27dqF0aNHK79ED8XFxdi9ezfu37+PzMzMEEuorq7G2bNnpcxGJBgSQNNXIzzBjlhqamoCnfU1XLx4EZs2bUJOTo6OBJr/wIEDUVRUZEiALhWmGZHJhoYGnfIEmbxy5UqfVJ5YtGgRTp48KeVUZSYYJ54/f47y8nL5PRg6Asjc9evXkZCQoLT42+hL69atw6RJk5TWvom5c+di/vz5aGlpUVr88lOfGzduBAZVi5DFEDM8LVNkk4GGftYfcODAATlgWivgLEa9jBBCwNOnT3UE0G9mz56t1Po+mCIzhnEaV0F9Hj9+rNT0CCGACgcHkcmTJyu1/gHKq10PUJ93794pNT0CBKhK09y1BNCUOM/2J1BebeLD3Ka+vl6p6RFiAUa5s9afVDAPoGswVZ4xY4YMQEw4egKu6nJUfP4JSnfPx9OfTUDxhyPxZFkeipeNQMmqb+L55u/g5ScfoeYfp+Fudih3mYP6cGCNEEKA1nfC4fbt23IpzCnz7du3cqoxCzTtQaXXVVWGF0LpL7d8F2//fhKtNV+JCBYHe5pIcDJzYM/Mhi0xGb5WF1q+vIfq8wdQ8tM85W5zkACz9YCOAJqN0WgbgTkDoysDDD/VemfBp9HeKi//Vig+De7qMqFwDmwJybDamd5a5e8WcaEs/Ge1wRorVqVxCRgwfhq7aRfqmiUYOgKovNmiobdA5b769JeoufRr2NNzYRUjrrog5WnzeeB1O+FxNsPrbIK7tQU+j1u0i8FqdSJt6nx5bXtol4Bgv48UKv96DA03PkdMWpZGcVGEkl5HrTD5FKRO/j5yfrwB2T/5ObJnf4iEUSIhExd56t8gbfZyeU9XERIDIonW+mrU/GG/MPmBoqZRvrUJFuECBfv/hm98XIShaw4h64frkfWD1chZvA35vziOMYX3MbrwHmwxsfK+riKqBJQXboIthSs4f12avMeJuOx8jPnVTSQO+8D/gwF4S2JOnkJb1xE1AlrrKtH8+A4scgSpBodexCCXE/n7/iKviQSiRsDbf/4RVmHmjOoSQn+vqwnZC7dEVKioEeD49xcy4mv0h0XM7xnfW+ZviBCiQgAnWldlGWDTrM/FtGYT06A9KbJpd1QI8DiqRVKjzzc438cNGaXUIofoWEBtpXiyP8MLQFhAbHqOUokcokKA2+kU/4tHaxgQeR0ssYlKLXKICgF2sW5Am1cGPkLM/lKQNrf5C4zeQlQIsIq01+pT1ffDYrGJ1LdGqUUOUSEgJn0QfFzaKRzIXMBmQ8vrUn9DBBEdCxAB0J6WLQNfwA7EEtdT/Qoen/GqrbcQFQKIpLHTRd7veh8HuSCw2+C49WelITKIGgEZM5eI9X2jxg2EMCI1rji/771VRADRs4DRkxCTMVgmQAGNhRtYvG6Ufbxaaeh99CgB2jdKHcHgjw7A0/BG6M9/fiuwxCeh6cm/ULrzR3DWVcnrzOARZHm93Zs6Ay9HGZhY8vLykJX1fneGr8F57GTlypWyrmLOnDl4/fp1YB+QW2m8n6+fzPYVeQ336/lGV8V/f7MezcVFsCamKiSQDPFskRN4m+qRMGws4sZ+C/G5BbDGxcPT0ghvTTlaSv8Dx6MbGLz6ILLnrvJ3puDw4cM4fvw4kpOTZZ3y8MUuN3GDX5AGLIAXdWdbjMpTQe6+Ukmjwh1nEqrFiM2/gy07H75mh9z0pPIkgfsEttQsuOpew3H9T6g6tw8VJ3eg5rNDqCv6DM6KUsRlDUH9tbNKT+FhppvOBXgRFekqVAsIV1SitPjgwBeIH/NtuOuFyXs12/KUxxYj3cKakAJrUqp/t5h10Q5RXJUv4XYZ7/lrwWcbIYQAo1fIRuBIC/fpdOF9RiQXbP09hm7+VG55expq4aNSggwZHaRL+T/lN7GS9PE3cU2bswnvHt2UfZiB1s1teyMEYgCV50U89ZGUJBhWhDSLAVu3bpWHqMyYNQIFoRucOnVKaTGGo/gmHHevoLHkLrwiSIqEQYyUBT66qSCozRqDhNyhiBFWkzx5LlLHTtWNZHAMUF2TL26CY4DuhAhPh/DQIclQFTMjIJJoFfmCxS2SJm6gxKfAbg0fq3ie8cSJEwECSDq/37lzxzwIqgg+8ERL4KmLaCI2PhkxAzJhFzNFe8oTVVVVOsuk5aWkpCg1PUII4BSofYtCt+Cpkf4EjrT2NR31yc3NVWp6BAggS8T48eN1L0jZ0YMHD+QL0P4AHoTgfK+1AOpD1zZCgAB1npwyZYrOT9jOI7Jr1qxRWvo2NmzYIM8vqvoQnH2olxF0LkCm+J6fh4xUiyAYF3h67ODBg0pL38SWLVvkaTDtlMcZgNbAcwxMxoKhI4C+wmjJNJesqVCtoLCwUJ7F64tYsWIFLl++HBLseDBi8eLF8ntwAkbopkGOOn2ex0noM9nZ2TpT4u88gUXw2Ny8efMwalTkt7IJyvLs2TNcunRJzvm0Ui7GtPJSYcauly9fyu/tEkCwY54PPnPmDLZt24ZBgwYpv7wHOyKztBJOkySN+QKPqK1apV+YmIHnkJkQnT9/XpJqlB2GA82ZclBpVfHgweJijXrQooPnfxUhBKggozwHxNOXRpag1vmdhQTs2bNH/m2BGSjw6dOnpVAcPR5mCh61jiJYYS34nMrKSuzYsQMbN240VZ4wpZ037d27V2aB7Ix1KkpoH8zvHD0WM4GuXr2KpUuXYsiQIdi/f788cM15mfGGAUq9vzPF6FmUj1bJpe+RI0faVZ4Ia3e8ef369bh79y5GjBghM0L+MQLzarKsjj7BT+3R+pKSEhmVR44cibVr18o/uaHSDFJ0GTOy2oP2eSyUg/LwHCAzQOYxzFsWLlzYrvKEqQtoQYHJ+osXL+SpbBKiLizYTnAK5dn86dOn49ixY5IsmjhdSb2mOyBhVJifKvn85F+yMGBPnToVS5YskX9ExdlMm8yFQ4cIUKESoYICMOsi+3wgzY+Fft0TSgeDfdJluGTnkdihQ4cqv/hBeYzm+nDoFAHBoEBq0aKr5h0OqukTqvlzpLXtXUG3CAh+uKp4d4Uyglnf3SW7WwT8P6DnHbWf4WtOAPA/nOmwJxsDXooAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "LiveChat"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://api.livechatinc.com/v2/authorize/saml/callback"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_livechat_my_livechatinc_com,
    citrixspa_routing_domain.rd_livechat_customer_fqdn,
  ]
}
