---
page_title: "spa_routing_domain Data Source - spa"
description: |-
  Fetches a single SPA routing domain by FQDN.
---

# spa_routing_domain (Data Source)

Fetches a single SPA routing domain by its FQDN.

For more details on the underlying API, see the [Application Domains API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-application-domains).

## Example Usage

```terraform
data "spa_routing_domain" "example" {
  fqdn = "intranet.example.com"
}
```

## Schema

### Required

- `fqdn` (String) Fully qualified domain name of the routing domain to look up.

### Read-Only

- `type` (String) Type of routing entry. Valid values: `internal`, `external`, `external_via_connector`, `internal_bypass_proxy`.
- `app_type` (String) Type of application bound to this routing entry (`ztna`, `web`, `saas`).
- `comment` (String) Admin description for the routing entry.
- `flag` (String) Whether the routing entry is `enabled` or `disabled`.
- `error` (String) Any error associated with this routing entry.
- `ip` (Boolean) Whether the secure access app has IP-based configuration.
- `location_ids` (List of String) List of resource location UUIDs.
