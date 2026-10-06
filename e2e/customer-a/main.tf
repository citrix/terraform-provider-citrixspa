# =============================================================================
# Customer A — GOLDEN standard configuration
# =============================================================================
# This is the reference ("golden") configuration for the dedicated Customer A
# E2E tenant. Customer A is expected to ALWAYS match this config, so the e2e
# `golden` scenario applies it and asserts there is no drift, and the `migrate`
# scenario copies Customer A's live state into Customer B.
#
# NOTE: this is a COMPREHENSIVE reference config exercising the full breadth of
# supported resources — web/saas/ztna application variants, routing domains
# (internal/external/conflicting/bypass, hostname and IP/CIDR/range), access
# policies, a security group, and a session policy.
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
  ztna_fqdn = "${local.prefix}-ztna.internal.example.com"

  web_flags_fqdn      = "${local.prefix}-web-flags.example.com"
  web_flags_rel_fqdn  = "api.${local.prefix}-web-flags.example.com"
  web_incomplete_fqdn = "${local.prefix}-web-incomplete.example.com"
  saas_sso_fqdn       = "${local.prefix}-saas-sso.example.com"

  conflicting_fqdn = "${local.prefix}-conflicting.example.com"
  bypass_fqdn      = "${local.prefix}-bypass.internal.example.com"

  web_cat_fqdn = "${local.prefix}-web-cat.example.com"
  web_def_fqdn = "${local.prefix}-web-def.example.com"
  ztna_a_fqdn  = "${local.prefix}-ztna-a.internal.example.com"
  ztna_b_fqdn  = "${local.prefix}-ztna-b.internal.example.com"

  # ZTNA destination-subtype / protocol variants.
  ztna_cidr       = "10.60.0.0/24"
  ztna_range      = "10.60.1.1-10.60.1.50"
  ztna_udp_fqdn   = "${local.prefix}-ztna-udp.internal.example.com"
  ztna_mixed_fqdn = "${local.prefix}-ztna-mixed.internal.example.com"

  # Enrichment cases derived from a real customer configuration.
  web_cp_fqdn   = "${local.prefix}-web-cp.example.com"
  web_wild_fqdn = "${local.prefix}-web-wild.example.com"
  ztna_mp_fqdn  = "${local.prefix}-ztna-mp.internal.example.com"

  # Web apps exercising an http:// scheme and a path-bearing url.
  web_http_fqdn = "${local.prefix}-web-http.example.com"
  web_path_fqdn = "${local.prefix}-web-path.example.com"

  # Synthetic resource-location UUID. The API does not validate location UUIDs,
  # so a fixed literal value is stored verbatim and is identical in every tenant
  # — keeping the migrate scenario stable (a real per-tenant UUID would differ
  # between Customer A and Customer B and break the A→B comparison).
  mock_location_uuid = "00000000-0000-0000-0000-0000000000aa"
  ztna_loc_fqdn      = "${local.prefix}-ztna-loc.internal.example.com"

  # --- Repetition for the -ExtractLocals tooling pass ------------------------
  # `spa_manager.ps1 -List -ExtractLocals` hoists resource locations and
  # user/group identities that repeat across the generated config into a
  # `locals` block. It reads the tenant through the API, never this file, so the
  # repetition has to exist in the tenant itself — these values are deliberately
  # referenced from several routing domains, applications and policies below,
  # and e2e/tooling/run-spa-manager-roundtrip.sh counts the resulting collapse.
  #
  # A resource location is only collapsible if every application reports it
  # under the SAME name, hence the shared name locals rather than inline
  # strings.
  mock_location_uuid_b = "00000000-0000-0000-0000-0000000000bb"
  mock_location_name   = "${local.prefix}-mock-location"
  mock_location_name_b = "${local.prefix}-mock-location-b"

  ztna_loc_b_fqdn  = "${local.prefix}-ztna-loc-b.internal.example.com"
  ztna_loc_ab_fqdn = "${local.prefix}-ztna-loc-ab.internal.example.com"

  # Two identities reused across policies. metadata is derived with join() for
  # the same reason the generated locals block derives it: the display-name map
  # and values[] must never disagree.
  shared_group_tokens = [
    "OID:/azuread/00000000-0000-0000-0000-000000000001",
    "SID:/example.com/S-1-5-21-1111111111-2222222222-3333333333-1001",
  ]
  second_group_tokens = ["OID:/azuread/00000000-0000-0000-0000-000000000002"]

  shared_group_metadata = { "Golden Shared Group" = join(",", local.shared_group_tokens) }
  second_group_metadata = { "Golden Second Group" = join(",", local.second_group_tokens) }
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

# --- ZTNA application (with a destination) ------------------------------------
resource "citrixspa_routing_domain" "ztna" {
  fqdn         = local.ztna_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ZTNA routing domain"
}

resource "citrixspa_application" "ztna" {
  name           = "${local.prefix}-ztna"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_fqdn
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna]
}

