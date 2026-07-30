# =============================================================================
# Customer A — GOLDEN standard configuration
# =============================================================================
# This is the reference ("golden") configuration for the dedicated Customer A
# E2E tenant. Customer A is expected to ALWAYS match this config, so the e2e
# `golden` scenario applies it and asserts there is no drift, and the `migrate`
# scenario copies Customer A's live state into Customer B.
#
# NOTE: this is an intentionally SMALL starter set (a couple of apps + a policy)
# so the folder/structure exists. Additional cases (web/saas/ztna variants,
# security groups, session policies, etc.) will be added later.
#
# Names use a STABLE prefix ("${name_prefix}-golden-...") — no per-run tag —
# because Customer A is a persistent, pre-provisioned tenant.
# =============================================================================

locals {
  prefix = "${var.name_prefix}-golden"

  # Minimal valid 8x8 PNG icon (base64).
  icon = "iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsIAAA7CARUoSoAAAAAaSURBVChTY6AzeCuj8h+EoVwwYILSAwcYGACG/ARbHXQf2wAAAABJRU5ErkJggg=="

  web_fqdn  = "${local.prefix}-web.example.com"
  saas_fqdn = "${local.prefix}-saas.example.com"
}

# --- Routing domain for the web application ----------------------------------
resource "citrixspa_routing_domain" "web" {
  fqdn         = local.web_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web routing domain"
}

# --- Web application (complete) ----------------------------------------------
resource "citrixspa_application" "web" {
  name           = "${local.prefix}-web"
  type           = "web"
  state          = "complete"
  description    = "Golden web application"
  url            = "https://${local.web_fqdn}"
  using_template = false
  related_urls   = [local.web_fqdn]
  icon           = local.icon

  depends_on = [citrixspa_routing_domain.web]
}

# --- SaaS application ---------------------------------------------------------
resource "citrixspa_routing_domain" "saas" {
  fqdn         = local.saas_fqdn
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden SaaS routing domain"
}

resource "citrixspa_application" "saas" {
  name           = "${local.prefix}-saas"
  type           = "saas"
  state          = "complete"
  description    = "Golden SaaS application"
  url            = "https://${local.saas_fqdn}"
  using_template = false
  related_urls   = [local.saas_fqdn]
  icon           = local.icon

  depends_on = [citrixspa_routing_domain.saas]
}

# --- Access policy for the web application -----------------------------------
resource "citrixspa_access_policy" "web" {
  name        = "${local.prefix}-web-policy"
  description = ""
  active      = true
  priority    = 1

  apps = [citrixspa_application.web.id]

  access_rules = [
    {
      name        = "allow-all"
      description = ""
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["EMAIL:/e2e/golden-user@example.com"]
        }
      ]
    }
  ]
}

# =============================================================================
# Outputs
# =============================================================================
output "web_application_id" {
  value = citrixspa_application.web.id
}

output "saas_application_id" {
  value = citrixspa_application.saas.id
}

output "web_access_policy_id" {
  value = citrixspa_access_policy.web.id
}

output "web_routing_domain_fqdn" {
  value = citrixspa_routing_domain.web.fqdn
}
