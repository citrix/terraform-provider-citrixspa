---
page_title: "citrixspa_application Resource - citrixspa"
description: |-
  Resource for creating and managing SPA applications.
---

# citrixspa_application (Resource)

Resource for creating and managing SPA applications. Supports web, SaaS, and ZTNA application types.

For more details on the underlying API, see the [Applications API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-applications).

-> **Tip** Ready-to-adapt templates covering the full SPA SaaS app catalog (one file per vendor, grouped by SSO type under `saml/` and `form/`) are available in the [`templates/`](../../templates/README.md) directory. Each file defines a `citrixspa_application` together with the `citrixspa_routing_domain` resources it needs.

~> **Note** Application names should be unique within an account.

Applications can be created in an `incomplete` state and later transitioned to `complete` by setting the `state` attribute to `"complete"`.

### Routing Domain Dependencies

Applications rely on routing domains derived from their `url` and `related_urls`. When managing both resources in the same configuration, you must declare explicit dependencies so that routing domains are created before the application:

```terraform
resource "citrixspa_application" "my_web_application" {
  name = "My Web Application"
  type = "web"
  url  = "https://example.com"
  related_urls   = ["api.example.com"]
  # ...

  depends_on = [
    citrixspa_routing_domain.example_com,
    citrixspa_routing_domain.api_example_com,
  ]
}
```

-> **Note** When using the migration script (`spa_manager.ps1`), dependency ordering is handled automatically.

#### Deleting an application also deletes matching routing domains

The dependency runs in both directions. When an application is deleted, Citrix Secure Private Access also deletes the routing domains whose `fqdn` matches one of that application's `url` or `related_urls` (each reduced to its hostname) or `destination` values — unless another live application still references the same FQDN. The match is made on the FQDN string alone, so a routing domain created independently is removed too. The full rule is documented as [Note 3 on the `citrixspa_routing_domain` page](routing_domain.md).

`terraform destroy` of a configuration that declares both is unaffected: `depends_on` destroys the application first, and the routing domain's own delete then succeeds against an already-removed backend entry. Drift appears in two other situations:

- **The application is removed from the configuration while a `citrixspa_routing_domain` for one of its FQDNs is kept.** The backend deletes that routing domain. On the next `terraform plan`, the refresh finds it missing, warns, and plans to create it again; the state file itself is corrected on the following `terraform apply`.
- **The application and the routing domains are managed in separate configurations or state files.** The same thing happens, but the team that owns the routing-domain configuration sees unexplained drift without having changed anything. If the two are managed separately, coordinate application deletions with that owner.

In both cases `terraform state rm` is not required. If the routing domain is no longer needed, remove its resource block from the configuration; if it is still required, the next `terraform apply` recreates it.

When two applications share an FQDN and only one is deleted, the routing domain is retained and there is no drift. The provider still warns on the delete, because it cannot see the account's other applications.

### Field Requirements by Application Type

| Field | `web` | `saas` | `ztna` |
|---|---|---|---|
| `url` | Required | Required | Not used |
| `related_urls` | Required | Required | Not used |
| `destination` | Not used | Not used | Required |

## Example Usage

### Web Application

```terraform
resource "citrixspa_application" "my_web_application" {
  name        = "My Web Application"
  type        = "web"
  state       = "complete"
  description = "A sample web application"
  url         = "https://example.com"
  related_urls    = ["*.example.com"]

  icon = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAC1HAwCAAAAC0lEQVQIW2NgAAIAAAUAAdafFs0AAAAASUVORK5CYII="

  hidden           = false
  agentless_access = false
  mobile_security  = false
  using_template   = false
  sbs_only_launch  = false

  sso = { type = "nosso" }

  locations = [
    {
      name = "Resource Location 1"
      uuid = "00000000-0000-0000-0000-000000000000"
    }
  ]

  depends_on = [
    citrixspa_routing_domain.example_com,
    citrixspa_routing_domain.wildcard_example_com
  ]
}
```

### ZTNA Application

```terraform
resource "citrixspa_application" "ztna_app" {
  name        = "Internal Database"
  type        = "ztna"
  state       = "complete"
  description = "Internal database server"

  destination = [
    {
      destination = "database.internal.com"
      port        = "5432"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_HOSTNAME"
    },
    {
      destination = "10.0.1.0/24"
      port        = "3306"
      protocol    = "PROTOCOL_TCP"
      subtype     = "SUBTYPE_IP_AND_CIDR"
    }
  ]

  icon            = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAC1HAwCAAAAC0lEQVQIW2NgAAIAAAUAAdafFs0AAAAASUVORK5CYII="

  hidden           = false
  agentless_access = false
  mobile_security  = false
  using_template   = false
  sbs_only_launch  = false

  depends_on = [
    citrixspa_routing_domain.database_internal_com,
    citrixspa_routing_domain.routing_domain_10_0_1_0_24
  ]
}
```