# --- Web application: non-default flags, keywords, multiple related URLs ------
resource "citrixspa_routing_domain" "web_flags" {
  fqdn         = local.web_flags_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-flags routing domain"
}

resource "citrixspa_routing_domain" "web_flags_related" {
  fqdn         = local.web_flags_rel_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-flags related routing domain"
}

resource "citrixspa_application" "web_flags" {
  name             = "${local.prefix}-web-flags"
  type             = "web"
  state            = "complete"
  description      = "Golden web application with non-default flags and keywords"
  url              = "https://${local.web_flags_fqdn}"
  using_template   = false
  hidden           = true
  agentless_access = true
  mobile_security  = true
  sbs_only_launch  = true
  keywords         = ["golden", "e2e", "web"]
  related_urls     = [local.web_flags_fqdn, local.web_flags_rel_fqdn]
  icon             = local.icon

  depends_on = [
    citrixspa_routing_domain.web_flags,
    citrixspa_routing_domain.web_flags_related,
  ]
}

# --- Web application intentionally left in the "incomplete" state -------------
# No routing domain is created for its FQDN: the app stays incomplete, which is
# the case under test. related_urls is required by the API even when incomplete.
resource "citrixspa_application" "web_incomplete" {
  name           = "${local.prefix}-web-incomplete"
  type           = "web"
  state          = "incomplete"
  description    = "Golden web application intentionally left incomplete"
  url            = "https://${local.web_incomplete_fqdn}"
  using_template = false
  related_urls   = [local.web_incomplete_fqdn]
  icon           = local.icon
}

# --- SaaS application with SAML SSO ------------------------------------------
resource "citrixspa_application" "saas_sso" {
  name           = "${local.prefix}-saas-sso"
  type           = "saas"
  description    = "Golden SaaS application with SAML SSO"
  url            = "https://${local.saas_sso_fqdn}"
  using_template = false
  related_urls   = ["*.${local.saas_sso_fqdn}"]
  icon           = local.icon

  sso = {
    type              = "saml"
    assertion_url     = "https://sp.example.com/acs"
    audience          = "https://sp.example.com"
    name_id_format    = "emailAddress"
    name_id_source    = "email"
    custom_attributes = []
  }
}

# --- Web application: category + keywords ------------------------------------
resource "citrixspa_routing_domain" "web_cat" {
  fqdn         = local.web_cat_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-cat routing domain"
}

resource "citrixspa_application" "web_cat" {
  name           = "${local.prefix}-web-cat"
  type           = "web"
  state          = "complete"
  description    = "Golden web application with category and keywords"
  url            = "https://${local.web_cat_fqdn}"
  category       = "Productivity"
  using_template = false
  keywords       = ["golden", "web", "category"]
  related_urls   = [local.web_cat_fqdn]
  icon           = local.icon

  depends_on = [citrixspa_routing_domain.web_cat]
}

# --- Web application: all flags explicitly false -----------------------------
resource "citrixspa_routing_domain" "web_def" {
  fqdn         = local.web_def_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-def routing domain"
}

resource "citrixspa_application" "web_defaults" {
  name             = "${local.prefix}-web-def"
  type             = "web"
  state            = "complete"
  description      = "Golden web application with all flags explicitly false"
  url              = "https://${local.web_def_fqdn}"
  using_template   = false
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false
  related_urls     = [local.web_def_fqdn]
  icon             = local.icon

  depends_on = [citrixspa_routing_domain.web_def]
}

# --- ZTNA application: multiple destinations on non-default ports -------------
resource "citrixspa_routing_domain" "ztna_a" {
  fqdn         = local.ztna_a_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ztna-a routing domain"
}

resource "citrixspa_routing_domain" "ztna_b" {
  fqdn         = local.ztna_b_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ztna-b routing domain"
}

resource "citrixspa_application" "ztna_multi" {
  name           = "${local.prefix}-ztna-multi"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with multiple destinations"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_a_fqdn
      port        = "8080"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    },
    {
      destination = local.ztna_b_fqdn
      port        = "9090"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [
    citrixspa_routing_domain.ztna_a,
    citrixspa_routing_domain.ztna_b,
  ]
}

# --- SaaS application: no SSO -------------------------------------------------
resource "citrixspa_application" "saas_nosso" {
  name           = "${local.prefix}-saas-nosso"
  type           = "saas"
  description    = "Golden SaaS application with no SSO"
  url            = "https://${local.prefix}-saas-nosso.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-nosso.example.com"]
  icon           = local.icon

  sso = { type = "nosso" }
}

# --- SaaS application: basic SSO ---------------------------------------------
resource "citrixspa_application" "saas_basic" {
  name           = "${local.prefix}-saas-basic"
  type           = "saas"
  description    = "Golden SaaS application with basic SSO"
  url            = "https://${local.prefix}-saas-basic.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-basic.example.com"]
  icon           = local.icon

  sso = {
    type            = "basic"
    username_format = "userPrincipalName"
  }
}

