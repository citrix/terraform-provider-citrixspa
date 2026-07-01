---
page_title: "spa_access_policy Data Source - spa"
description: |-
  Fetches a single SPA access policy by ID or name.
---

# spa_access_policy (Data Source)

Fetches a single SPA access policy by ID or name. When looking up by name, the name must match exactly one policy.

For more details on the underlying API, see the [Access Policies API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-access-policies).

## Example Usage

```terraform
# Look up by ID
data "spa_access_policy" "by_id" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Look up by name
data "spa_access_policy" "by_name" {
  name = "Allow Developers"
}
```

## Schema

### Optional

- `id` (String) Access policy identifier. Either `id` or `name` must be specified.
- `name` (String) Name of the access policy. Either `id` or `name` must be specified.

### Read-Only

- `description` (String) Description of the access policy.
- `active` (Boolean) Whether the access policy is active.
- `priority` (Number) Priority of the access policy.
- `apps` (Set of String) Set of application IDs associated with the access policy.
- `modified` (String) Time the access policy was last modified (ISO 8601, e.g. `2026-05-11T09:49:40Z`).
- `access_rules` (Attributes List) Access rules for the policy. (see [below for nested schema](#nestedatt--access_rules))

<a id="nestedatt--access_rules"></a>
### Nested Schema for `access_rules`

Read-Only:

- `id` (String) Access rule ID.
- `name` (String) Access rule name.
- `description` (String) Access rule description.
- `priority` (Number) Access rule priority.
- `active` (Boolean) Whether the access rule is active.
- `access` (String) Access type for HTTP apps (web/SaaS). Values: `ACCESS_ALLOW`, `ACCESS_DENY`.
- `access_native` (String) Access type for TCP apps (ZTNA). Values: `ACCESS_ALLOW`, `ACCESS_DENY`.
- `advanced_settings` (Attributes) Advanced settings for the access rule. (see [below for nested schema](#nestedatt--access_rules--advanced_settings))
- `conditions` (Attributes List) Conditions for the access rule. (see [below for nested schema](#nestedatt--access_rules--conditions))
- `restrictions` (Attributes) Restrictions for the access rule. (see [below for nested schema](#nestedatt--access_rules--restrictions))
- `rules` (Attributes List) Matching rules within the access rule. (see [below for nested schema](#nestedatt--access_rules--rules))

<a id="nestedatt--access_rules--advanced_settings"></a>
### Nested Schema for `access_rules.advanced_settings`

Read-Only:

- `domain_overrides` (Attributes List) Domain override settings. (see [below for nested schema](#nestedatt--access_rules--advanced_settings--domain_overrides))

<a id="nestedatt--access_rules--advanced_settings--domain_overrides"></a>
### Nested Schema for `access_rules.advanced_settings.domain_overrides`

Read-Only:

- `fqdn` (String) Fully qualified domain name.
- `location_ids` (List of String) List of location IDs.
- `type` (String) Domain override type.

<a id="nestedatt--access_rules--conditions"></a>
### Nested Schema for `access_rules.conditions`

Read-Only:

- `platform_filter` (String) Platform filter. Values: `PLATFORM_FILTER_MOBILE`, `PLATFORM_FILTER_PC`, `PLATFORM_FILTER_ANY`.
- `user_and_groups` (Map of String) User and groups configuration.

<a id="nestedatt--access_rules--restrictions"></a>
### Nested Schema for `access_rules.restrictions`

Read-Only:

- `redirect_sbs` (Boolean) Whether to redirect to Secure Browser Service.
- `enhanced_security_settings` (Map of String) Enhanced security settings. Supported keys and their accepted values:
  - `_browserV1`: Only accepted value: `"embeddedBrowser"`.
  - `clipboardV1`: Values: `"enabled"` (default), `"disabled"`.
  - `downloadV1`: Values: `"enabled"` (default), `"disabled"`.
  - `printingV1`: Values: `"enabled"` (default), `"disabled"`.
  - `watermarkV1`: Values: `"enabled"`, `"disabled"` (default).
  - `keyLoggingV1`: Values: `"enabled"`, `"disabled"`.
  - `screenCaptureV1`: Values: `"enabled"`, `"disabled"`.
  - `proxyTrafficV1`: Values: `"direct"`, `"secureBrowse"`.
  - `uploadV1`: Values: `"enabled"`, `"disabled"`.

<a id="nestedatt--access_rules--rules"></a>
### Nested Schema for `access_rules.rules`

Read-Only:

- `type` (String) Rule type. Values: `TYPE_TAG`, `TYPE_USERGROUP`, `TYPE_PLATFORM`, `TYPE_MACHINEGROUP`, `TYPE_MULTIURLDOMAIN`.
- `operator` (String) Rule operator. Values: `OPERATOR_EQ`, `OPERATOR_IN`, etc. When `type` is `TYPE_MULTIURLDOMAIN`, only `OPERATOR_IN` or `OPERATOR_NOT` are present.
- `tag_source` (String) Source of data retrieval for `TYPE_TAG` rules. Empty string when not applicable. Values: `""`, `NLS`, `CAS`, `EPA`, `ITM`, `ThirdPartyDevicePosture`, `CONTEXTUAL`.
- `tag_key` (String) Tag key for `TYPE_TAG` rules (e.g., `location-geo-country-isocode`). Empty string when not applicable.
- `values` (List of String) Rule values. Interpretation depends on `type`:
  - `TYPE_USERGROUP`: User or group identifiers in SID or OID format (e.g., `"SID:/..."`, `"OID:/..."`). The special value `"Everyone"` matches all users.
  - `TYPE_TAG`: Values corresponding to the chosen `tag_key` (e.g., ISO country codes).
  - `TYPE_MULTIURLDOMAIN`: Domain names.
- `metadata` (Map of String) Key-value pairs providing display labels for rule values.