### SaaS Application

-> **Note** This example sets `using_template = true`, which provisions the application from a template in the SPA service's built-in application catalog (maintained on the backend) selected by `template_name`. You can still supply `sso`, `icon`, and `related_urls` in the same configuration; they are applied on top of the catalog template, so the app is created fully configured in a single apply. Ready-to-adapt per-vendor examples are in the [`templates/`](../../templates/README.md) directory.

```terraform
resource "citrixspa_application" "saas_app" {
  name        = "Office 365"
  type        = "saas"
  state = "complete"
  description = "Microsoft Office 365 Suite"
  url         = "https://office.com"

  related_urls    = [
    "*.office.com",
    "*.outlook.office.com",
    "*.teams.microsoft.com"
  ]

  icon  = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAC1HAwCAAAAC0lEQVQIW2NgAAIAAAUAAdafFs0AAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Office365"
  sbs_only_launch  = false
  hidden           = false
  agentless_access = false
  mobile_security  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.microsoftonline.com/login.srf"
    audience          = "urn:federation:MicrosoftOnline"
    sign_assertion    = "ASSERTION"
    name_id_source    = "guid_b64"
    name_id_format    = "persistent"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  locations = [
    {
      name = "Resource Location 1"
      uuid = "00000000-0000-0000-0000-000000000000"
    }
  ]

  depends_on = [
    citrixspa_routing_domain.office_com,
    citrixspa_routing_domain.wildcard_office_com,
    citrixspa_routing_domain.wildcard_outlook_office_com,
    citrixspa_routing_domain.wildcard_teams_microsoft_com
  ]
}
```

### SAML SSO Configuration

```terraform
resource "citrixspa_application" "saml_app" {
  name        = "SAML Application"
  type        = "saas"
  state       = "complete"
  description = "Application with SAML SSO"
  url         = "https://app.example.com"
  related_urls = ["*.app.example.com"]

  icon             = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAMAAAC1HAwCAAAAC0lEQVQIW2NgAAIAAAUAAdafFs0AAAAASUVORK5CYII="
  using_template   = false
  sbs_only_launch  = false
  hidden           = false
  agentless_access = false
  mobile_security  = false

  sso = {
    type             = "saml"
    saml_type        = "SP"
    assertion_url    = "https://app.example.com/saml/acs"
    audience         = "https://app.example.com"
    name_id_source   = "email"
    name_id_format   = "emailAddress"
    sign_assertion   = "ASSERTION"
    sp_initiated_only = false
    custom_attributes = [
      {
        format = "uri"
        name   = "role"
        value  = "admin"
      }
    ]
  }

  locations = [
    {
      name = "Resource Location 1"
      uuid = "00000000-0000-0000-0000-000000000000"
    }
  ]
}
```

## Schema

### Required

- `name` (String) Name of the application.
- `type` (String) Type of application. Valid values: `web`, `saas`, `ztna`.
- `icon` (String) Base64-encoded icon data for the application.
- `sbs_only_launch` (Boolean) Enable Secure Browser Service (SBS) only launch. Must be explicitly set to avoid provider errors; the backend defaults it to `false` if omitted.

### Optional