# --- SaaS application: Kerberos SSO ------------------------------------------
resource "citrixspa_application" "saas_kerberos" {
  name           = "${local.prefix}-saas-kerberos"
  type           = "saas"
  description    = "Golden SaaS application with Kerberos SSO"
  url            = "https://${local.prefix}-saas-kerberos.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-kerberos.example.com"]
  icon           = local.icon

  sso = {
    type            = "kerberos"
    user_realm      = "EXAMPLE.COM"
    username_format = "userPrincipalName"
  }
}

# --- SaaS application: form SSO ----------------------------------------------
resource "citrixspa_application" "saas_form" {
  name           = "${local.prefix}-saas-form"
  type           = "saas"
  description    = "Golden SaaS application with form SSO"
  url            = "https://${local.prefix}-saas-form.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-form.example.com"]
  icon           = local.icon

  sso = {
    type            = "form"
    action_url      = "https://${local.prefix}-saas-form.example.com/login"
    logonform_url   = "https://${local.prefix}-saas-form.example.com/logon"
    username_field  = "username"
    password_field  = "password"
    attribute       = "email"
    username_format = "userPrincipalName"
  }
}

# --- SaaS application: SAML SSO (full) ---------------------------------------
resource "citrixspa_application" "saas_saml_full" {
  name           = "${local.prefix}-saas-saml-full"
  type           = "saas"
  description    = "Golden SaaS application with full SAML SSO"
  url            = "https://${local.prefix}-saas-saml-full.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-saml-full.example.com"]
  icon           = local.icon

  sso = {
    type              = "saml"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
    assertion_url     = "https://sp.example.com/acs"
    audience          = "https://sp.example.com"
    relay_state       = "https://sp.example.com/home"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    custom_attributes = [
      {
        format      = "uri"
        name        = "role"
        value       = "admin"
        prefix_expr = false
      },
      {
        format = "basic"
        name   = "department"
        value  = "engineering"
      }
    ]
  }
}

# --- SaaS application: SAML SSO (variant) ------------------------------------
resource "citrixspa_application" "saas_saml_variant" {
  name           = "${local.prefix}-saas-saml-variant"
  type           = "saas"
  description    = "Golden SaaS application with SAML SSO variant"
  url            = "https://${local.prefix}-saas-saml-variant.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-saml-variant.example.com"]
  icon           = local.icon

  sso = {
    type           = "saml"
    saml_type      = "IDP"
    assertion_url  = "https://sp2.example.com/acs"
    audience       = "https://sp2.example.com"
    name_id_source = "upn"
    name_id_format = "persistent"
    sign_assertion = "RESPONSE"
  }
}

# --- ZTNA application: IP-and-CIDR destination subtype -----------------------
resource "citrixspa_routing_domain" "ztna_cidr" {
  fqdn         = local.ztna_cidr
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = true
  location_ids = []
  comment      = "Golden ztna CIDR routing domain"
}

resource "citrixspa_application" "ztna_cidr" {
  name           = "${local.prefix}-ztna-cidr"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with an IP-and-CIDR destination"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_cidr
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_IP_AND_CIDR"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_cidr]
}

# --- ZTNA application: IP-range destination subtype --------------------------
resource "citrixspa_routing_domain" "ztna_range" {
  fqdn         = local.ztna_range
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = true
  location_ids = []
  comment      = "Golden ztna IP-range routing domain"
}

resource "citrixspa_application" "ztna_range" {
  name           = "${local.prefix}-ztna-range"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with an IP-range destination"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_range
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_IP_RANGE"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_range]
}

# --- ZTNA application: UDP protocol destination ------------------------------
resource "citrixspa_routing_domain" "ztna_udp" {
  fqdn         = local.ztna_udp_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ztna UDP routing domain"
}

resource "citrixspa_application" "ztna_udp" {
  name           = "${local.prefix}-ztna-udp"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with a UDP destination"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_udp_fqdn
      port        = "53"
      protocol    = "PROTOCOL_UDP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_udp]
}

# --- ZTNA application: mixed TCP + UDP destinations --------------------------
resource "citrixspa_routing_domain" "ztna_mixed" {
  fqdn         = local.ztna_mixed_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ztna mixed-protocol routing domain"
}

resource "citrixspa_application" "ztna_mixed" {
  name           = "${local.prefix}-ztna-mixed"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with mixed TCP and UDP destinations"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_mixed_fqdn
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    },
    {
      destination = local.ztna_mixed_fqdn
      port        = "53"
      protocol    = "PROTOCOL_UDP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_mixed]
}

# --- SaaS application: created from a template -------------------------------
resource "citrixspa_application" "saas_template" {
  name           = "${local.prefix}-saas-template"
  type           = "saas"
  description    = "Golden SaaS application created from a template"
  url            = "https://${local.prefix}-saas-template.example.com"
  using_template = true
  template_name  = "Office365"
  related_urls   = ["*.${local.prefix}-saas-template.example.com"]
  icon           = local.icon
}

# --- Web application: custom_properties (connector stickiness) ---------------
resource "citrixspa_routing_domain" "web_cp" {
  fqdn         = local.web_cp_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web custom-properties routing domain"
}

