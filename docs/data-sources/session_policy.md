---
page_title: "spa_session_policy Data Source - spa"
description: |-
  Fetches a single SPA session policy by ID or name.
---

# spa_session_policy (Data Source)

Fetches a single SPA session policy by ID or name. When looking up by name, the name must match exactly one policy.

## Example Usage

```terraform
# Look up by ID
data "spa_session_policy" "by_id" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Look up by name
data "spa_session_policy" "by_name" {
  name = "Route external users"
}
```

## Schema

### Optional

- `id` (String) Session policy identifier. Either `id` or `name` must be specified.
- `name` (String) Name of the session policy. Either `id` or `name` must be specified.

### Read-Only

- `description` (String) Description of the session policy.
- `active` (Boolean) Whether the session policy is active.
- `priority` (Number) Priority of the session policy.
- `generic_rules` (Attributes List) Rules within the session policy. (see [below for nested schema](#nestedatt--generic_rules))

<a id="nestedatt--generic_rules"></a>
### Nested Schema for `generic_rules`

Read-Only:

- `id` (String) Rule identifier. May be null if no identifier has been assigned to the rule.
- `name` (String) Rule name.
- `description` (String) Rule description.
- `priority` (Number) Rule priority.
- `active` (Boolean) Whether the rule is active.
- `actions` (Attributes) Actions applied when this rule matches. (see [below for nested schema](#nestedatt--generic_rules--actions))
- `condition` (Attributes List) Conditions for this rule. (see [below for nested schema](#nestedatt--generic_rules--condition))

<a id="nestedatt--generic_rules--actions"></a>
### Nested Schema for `generic_rules.actions`

Read-Only:

- `routing` (String) Routing direction (`"default"` or `"external"`).
- `disable_security_groups` (String) Whether security groups are disabled (`"true"` or `"false"`).
- `local_lan_access` (String) Local LAN access setting (`"enabled"` or `"disabled"`).

<a id="nestedatt--generic_rules--condition"></a>
### Nested Schema for `generic_rules.condition`

Read-Only:

- `type` (String) Condition type (e.g., `TYPE_USERGROUP`, `TYPE_PLATFORM`).
- `operator` (String) Condition operator (e.g., `OPERATOR_IN`).
- `tag_source` (String) Tag source for `TYPE_TAG` conditions.
- `tag_key` (String) Tag key for `TYPE_TAG` conditions.
- `values` (List of String) Condition values.
- `metadata` (Map of String) Optional metadata as key-value pairs.
