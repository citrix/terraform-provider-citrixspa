---
page_title: "citrixspa_applications Data Source - citrixspa"
description: |-
  Fetches a list of all SPA applications.
---

# citrixspa_applications (Data Source)

Fetches a paginated list of all SPA applications.

For more details on the underlying API, see the [Applications API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-applications).

-> **Note** For full details on each application (including SSO and policies), use the singular `citrixspa_application` data source with a specific `id`. In the list response, `sso` is returned as type only.

## Example Usage

```terraform
# Fetch all applications
data "citrixspa_applications" "all" {}

# Fetch with pagination
data "citrixspa_applications" "page" {
  offset = 0
  limit  = 50
}
```

## Schema

### Optional

- `offset` (Number) Offset for pagination. Defaults to `0`.
- `limit` (Number) Maximum number of results to return. Defaults to all results when omitted.

### Read-Only

- `applications` (Attributes List) List of applications. (see [below for nested schema](#nestedatt--applications))

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Read-Only:

- `id` (String) Application identifier.
- `name` (String) Name of the application.
- `type` (String) Type of application.
- `description` (String) Description of the application.
- `url` (String) Application URL.
- `category` (String) Category of the application.
- `hidden` (Boolean) Whether the application is hidden.
- `agentless_access` (Boolean) Whether agentless access is enabled.
- `mobile_security` (Boolean) Whether mobile security is enabled.
- `sbs_only_launch` (Boolean) Whether SBS-only launch is enabled.
- `using_template` (Boolean) Whether the application was provisioned from a built-in SPA catalog template (backend); unrelated to the provider's `templates/` directory.
- `template_name` (String) Name of the built-in catalog template (backend) from which the application was provisioned.
- `icon` (String) Base64-encoded icon data.
- `icon_url` (String) Application icon URL.
- `related_urls` (Set of String) Related URLs.
- `keywords` (Set of String) Keywords associated with the application.
- `locations` (Attributes List) Null for list response.
- `policies` (Attributes List) Null for list response.
- `destination` (Attributes List) Destinations for ZTNA applications. Each entry has `destination` (String — a hostname for `SUBTYPE_HOSTNAME`; a single IP or CIDR range for `SUBTYPE_IP_AND_CIDR`; or an IP range for `SUBTYPE_IP_RANGE`), `port`, `protocol`, and `subtype` (all String).
- `custom_properties` (Map of String) Custom properties.
- `customer_domain_fields` (Map of String) Customer domain fields.
- `sso` (String) SSO configuration type for list response. One of: `basic`, `saml`, `kerberos`, `form`, `nosso`.
- `state` (String) Application state.
- `policy_count` (String) Number of policies.
- `created_time` (String) Timestamp of when the application was created (e.g., `"2023-01-19T10:00:22Z"`).