resource "citrixspa_application" "web_cp" {
  name           = "${local.prefix}-web-cp"
  type           = "web"
  state          = "complete"
  description    = "Golden web application with connector-stickiness custom property"
  url            = "https://${local.web_cp_fqdn}"
  using_template = false
  related_urls   = [local.web_cp_fqdn]
  icon           = local.icon

  custom_properties = {
    stickiness_type = "Connector"
  }

  depends_on = [citrixspa_routing_domain.web_cp]
}

# --- Web application: wildcard related URL (apex RD + wildcard RD) ------------
# A complete web app whose url apex host and wildcard related URL each need
# their own routing domain entry.
resource "citrixspa_routing_domain" "web_wild_apex" {
  fqdn         = local.web_wild_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-wild apex routing domain"
}

resource "citrixspa_routing_domain" "web_wild" {
  fqdn         = "*.${local.web_wild_fqdn}"
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-wild wildcard routing domain"
}

resource "citrixspa_application" "web_wild" {
  name           = "${local.prefix}-web-wild"
  type           = "web"
  state          = "complete"
  description    = "Golden web application with a wildcard related URL"
  url            = "https://${local.web_wild_fqdn}"
  using_template = false
  related_urls   = [local.web_wild_fqdn, "*.${local.web_wild_fqdn}"]
  icon           = local.icon

  depends_on = [
    citrixspa_routing_domain.web_wild_apex,
    citrixspa_routing_domain.web_wild,
  ]
}

# --- SaaS application: Kerberos SSO (samaccountname username format) ----------
resource "citrixspa_application" "saas_kerberos_sam" {
  name           = "${local.prefix}-saas-kerberos-sam"
  type           = "saas"
  description    = "Golden SaaS application with Kerberos SSO samaccountname format"
  url            = "https://${local.prefix}-saas-kerberos-sam.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-kerberos-sam.example.com"]
  icon           = local.icon

  sso = {
    type            = "kerberos"
    user_realm      = "example.com"
    username_format = "samaccountname"
  }
}

# --- SaaS application: SAML SSO with an empty audience ------------------------
resource "citrixspa_application" "saas_saml_empty_aud" {
  name           = "${local.prefix}-saas-saml-ea"
  type           = "saas"
  description    = "Golden SaaS application with SAML SSO and empty audience"
  url            = "https://${local.prefix}-saas-saml-ea.example.com"
  using_template = false
  related_urls   = ["*.${local.prefix}-saas-saml-ea.example.com"]
  icon           = local.icon

  sso = {
    type              = "saml"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
    assertion_url     = "https://sp3.example.com/acs"
    audience          = ""
    relay_state       = "https://sp3.example.com/home"
    name_id_source    = "upn"
    name_id_format    = "emailAddress"
    custom_attributes = []
  }
}

# --- ZTNA application: multiple ports on a single host ------------------------
resource "citrixspa_routing_domain" "ztna_mp" {
  fqdn         = local.ztna_mp_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden ztna multi-port routing domain"
}

