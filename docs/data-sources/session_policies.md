---
page_title: "spa_session_policies Data Source - spa"
description: |-
  Fetches a list of SPA session policies.
---

# spa_session_policies (Data Source)

Fetches a paginated list of SPA session policies, including their rules, conditions, and actions.

## Example Usage

```terraform
# Fetch all session policies
data "spa_session_policies" "all" {}

# Fetch with name filter and pagination
data "spa_session_policies" "filtered" {
  name    = "Route external"
  orderby = "name"
  offset  = 0
  limit   = 50
}

output "session_policy_ids" {
  value = data.spa_session_policies.all.session_policies[*].id
}
```

## Schema

### Optional

- `offset` (Number) Offset for pagination. Defaults to `0`.
- `limit` (Number) Maximum number of results to return. Defaults to all results when omitted.
- `name` (String) Filter policies by name (exact match).
- `orderby` (String) Order results by a field. Supported values: `name`, `priority`.

### Read-Only

- `session_policies` (Attributes List) List of session policies. (see [below for nested schema](#nestedatt--session_policies))

<a id="nestedatt--session_policies"></a>
### Nested Schema for `session_policies`

Read-Only:

- `id` (String) Session policy ID.
- `name` (String) Session policy name.
- `description` (String) Session policy description.
- `active` (Boolean) Whether the policy is active.
- `priority` (Number) Policy priority.
- `generic_rules` (Attributes List) Rules within the session policy. (see [below for nested schema](#nestedatt--session_policies--generic_rules))

<a id="nestedatt--session_policies--generic_rules"></a>
### Nested Schema for `session_policies.generic_rules`

Read-Only:

- `id` (String) Rule identifier. May be null if no identifier has been assigned to the rule.
- `name` (String) Rule name.
- `description` (String) Rule description.
- `priority` (Number) Rule priority.
- `active` (Boolean) Whether the rule is active.
- `actions` (Attributes) Actions applied when this rule matches. (see [below for nested schema](#nestedatt--session_policies--generic_rules--actions))
- `condition` (Attributes List) Conditions for this rule. (see [below for nested schema](#nestedatt--session_policies--generic_rules--condition))

<a id="nestedatt--session_policies--generic_rules--actions"></a>
### Nested Schema for `session_policies.generic_rules.actions`

Read-Only:

- `routing` (String) Routing direction (`"default"` or `"external"`).
- `disable_security_groups` (String) Whether security groups are disabled (`"true"` or `"false"`).
- `local_lan_access` (String) Local LAN access setting (`"enabled"` or `"disabled"`).

<a id="nestedatt--session_policies--generic_rules--condition"></a>
### Nested Schema for `session_policies.generic_rules.condition`

Read-Only:

- `type` (String) Condition type (e.g., `TYPE_USERGROUP`, `TYPE_PLATFORM`).
- `operator` (String) Condition operator (e.g., `OPERATOR_IN`).
- `tag_source` (String) Tag source for `TYPE_TAG` conditions.
- `tag_key` (String) Tag key for `TYPE_TAG` conditions.
- `values` (List of String) Condition values.
- `metadata` (Map of String) Optional metadata as key-value pairs.
