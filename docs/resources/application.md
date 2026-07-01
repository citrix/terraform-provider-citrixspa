---
page_title: "spa_application Resource - spa"
description: |-
  Resource for creating and managing SPA applications.
---

# spa_application (Resource)

Resource for creating and managing SPA applications. Supports web, SaaS, and ZTNA application types.

For more details on the underlying API, see the [Applications API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-applications).

~> **Note** Application names should be unique within an account.

Applications can be created in an `incomplete` state and later transitioned to `complete` by setting the `state` attribute to `"complete"`.

### Routing Domain Dependencies

Applications rely on routing domains derived from their `url` and `related_urls`. When managing both resources in the same configuration, you must declare explicit dependencies so that routing domains are created before the application:

```terraform
resource "spa_application" "my_web_application" {
  name = "My Web Application"
  type = "web"
  url  = "https://example.com"
  related_urls   = ["api.example.com"]
  # ...

  depends_on = [
    spa_routing_domain.example_com,
    spa_routing_domain.api_example_com,
  ]
}
```

-> **Note** When using the migration script (`spa_manager.ps1`), dependency ordering is handled automatically.

### Field Requirements by Application Type

| Field | `web` | `saas` | `ztna` |
|---|---|---|---|
| `url` | Required | Required | Not used |
| `related_urls` | Required | Required | Not used |
| `destination` | Not used | Not used | Required |

## Example Usage

### Web Application

```terraform
resource "spa_application" "my_web_application" {
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
    spa_routing_domain.example_com,
    spa_routing_domain.wildcard_example_com
  ]
}
```

### ZTNA Application

```terraform
resource "spa_application" "ztna_app" {
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
    spa_routing_domain.database_internal_com,
    spa_routing_domain.routing_domain_10_0_1_0_24
  ]
}
```

### SaaS Application

```terraform
resource "spa_application" "saas_app" {
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

  sso = { type = "nosso" }

  locations = [
    {
      name = "Resource Location 1"
      uuid = "00000000-0000-0000-0000-000000000000"
    }
  ]

  depends_on = [
    spa_routing_domain.office_com,
    spa_routing_domain.wildcard_office_com,
    spa_routing_domain.wildcard_outlook_office_com,
    spa_routing_domain.wildcard_teams_microsoft_com
  ]
}
```

### SAML SSO Configuration

```terraform
resource "spa_application" "saml_app" {
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
- `using_template` (Boolean) Whether the application uses a template. Required for `web` and `saas` applications.
- `template_name` (String) Name of the template to use.
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
terraform import spa_application.web_app 00000000-0000-0000-0000-000000000000
```