resource "citrixspa_application" "ztna_mp" {
  name           = "${local.prefix}-ztna-mp"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application with multiple ports on one host"
  using_template = false
  icon           = local.icon

  destination = [
    {
      destination = local.ztna_mp_fqdn
      port        = "80"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    },
    {
      destination = local.ztna_mp_fqdn
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_mp]
}

# --- Web application: http:// scheme url --------------------------------------
resource "citrixspa_routing_domain" "web_http" {
  fqdn         = local.web_http_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-http routing domain"
}

resource "citrixspa_application" "web_http" {
  name           = "${local.prefix}-web-http"
  type           = "web"
  state          = "complete"
  description    = "Golden web application served over an http scheme"
  url            = "http://${local.web_http_fqdn}"
  using_template = false
  related_urls   = [local.web_http_fqdn]
  icon           = local.icon

  depends_on = [citrixspa_routing_domain.web_http]
}

# --- Web application: path-bearing url ----------------------------------------
resource "citrixspa_routing_domain" "web_path" {
  fqdn         = local.web_path_fqdn
  type         = "external"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden web-path routing domain"
}

resource "citrixspa_application" "web_path" {
  name           = "${local.prefix}-web-path"
  type           = "web"
  state          = "complete"
  description    = "Golden web application with a path-bearing url"
  url            = "https://${local.web_path_fqdn}/login/index.php"
  using_template = false
  related_urls   = [local.web_path_fqdn]
  icon           = local.icon

  depends_on = [citrixspa_routing_domain.web_path]
}

# --- ZTNA routing domain + application carrying a resource location ----------
# Exercises RD `location_ids` and app `locations` using a synthetic UUID (see
# local.mock_location_uuid). The API does not validate the UUID, so the fixed
# literal round-trips cleanly and stays identical across tenants for migrate.
resource "citrixspa_routing_domain" "ztna_loc" {
  fqdn         = local.ztna_loc_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = [local.mock_location_uuid]
  comment      = "Golden ztna routing domain bound to a resource location"
}

resource "citrixspa_application" "ztna_loc" {
  name           = "${local.prefix}-ztna-loc"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application bound to a resource location"
  using_template = false
  icon           = local.icon

  locations = [
    {
      name = local.mock_location_name
      uuid = local.mock_location_uuid
    }
  ]

  destination = [
    {
      destination = local.ztna_loc_fqdn
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_loc]
}

# --- Second and combined resource-location bindings --------------------------
# Together with citrixspa_routing_domain.ztna_loc and
# citrixspa_application.ztna_loc above, these spread two location UUIDs over
# three routing domains and two applications. Without that repetition the
# -ExtractLocals pass in the tooling round-trip would have nothing to collapse
# and its assertions would pass vacuously.
resource "citrixspa_routing_domain" "ztna_loc_b" {
  fqdn         = local.ztna_loc_b_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = [local.mock_location_uuid_b]
  comment      = "Golden ztna routing domain bound to the second resource location"
}

resource "citrixspa_routing_domain" "ztna_loc_ab" {
  fqdn         = local.ztna_loc_ab_fqdn
  type         = "internal"
  app_type     = "ztna"
  flag         = "enabled"
  ip           = false
  location_ids = [local.mock_location_uuid, local.mock_location_uuid_b]
  comment      = "Golden ztna routing domain bound to both resource locations"
}

resource "citrixspa_application" "ztna_loc_ab" {
  name           = "${local.prefix}-ztna-loc-ab"
  type           = "ztna"
  state          = "complete"
  description    = "Golden ZTNA application bound to both resource locations"
  using_template = false
  icon           = local.icon

  locations = [
    {
      name = local.mock_location_name
      uuid = local.mock_location_uuid
    },
    {
      name = local.mock_location_name_b
      uuid = local.mock_location_uuid_b
    }
  ]

  destination = [
    {
      destination = local.ztna_loc_ab_fqdn
      port        = "443"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    }
  ]

  depends_on = [citrixspa_routing_domain.ztna_loc_ab]
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

# --- Inactive access policy with a condition (SaaS app) ----------------------
resource "citrixspa_access_policy" "saas_inactive" {
  name        = "${local.prefix}-saas-policy"
  description = "Golden inactive access policy with a platform condition"
  active      = false
  priority    = 2

  apps = [citrixspa_application.saas.id]

  access_rules = [
    {
      name        = "deny-conditional"
      description = ""
      priority    = 1
      active      = true
      access      = "ACCESS_DENY"

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]
}

# --- Access policy with browser/security restrictions ------------------------
# Exercises access_native, per-rule restrictions (enhanced_security_settings +
# redirect_sbs), and rule metadata. Attached to the web-flags application.
resource "citrixspa_access_policy" "restrict" {
  name        = "${local.prefix}-restrict-policy"
  description = "Golden access policy with browser/security restrictions"
  active      = true
  priority    = 3

  apps = [citrixspa_application.web_flags.id]

  access_rules = [
    {
      name          = "allow-embedded-browser"
      description   = ""
      priority      = 1
      active        = true
      access        = "ACCESS_ALLOW"
      access_native = "ACCESS_ALLOW"

      restrictions = {
        redirect_sbs = false
        # Mixed on purpose: download/upload/clipboard/printing = "disabled" are
        # non-default and must round-trip (restrictive coverage), while
        # watermarkV1 = "disabled" IS the backend default and is omitted from the
        # API response — exercising the provider's default-key reconciliation.
        enhanced_security_settings = {
          _browserV1      = "embeddedBrowser"
          watermarkV1     = "disabled"
          downloadV1      = "disabled"
          uploadV1        = "disabled"
          clipboardV1     = "disabled"
          printingV1      = "disabled"
          keyLoggingV1    = "disabled"
          screenCaptureV1 = "disabled"
          proxyTrafficV1  = "direct"
        }
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["EMAIL:/e2e/golden-user@example.com"]
          metadata = {
            display = "Golden Group"
          }
        }
      ]
    },
    {
      name          = "allow-insecure-content"
      description   = ""
      priority      = 2
      active        = true
      access        = "ACCESS_ALLOW"
      access_native = "ACCESS_ALLOW"

      restrictions = {
        redirect_sbs = false
        enhanced_security_settings = {
          insecure_content_allowed_for_urls_v1 = "enabled"
          _browserV1                           = "embeddedBrowser"
        }
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]
}

# --- Access policy with OID/SID identity values ------------------------------
# Exercises identity-provider (OID:) and Windows SID (SID:) usergroup values.
resource "citrixspa_access_policy" "identity" {
  name        = "${local.prefix}-identity-policy"
  description = "Golden access policy with OID and SID identity values"
  active      = true
  priority    = 4

  apps = [citrixspa_application.web_cat.id]

  access_rules = [
    {
      name        = "allow-identity"
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
          values = [
            "OID:/azuread/00000000-0000-0000-0000-000000000001",
            "SID:/example.com/S-1-5-21-1111111111-2222222222-3333333333-1001",
          ]
        }
      ]
    }
  ]
}

# --- Access policy: multiple heterogeneous rule types in one access_rule -----
# Exercises TYPE_USERGROUP + TYPE_TAG (geo/ITM) + TYPE_MULTIURLDOMAIN together.
resource "citrixspa_access_policy" "multi_type" {
  name        = "${local.prefix}-multitype-policy"
  description = "Golden access policy with multiple heterogeneous rule types"
  active      = true
  priority    = 5

  apps = [citrixspa_application.web_defaults.id]

  access_rules = [
    {
      name        = "multi-type"
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
          values     = ["Everyone"]
        },
        {
          type       = "TYPE_TAG"
          operator   = "OPERATOR_IN"
          tag_source = "ITM"
          tag_key    = "location-geo-country-isocode"
          values     = ["AF"]
          metadata = {
            AF = "Afghanistan"
          }
        },
        {
          type       = "TYPE_MULTIURLDOMAIN"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["${local.prefix}-portal.cloud.com"]
        }
      ]
    }
  ]
}

# --- Access policy: platform_filter variants (PC + MOBILE) -------------------
resource "citrixspa_access_policy" "platforms" {
  name        = "${local.prefix}-platforms-policy"
  description = "Golden access policy with PC and mobile platform conditions"
  active      = true
  priority    = 6

  apps = [citrixspa_application.web_http.id]

  access_rules = [
    {
      name        = "pc-and-mobile"
      description = ""
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_PC"
        },
        {
          platform_filter = "PLATFORM_FILTER_MOBILE"
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]
}

# --- Access policy: multiple apps + mixed ACCESS_ALLOW / ACCESS_DENY ---------
resource "citrixspa_access_policy" "mixed" {
  name        = "${local.prefix}-mixed-policy"
  description = "Golden access policy spanning multiple apps with mixed access"
  active      = true
  priority    = 7

  apps = [
    citrixspa_application.web_path.id,
    citrixspa_application.web_cp.id,
  ]

  access_rules = [
    {
      name        = "allow-first"
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
    },
    {
      name        = "deny-rest"
      description = ""
      priority    = 2
      active      = true
      access      = "ACCESS_DENY"

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]
}

# --- Access policy: TYPE_MULTIURLDOMAIN with OPERATOR_NOT ---------------------
# MULTIURLDOMAIN cannot stand alone; it needs a companion user/machine group rule.
resource "citrixspa_access_policy" "urldomain_not" {
  name        = "${local.prefix}-urlnot-policy"
  description = "Golden access policy with a MULTIURLDOMAIN OPERATOR_NOT rule"
  active      = true
  priority    = 8

  apps = [citrixspa_application.web_wild.id]

  access_rules = [
    {
      name        = "url-not"
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
          values     = ["Everyone"]
        },
        {
          type       = "TYPE_MULTIURLDOMAIN"
          operator   = "OPERATOR_NOT"
          tag_source = ""
          tag_key    = ""
          values     = ["${local.prefix}-blocked.cloud.com"]
        }
      ]
    }
  ]
}

# --- Access policy: platform condition + user-group rule ---------------------
# (user_and_groups condition deliberately omitted — see the note on the rule below)
resource "citrixspa_access_policy" "user_groups" {
  name        = "${local.prefix}-usergroups-policy"
  description = "Golden access policy with a platform condition and a user-group rule"
  active      = true
  priority    = 9

  apps = [citrixspa_application.saas_nosso.id]

  access_rules = [
    {
      name        = "ug-condition"
      description = ""
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"

      # user_and_groups map intentionally omitted. The provider translates it
      # into the TYPE_USERGROUP rule below, and the translated tokens flow into
      # SPAConfig's feature-flag-gated user/group validation. Restore it only
      # after the matching SPAConfig change is confirmed live in production, so
      # a flag rollout cannot turn this golden scenario red.
      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]
}

# --- Access policy: access rule with NO name (name is optional) --------------
# Regression coverage for the optional access_rules[].name field. The SPA service
# does not require a name on an access rule, so this rule intentionally omits it
# to keep that path exercised on every golden run.
resource "citrixspa_access_policy" "no_rule_name" {
  name        = "${local.prefix}-no-rulename-policy"
  description = "Golden access policy whose access rule omits the optional name"
  active      = true
  priority    = 10

  apps = [citrixspa_application.saas_basic.id]

  access_rules = [
    {
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
          # Resolvable directory tokens plus display-name metadata so the SPA
          # Console edit page can render this rule (a non-resolvable placeholder
          # like "Everyone" with no metadata crashes that page).
          values = [
            "OID:/azuread/00000000-0000-0000-0000-000000000001",
            "SID:/example.com/S-1-5-21-1111111111-2222222222-3333333333-1001",
          ]
          metadata = {
            "display-name" = "OID:/azuread/00000000-0000-0000-0000-000000000001,SID:/example.com/S-1-5-21-1111111111-2222222222-3333333333-1001"
          }
        }
      ]
    }
  ]
}

