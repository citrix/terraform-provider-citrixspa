# Service Now — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_service_now_your_organization_service_now_com" {
  fqdn         = "<your-organization>.service-now.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_servicenow_com" {
  fqdn         = "*.servicenow.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_servicenowservices_com" {
  fqdn         = "*.servicenowservices.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_akamaiedge_net" {
  fqdn         = "*.akamaiedge.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_akamaihd_net" {
  fqdn         = "*.akamaihd.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_service_now_cloudflare_net" {
  fqdn         = "*.cloudflare.net"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Service Now"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_service_now" {
  name         = "Service Now"
  type         = "saas"
  state        = "complete"
  description  = "Service Management"
  url          = "https://<your-organization>.service-now.com/"
  related_urls = ["<Customer FQDN>", "*.servicenow.com", "*.servicenowservices.com", "*.akamaiedge.net", "*.akamaihd.net", "*.cloudflare.net"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAATMAAABACAMAAABBR2C3AAAAYFBMVEVfYGJfYGLjGyNfYGJfYGJfYGLjGyNfYGJfYGJfYGJfYGJfYGJfYGLjGyNfYGLjGyPjGyNfYGLjGyNfYGJfYGLjGyPjGyPjGyPjGyPjGyPjGyPjGyPjGyPjGyNfYGLjGyMsGVy+AAAAAXRSTlMAQObYZgAAAAFiS0dEAIgFHUgAAAAJcEhZcwAACxMAAAsTAQCanBgAAAAHdElNRQfiBQcGOBdSruJuAAAD7ElEQVR42u2aa4sqMQyG0x1kwAVdWESRpf3///KgjtqkfZNU51zg5P3izKSt08deklSiUCgUCoVCoVAoFAqFQqFQKDSifJdm3fKnE7NerrY5z7360828yXnHDUn92kt7Xev+eVmuOsCOXaydZ0eDR6cW12dmQjgvmvhzTvv68dm2v6mKwpYbMAlZy4NDqYSRCdtRK4+rKVBktxujYFZbZjBkaqRKy1l/q7ZDhcvZ+WIzW4rs/cgkGGDrGK4f3/0ZtlhTf2xLaLNivXXodCzFhoaZFZuZjSylyeZSGRGzDIYZY5az1vRtNiNrgVqJ2cEokdmv/+gMu5k043J7+/Ayu4+yJH+c3t11D6nuHczEinfdN+QqKCme3VTBQtIiqtflbsf4Ftp5VlGCVXegmRoxw3Tcf7TQnjePq7N8VjAi19RkTy5LyYwIZI2owowYs1uxGbDVyTNmcray21Nvv+gwKuCBiizxZ99sMJDes8n6CbaI2Sf8+XpEqcMMrFH1zZ9ixhYR0nuGkbInO/ZdwI3dMI+PYMu95abldB1mzSytoS1VvySzL9cWMGLiy1M257qYzAntFLfnhM2SGRnM5L6JmDVT3dxXMTTt1TOaQU2t3DJLSrT08DMGmR0AEY3Zz3vM1HfUXAezmvQgEvJ8eY3tIDOymZ2WjZbopy5yHme2dccu0r18j9lAKCfDuheZgSK143LwBZvS5V6HGQ8kpvWY5d/ETHHfXC/5HjNqmdFfYPYxxOxMg8z60FZhtvs3mS2XX1UBaTsN5Byv2niWFdT1WQ5WWpEZrTI3m9Ldi6FcreJ+OfZVWoOZI7vlZ3bAzE4NqvMQM+56TCswy5jZ/GeY0TIFEbM2yirDzKTP+gozEV41cdqst2wO8CFmoFrL7H75ArLnMFBC0feYvRg7eZntAbSe33pYk1lWfu/FScddm5UDACVGT8ZKKnKOkEjRChQZmEuMNrMZRNmTHqQbw8FgRnrGRAnmhpgdtYx2C+jsZYbSOc8IYeObeG2Z3U5MbplzpJGcoz/eVBNsCjNyIktaXpVyJzeYPcyoG8M2ue3UVtmgDDBlP7OfR98/QIqsC8jJjEDPsm5diVnntJMcZpMZPmU5Nsmj15ndp6B42eoQcled5GVyM0M54epkvmmYmbcpyajOZkaOs7yijb2xWJM7nN/KOaPqEljM1JNTGECRlxnZx5+eRc+ZC9L/U9GE3C8zs/57sLXO0XVm9GGdGL/DDOWC0L8AnDFOp6mpvzaiTLHxf41iObrWGXvBVQbDc30oGm5K97hJq/FctkZeC7v5AEs/s9Py2b8UBIRCoVAoFAqFQqFQKBQKhUL/u34BP/bn+lVvlEkAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Service Now"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.service-now.com/navpage.do"
    audience          = "https://<your-organization>.service-now.com"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_service_now_your_organization_service_now_com,
    citrixspa_routing_domain.rd_service_now_customer_fqdn,
    citrixspa_routing_domain.rd_service_now_servicenow_com,
    citrixspa_routing_domain.rd_service_now_servicenowservices_com,
    citrixspa_routing_domain.rd_service_now_akamaiedge_net,
    citrixspa_routing_domain.rd_service_now_akamaihd_net,
    citrixspa_routing_domain.rd_service_now_cloudflare_net,
  ]
}
