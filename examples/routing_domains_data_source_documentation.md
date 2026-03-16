# Routing Domains Data Source

This data source allows you to retrieve a list of all routing domains from the SPA (Secure Private Access) API.

## API Endpoint

The data source uses the following API endpoint:

```
GET https://api.cloud.com/accessSecurity/routingDomains
```

## Example Usage

```hcl
# Get all routing domains
data "spa_routing_domains" "all" {
}

# Get routing domains with pagination
data "spa_routing_domains" "paginated" {
  offset = 0
  limit  = 50
}

# Output the routing domains
output "routing_domains" {
  value = data.spa_routing_domains.all.routing_domains
}
```

## Arguments Reference

The following arguments are supported:

- `offset` - (Optional) The offset for pagination. Default is 0.
- `limit` - (Optional) The limit for pagination. If not specified, all routing domains will be returned.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `routing_domains` - A list of routing domain objects. Each routing domain has the following attributes:

  - `id` - The routing domain identifier
  - `fqdn` - The fully qualified domain name
  - `type` - The type of routing entry
  - `app_type` - The type of app binding to this routing entry
  - `comment` - Admin description for the routing entry
  - `flag` - Whether routing entry is enabled or disabled
  - `error` - Any error associated with this routing entry
  - `ip` - Whether the secure access app has IP-based configuration
  - `location_ids` - List of resource location UUIDs
  - `status` - Status information for the routing domain (map of key-value pairs)

- `total` - The total number of routing domains available
- `count` - The number of routing domains returned in this response
- `offset` - The offset used for pagination
- `limit` - The limit used for pagination

## Notes

- This data source complements the existing `spa_routing_domain` data source, which retrieves a single routing domain by FQDN.
- The API endpoint requires proper authentication via the SPA provider configuration.
- Use pagination parameters for large datasets to improve performance.