# --- Access policies reusing one identity across several policies ------------
# The same user/group appearing in more than one policy is the case the
# -ExtractLocals pass is built for, and until these existed no identity in this
# tenant was used twice. The two single-identity policies collapse to
# `local.users.<key>.tokens`; the policy below them carries two identities in one
# rule, which is the only thing that exercises the concat()/merge() shape.
resource "citrixspa_access_policy" "shared_identity_a" {
  name        = "${local.prefix}-shared-identity-a-policy"
  description = "Golden access policy scoped to the shared identity"
  active      = true
  priority    = 11

  apps = [citrixspa_application.ztna_loc.id]

  access_rules = [
    {
      name        = "shared-identity"
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
          values     = local.shared_group_tokens
          metadata   = local.shared_group_metadata
        }
      ]
    }
  ]
}

resource "citrixspa_access_policy" "shared_identity_b" {
  name        = "${local.prefix}-shared-identity-b-policy"
  description = "Golden access policy reusing the shared identity"
  active      = true
  priority    = 12

  apps = [citrixspa_application.ztna_loc_ab.id]

  access_rules = [
    {
      name        = "shared-identity"
      description = ""
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"

      # Exercises the resource-location path too: a domain override pins this
      # rule to a location UUID that also appears on the routing domains above.
      advanced_settings = {
        domain_overrides = [
          {
            fqdn         = local.ztna_loc_ab_fqdn
            type         = "internal"
            location_ids = [local.mock_location_uuid]
          }
        ]
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = local.shared_group_tokens
          metadata   = local.shared_group_metadata
        }
      ]
    }
  ]
}

