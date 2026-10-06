---
page_title: "citrixspa_routing_domain Resource - citrixspa"
description: |-
  Resource for creating and managing SPA routing domains.
---

# citrixspa_routing_domain (Resource)

Resource for creating and managing SPA routing domains. A routing domain defines how traffic for a specific FQDN is routed.

For more details on the underlying API, see the [Application Domains API documentation](https://developer-docs.citrix.com/en-us/secure-private-access/access-security/handling-application-domains).

~> **Note 1** Each FQDN must be unique across the account. The API will reject the creation of a routing domain if another routing domain with the same `fqdn` already exists.

-> **Note 2** Before a routing domain can be deleted, the provider automatically disables it (sets `flag` to `"disabled"`), as the API does not allow deletion of enabled routing domains. This happens transparently and is not visible as a separate step in the plan.

~> **Note 3 — deleting an application also deletes matching routing domains.** When an application is deleted, Citrix Secure Private Access also deletes the routing domains whose `fqdn` matches one of that application's `url`, `related_urls` or `destination` values. This happens for **every** matching routing domain, including one you created independently with `citrixspa_routing_domain`, through the Citrix Cloud console, or through the public API — there is no ownership link between an application and a routing domain, so the match is made on the FQDN string alone.

The rule in detail:

- **Matching is by FQDN string.** The application's `url` and each of its `related_urls` are reduced to their hostname before the match, so `https://example.com/path` matches a routing domain with `fqdn = "example.com"`. ZTNA `destination` values are matched as written, which preserves an IP address, a CIDR range or an IP range exactly as configured.
- **A shared FQDN is retained.** If any other application in the account still references the same FQDN, the routing domain is kept.
- **Wildcards are literal.** A routing domain for `*.example.com` is matched only by the literal string `*.example.com`; it is not removed because `other.example.com` was deleted.
- **Only the delete path is affected.** Nothing removes a routing domain in the background; a routing domain whose FQDN never appears on a deleted application is never touched.

**Effect on your configuration.** If a `citrixspa_routing_domain` resource is removed this way, the next `terraform plan` refresh reports it as missing and the provider emits a warning explaining why. While the resource is still declared, the plan also proposes creating it again; if you are removing that resource block, or running `terraform destroy`, Terraform simply drops the missing object and proposes nothing. A `plan` does not rewrite the state file; the recorded state is corrected on the next `terraform apply`. `terraform state rm` is not required in either case. To resolve the drift:

1. If you no longer need the routing domain, remove its resource block from your configuration.
2. If you do still need it, the next `terraform apply` recreates it. Terraform treats your configuration as the desired state, so a routing domain that stays declared is always restored.

This is most easily overlooked when the application and the routing domain are managed in **different configurations or state files**, where deleting the application in one workspace introduces drift in another workspace that was not modified. See [Routing Domain Dependencies](application.md) on the `citrixspa_application` page for the create-side ordering requirement.

## Example Usage

```terraform
resource "citrixspa_routing_domain" "intranet_example_com" {
  fqdn = "intranet.example.com"
  type = "internal"

  app_type = "web"
  comment  = "Internal intranet portal"
  flag     = "enabled"

  location_ids = [
    "00000000-0000-0000-0000-000000000000"
  ]
}
```

## Schema

### Required

- `fqdn` (String) Fully qualified domain name. This also serves as the unique identifier for the routing domain. Max length: 255 characters.
- `type` (String) Type of routing entry. Valid values: `internal`, `external`, `conflicting`, `internal_bypass_proxy`, `internal_via_gateway`, `external_fixed_ip`.
- `app_type` (String) Type of application bound to this routing entry. Valid values: `ztna`, `web`, `saas`.
- `flag` (String) Whether the routing entry is `enabled` or `disabled`.
- `ip` (Boolean) Whether the secure access app has IP-based configuration. Required for `ztna` app types.
- `location_ids` (List of String) List of resource location UUIDs associated with this routing domain.

### Optional

- `comment` (String) Admin description for the routing entry. Max length: 64 characters.

### Read-Only

- `error` (String) Any error associated with this routing entry, as reported by the API. Defaults to `"none"` when no error is present.

## Import

Import is supported using the FQDN:

```shell
terraform import citrixspa_routing_domain.internal intranet.example.com
```
