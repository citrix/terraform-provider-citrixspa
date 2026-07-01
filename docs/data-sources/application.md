---
page_title: "spa_application Data Source - spa"
description: |-
  Fetches a single SPA application by ID or name.
---

# spa_application (Data Source)

Fetches a single SPA application by ID or name. When looking up by name, the name must match exactly one application.

For more details on the underlying API, see the [Applications API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-applications).

## Example Usage

```terraform
# Look up by ID
data "spa_application" "by_id" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Look up by name
data "spa_application" "by_name" {
  name = "My Web Application"
}
```

## Schema

### Optional

- `id` (String) Application identifier. Either `id` or `name` must be specified.
- `name` (String) Name of the application. Either `id` or `name` must be specified.

### Read-Only

- `type` (String) Type of application (`web`, `saas`, `ztna`).
- `description` (String) Description of the application.
- `url` (String) Application URL.
- `category` (String) Category of the application.
- `hidden` (Boolean) Whether the application is hidden.
- `agentless_access` (Boolean) Whether agentless access is enabled.
- `mobile_security` (Boolean) Whether mobile security is enabled.
- `sbs_only_launch` (Boolean) Whether SBS-only launch is enabled.
- `using_template` (Boolean) Whether the application uses a template.
- `template_name` (String) Template name.
- `icon` (String) Base64-encoded icon data.
- `icon_url` (String) Application icon URL.
- `related_urls` (Set of String) Related URLs.
- `keywords` (Set of String) Keywords associated with the application.
- `locations` (Attributes List) Resource locations. Each entry has `name` (String) and `uuid` (String).
- `policies` (Attributes List) Policies. Each entry has `type` (String, values: `capability`, `patterns`) and `data` (Map of String).
- `destination` (Attributes List) Destinations for ZTNA applications. Each entry has `destination` (String — a hostname for `SUBTYPE_HOSTNAME`; a single IP or CIDR range for `SUBTYPE_IP_AND_CIDR`; or an IP range for `SUBTYPE_IP_RANGE`), `port`, `protocol`, and `subtype` (all String).
- `custom_properties` (Map of String) Custom properties.
- `customer_domain_fields` (Map of String) Customer domain fields.
- `sso` (Attributes) SSO configuration. (see [below for nested schema](#nestedatt--sso))
- `state` (String) Application state (`incomplete` or `complete`).
- `policy_count` (String) Number of policies associated with the application.
- `created_time` (String) Time the application was created (ISO 8601, e.g. `2026-04-08T14:37:24Z`).

<a id="nestedatt--sso"></a>
### Nested Schema for `sso`

Read-Only:

- `type` (String) SSO type: `saml`, `kerberos`, `basic`, `form`, or `nosso`.
- `saml_type` (String) SAML role.
- `sp_initiated_only` (Boolean) SP-initiated only.
- `assertion_url` (String) SAML ACS URL.
- `audience` (String) SAML audience.
- `relay_state` (String) SAML relay state.
- `sign_assertion` (String) SAML signature scope.
- `name_id_source` (String) SAML NameID source.
- `name_id_format` (String) SAML NameID format.
- `custom_attributes` (Attributes List) SAML custom attributes. Each entry has `format` (String), `name` (String), `value` (String), and `prefix_expr` (Boolean).
- `saml_sso_login_url` (String) SAML SSO login URL (server-computed).
- `saml_cert_issuer_name` (String) SAML cert issuer name (server-computed).
- `customer` (String) Customer ID (server-computed).
- `action_url` (String) Form SSO action URL.
- `logonform_url` (String) Form SSO logon form URL.
- `username_field` (String) Form SSO username field.
- `password_field` (String) Form SSO password field.
- `attribute` (String) Form SSO attribute.
- `username_format` (String) Username format.
- `user_realm` (String) Kerberos user realm.