resource "citrixspa_access_policy" "shared_identity_multi" {
  name        = "${local.prefix}-shared-identity-multi-policy"
  description = "Golden access policy whose rule carries two identities"
  active      = true
  priority    = 13

  apps = [citrixspa_application.saas_basic.id]

  access_rules = [
    {
      name        = "two-identities"
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
          # Order matters: -ExtractLocals only collapses a rule when each
          # identity's tokens sit contiguously in values[] in metadata order.
          values   = concat(local.shared_group_tokens, local.second_group_tokens)
          metadata = merge(local.shared_group_metadata, local.second_group_metadata)
        }
      ]
    }
  ]
}

# --- Security group referencing the web application --------------------------
resource "citrixspa_security_group" "web" {
  name    = "${local.prefix}-web-sg"
  app_ids = [citrixspa_application.web.id]

  system = {
    data_in  = "enabled"
    data_out = "disabled"
  }

  unpublished_app = {
    data_in  = "disabled"
    data_out = "disabled"
  }
}

# --- Security group with omitted data_in/data_out (they default to disabled) --
# Regression coverage for the optional system/unpublished_app data_in/data_out
# fields. The SPA service treats them as optional with a "disabled" default, so
# this group intentionally supplies empty nested objects to keep that path
# exercised on every golden run.
resource "citrixspa_security_group" "defaulted_dataflow" {
  name    = "${local.prefix}-defaulted-sg"
  app_ids = [citrixspa_application.web_defaults.id]

  system          = {}
  unpublished_app = {}
}


# Standalone domains (no application) exercising the less-common routing-domain
# types. Both are enumerable via the list API, so they round-trip through the
# tooling and migrate scenarios. (external_via_connector is intentionally
# excluded: the type is fully retired by the SPA service and is rejected with a
# 400 on create/update, so it can no longer be provisioned at all.)
resource "citrixspa_routing_domain" "conflicting" {
  fqdn         = local.conflicting_fqdn
  type         = "conflicting"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden conflicting routing domain"
}

resource "citrixspa_routing_domain" "bypass" {
  fqdn         = local.bypass_fqdn
  type         = "internal_bypass_proxy"
  app_type     = "web"
  flag         = "enabled"
  ip           = false
  location_ids = []
  comment      = "Golden internal-bypass-proxy routing domain"
}

# --- Session policy ----------------------------------------------------------
resource "citrixspa_session_policy" "golden" {
  name        = "${local.prefix}-session-policy"
  description = "Golden session policy"
  active      = false

  generic_rules = [
    {
      name        = "Default Rule"
      description = "Allow all users"
      priority    = 1
      active      = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        },
        {
          # Same identity as the access policies above. Session-policy
          # conditions are a separate emission path in spa_manager.ps1, so this
          # is the only thing that proves -ExtractLocals reaches them.
          type     = "TYPE_USERGROUP"
          operator = "OPERATOR_IN"
          values   = local.shared_group_tokens
          metadata = local.shared_group_metadata
        }
      ]
    }
  ]
}