- `description` (String) Description of the application.
- `category` (String) Category of the application (e.g., `Productivity`, `Database`).
- `hidden` (Boolean) Whether to hide the application from end users.
- `agentless_access` (Boolean) Enable agentless access to the application.
- `mobile_security` (Boolean) Enable mobile security for the application.
- `url` (String) Application URL. Required for `web` and `saas` applications; must include the `https://` scheme. Not used for `ztna` applications.
- `related_urls` (Set of String) Related URLs for the application. Required for `web` and `saas` applications; must not include `https://` or trailing slashes (e.g., `"example.com"`). Not used for `ztna` applications.
- `using_template` (Boolean) When `true`, the application is provisioned from a template in the SPA service's built-in application catalog, which is maintained on the backend and selected via `template_name`. Any `sso`, `icon`, and `related_urls` you supply are applied on top of the catalog template, so the application can be created fully configured in a single apply. This is unrelated to the reusable configuration files in the provider's `templates/` directory. This flag is owned and normalized by the SPA Console on save, so it can be omitted and the Console-assigned value is adopted into state without producing drift. Optional; defaults to `false`.
- `template_name` (String) Name of the template to use from the SPA service's built-in application catalog (backend). Applies only when `using_template` is `true`.
- `keywords` (Set of String) Keywords associated with the application.
- `locations` (Attributes List) Resource locations associated with the application. (see [below for nested schema](#nestedatt--locations))
- `destination` (Attributes List) Destinations for ZTNA applications. Required for `ztna` applications. (see [below for nested schema](#nestedatt--destination))
- `custom_properties` (Map of String) Custom properties as key-value pairs. Complex values should be JSON-encoded strings.
- `customer_domain_fields` (Map of String) Customer domain fields as key-value pairs.
- `sso` (Attributes) SSO configuration. Set `type` to one of: `saml`, `kerberos`, `basic`, `form`, `nosso`. Only include fields relevant to the chosen SSO type. (see [below for nested schema](#nestedatt--sso))
- `state` (String) Application state. Valid values: `incomplete`, `complete`. Setting to `complete` finalizes the application.

### Read-Only

- `id` (String) GUID identifier of the application.
- `icon_url` (String) URL of the application icon.
- `policies` (Attributes List) Policies associated with the application, computed by the backend. (see [below for nested schema](#nestedatt--policies))
- `policy_count` (String) Number of policies associated with the application.

<a id="nestedatt--locations"></a>
### Nested Schema for `locations`

Required:

- `name` (String) Location name.
- `uuid` (String) Location UUID.

<a id="nestedatt--destination"></a>
### Nested Schema for `destination`

Optional:

- `destination` (String) Destination address. Must match the chosen `subtype`: a hostname for `SUBTYPE_HOSTNAME`; a single IP or CIDR range (e.g., `10.0.0.1` or `10.0.0.0/24`) for `SUBTYPE_IP_AND_CIDR`; or an IP range (e.g., `10.0.0.1-10.0.0.50`) for `SUBTYPE_IP_RANGE`.
- `port` (String) Port number.
- `protocol` (String) Protocol. Valid values: `PROTOCOL_TCP`, `PROTOCOL_UDP`.
- `subtype` (String) Destination subtype. Valid values: `SUBTYPE_HOSTNAME`, `SUBTYPE_IP_AND_CIDR`, `SUBTYPE_IP_RANGE`.

<a id="nestedatt--policies"></a>
### Nested Schema for `policies`

Read-Only:

- `type` (String) Policy type. Valid values: `capability`, `patterns`.
- `data` (Map of String) Policy data as key-value pairs.

<a id="nestedatt--sso"></a>
### Nested Schema for `sso`

Required:

- `type` (String) SSO type: `saml`, `kerberos`, `basic`, `form`, or `nosso`.

Optional (SAML):

- `saml_type` (String) SAML role: `SP`, `IDP`, or `SP_IDP`.
- `sp_initiated_only` (Boolean) Whether SSO is SP-initiated only.
- `assertion_url` (String) SAML assertion consumer service (ACS) URL.
- `audience` (String) SAML audience (entity ID of the service provider).
- `relay_state` (String) SAML relay state URL.
- `sign_assertion` (String) SAML signature scope: `ASSERTION`, `BOTH`, `NONE`, or `RESPONSE`.
- `name_id_source` (String) SAML NameID source: `email`, `upn`, `name`, `guid_b64`, or `sam`.
- `name_id_format` (String) SAML NameID format: `unspecified`, `emailAddress`, `persistent`, `transient`, `WindowsDomainQualifiedName`, or `X509SubjectName`.
- `custom_attributes` (Attributes List) SAML custom attributes (max 16). (see [below for nested schema](#nestedatt--sso--custom_attributes))

Optional (Form):

- `action_url` (String) Form SSO action URL.
- `logonform_url` (String) Form SSO logon form URL.
- `username_field` (String) Form SSO username HTML field name.
- `password_field` (String) Form SSO password HTML field name.
- `attribute` (String) Form SSO attribute (e.g., `email`, `upn`, `name`).

Optional (Kerberos):

- `user_realm` (String) Kerberos user realm.

Optional (Shared — Form, Kerberos, Basic):

- `username_format` (String) Username format.

Read-Only:

- `saml_sso_login_url` (String) SAML SSO login URL (computed by server).
- `saml_cert_issuer_name` (String) SAML certificate issuer name (computed by server).
- `customer` (String) Customer ID associated with the SSO configuration (computed by server).

<a id="nestedatt--sso--custom_attributes"></a>
### Nested Schema for `sso.custom_attributes`

Required:

- `name` (String) Attribute name.
- `value` (String) Attribute value.

Optional:

- `format` (String) Attribute format: `uri`, `unspecified`, or `basic`.
- `prefix_expr` (Boolean) Whether to use prefix expression.

## Import

Import is supported using the application ID:

```shell
terraform import citrixspa_application.web_app 00000000-0000-0000-0000-000000000000
```
