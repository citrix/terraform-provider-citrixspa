---
page_title: "citrixspa_routing_domains Data Source - citrixspa"
description: |-
  Fetches a list of all SPA routing domains.
---

# citrixspa_routing_domains (Data Source)

Fetches a paginated list of all SPA routing domains.

For more details on the underlying API, see the [Application Domains API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-application-domains).

## Example Usage

```terraform
# Fetch all routing domains
data "citrixspa_routing_domains" "all" {}

# Fetch with pagination: `limit` is an end index, so these are entries 0-49 and 50-99
data "citrixspa_routing_domains" "page" {
  offset = 0
  limit  = 50
}

data "citrixspa_routing_domains" "next_page" {
  offset = 50
  limit  = 100
}
```

## Schema

### Optional

- `offset` (Number) Index of the first routing domain to return. Defaults to `0`.
- `limit` (Number) Exclusive end index, matching the SPA API: the result is routing domains `[offset, limit)`. For example, `offset = 100, limit = 150` returns 50 entries, and a `limit` lower than `offset` is rejected by the API. This is **not** a result count. Defaults to all results from `offset` on when omitted.

### Read-Only

- `total` (Number) Total number of routing domains available.
- `routing_domains` (Attributes List) List of routing domains. (see [below for nested schema](#nestedatt--routing_domains))

<a id="nestedatt--routing_domains"></a>
### Nested Schema for `routing_domains`

Read-Only:

- `fqdn` (String) Fully qualified domain name.
- `type` (String) Type of routing entry.
- `app_type` (String) Type of application bound to this routing entry.
- `comment` (String) Admin description for the routing entry.
- `flag` (String) Whether the routing entry is `enabled` or `disabled`.
- `error` (String) Any error associated with this routing entry.
- `ip` (Boolean) Whether the secure access app has IP-based configuration.
- `location_ids` (List of String) List of resource location UUIDs.