# --- Terminate machine access: only the API-required identity fields ---------
# Regression coverage for the optional terminate_machine_access.name field. The
# SPA service identifies a machine by account_name, object_id and idp_type and
# does not require name (nor dns_host_name, domain_name or duration), so this
# resource intentionally omits those to keep the optional path exercised on
# every golden run. object_id is a synthetic, format-valid OID (the service
# validates only the OID format, not a real directory object).
resource "citrixspa_terminate_machine_access" "golden" {
  account_name = "${local.prefix}-machine"
  object_id    = "OID:/ad/${local.prefix}-machine"
  idp_type     = "AD"
}

# --- Terminate user access: only the API-required identity fields -------------
# Regression coverage for the optional terminate_user_access email, domain_name
# and duration fields. The SPA service identifies a user by account_name,
# object_id and idp_type and does not require those, so this resource
# intentionally omits them to keep the optional path exercised on every golden
# run. object_id is a synthetic, format-valid OID.
resource "citrixspa_terminate_user_access" "golden" {
  account_name = "${local.prefix}-user"
  object_id    = "OID:/ad/${local.prefix}-user"
  idp_type     = "AD"
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

output "ztna_application_id" {
  value = citrixspa_application.ztna.id
}

output "ztna_routing_domain_fqdn" {
  value = citrixspa_routing_domain.ztna.fqdn
}

output "web_flags_application_id" {
  value = citrixspa_application.web_flags.id
}

output "web_incomplete_application_id" {
  value = citrixspa_application.web_incomplete.id
}

output "saas_sso_application_id" {
  value = citrixspa_application.saas_sso.id
}

output "web_access_policy_id" {
  value = citrixspa_access_policy.web.id
}

output "saas_access_policy_id" {
  value = citrixspa_access_policy.saas_inactive.id
}

output "web_security_group_id" {
  value = citrixspa_security_group.web.id
}

output "session_policy_id" {
  value = citrixspa_session_policy.golden.id
}

output "terminate_machine_access_id" {
  value = citrixspa_terminate_machine_access.golden.id
}

output "terminate_user_access_id" {
  value = citrixspa_terminate_user_access.golden.id
}

output "web_routing_domain_fqdn" {
  value = citrixspa_routing_domain.web.fqdn
}

output "conflicting_routing_domain_fqdn" {
  value = citrixspa_routing_domain.conflicting.fqdn
}

output "bypass_routing_domain_fqdn" {
  value = citrixspa_routing_domain.bypass.fqdn
}

output "web_cat_application_id" {
  value = citrixspa_application.web_cat.id
}

output "web_defaults_application_id" {
  value = citrixspa_application.web_defaults.id
}

output "ztna_multi_application_id" {
  value = citrixspa_application.ztna_multi.id
}

output "saas_nosso_application_id" {
  value = citrixspa_application.saas_nosso.id
}

output "saas_basic_application_id" {
  value = citrixspa_application.saas_basic.id
}

output "saas_kerberos_application_id" {
  value = citrixspa_application.saas_kerberos.id
}

output "saas_form_application_id" {
  value = citrixspa_application.saas_form.id
}

output "saas_saml_full_application_id" {
  value = citrixspa_application.saas_saml_full.id
}

output "saas_saml_variant_application_id" {
  value = citrixspa_application.saas_saml_variant.id
}

output "ztna_cidr_application_id" {
  value = citrixspa_application.ztna_cidr.id
}

output "ztna_range_application_id" {
  value = citrixspa_application.ztna_range.id
}

output "ztna_udp_application_id" {
  value = citrixspa_application.ztna_udp.id
}

output "ztna_mixed_application_id" {
  value = citrixspa_application.ztna_mixed.id
}

output "saas_template_application_id" {
  value = citrixspa_application.saas_template.id
}

output "web_cp_application_id" {
  value = citrixspa_application.web_cp.id
}

output "web_wild_application_id" {
  value = citrixspa_application.web_wild.id
}

output "saas_kerberos_sam_application_id" {
  value = citrixspa_application.saas_kerberos_sam.id
}

output "saas_saml_empty_aud_application_id" {
  value = citrixspa_application.saas_saml_empty_aud.id
}

output "ztna_mp_application_id" {
  value = citrixspa_application.ztna_mp.id
}

output "restrict_access_policy_id" {
  value = citrixspa_access_policy.restrict.id
}

output "web_http_application_id" {
  value = citrixspa_application.web_http.id
}

output "web_path_application_id" {
  value = citrixspa_application.web_path.id
}

output "identity_access_policy_id" {
  value = citrixspa_access_policy.identity.id
}

output "multi_type_access_policy_id" {
  value = citrixspa_access_policy.multi_type.id
}

output "platforms_access_policy_id" {
  value = citrixspa_access_policy.platforms.id
}

output "mixed_access_policy_id" {
  value = citrixspa_access_policy.mixed.id
}

output "urldomain_not_access_policy_id" {
  value = citrixspa_access_policy.urldomain_not.id
}

output "user_groups_access_policy_id" {
  value = citrixspa_access_policy.user_groups.id
}

output "ztna_loc_application_id" {
  value = citrixspa_application.ztna_loc.id
}

output "ztna_loc_routing_domain_fqdn" {
  value = citrixspa_routing_domain.ztna_loc.fqdn
}
