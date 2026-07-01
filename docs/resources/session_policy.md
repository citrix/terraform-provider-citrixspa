---
page_title: "spa_session_policy Resource - spa"
description: |-
  Resource for creating and managing SPA session policies.
---

# spa_session_policy (Resource)

Resource for creating and managing SPA session policies. Session policies apply routing and security behaviour at the session level across all applications, unlike access policies which are per-application. A session policy contains one or more rules, each with conditions and actions.

~> **Note** Not all action fields are available in every tenant configuration. Only include action fields applicable to your deployment.

## Example Usage

```terraform
resource "spa_session_policy" "route_external" {
  name        = "Route external users"
  description = "Apply external routing for non-corporate devices"
  active      = true
  # priority is optional — omit to let the server auto-assign

  generic_rules = [
    {
      name        = "External platform rule"
      description = "Route PC clients externally"
      priority    = 1
      active      = true

      actions = {
        routing          = "external"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}
```

## Schema

### Required

- `name` (String) Name of the session policy.
- `active` (Boolean) Whether the session policy is active.
- `generic_rules` (Attributes List) One or more rules within the session policy. (see [below for nested schema](#nestedatt--generic_rules))

### Optional

- `description` (String) Description of the session policy. Defaults to empty string when not set.
- `priority` (Number) Priority of the session policy. Optional — if omitted, the server assigns a value automatically and Terraform tracks it in state for subsequent operations.

### Read-Only

- `id` (String) GUID identifier of the session policy.

<a id="nestedatt--generic_rules"></a>
### Nested Schema for `generic_rules`

Required:

- `active` (Boolean) Whether the rule is active.
- `priority` (Number) Rule priority within the policy (lower number = higher priority).

Optional:

- `id` (String) Rule identifier. Typically omitted on creation; may be assigned by the API.
- `name` (String) Rule name.
- `description` (String) Rule description.
- `actions` (Attributes) Actions applied when this rule matches. (see [below for nested schema](#nestedatt--generic_rules--actions))
- `condition` (Attributes List) Conditions that must all match for the rule to fire. (see [below for nested schema](#nestedatt--generic_rules--condition))

<a id="nestedatt--generic_rules--actions"></a>
### Nested Schema for `generic_rules.actions`

~> Not all action fields are available in every tenant configuration. Only include fields applicable to your deployment.

Optional:

- `routing` (String) Routing direction. Valid values: `"default"`, `"external"`.
- `disable_security_groups` (String) Whether to disable security groups. Valid values: `"true"`, `"false"`.
- `local_lan_access` (String) Local LAN access setting. Valid values: `"enabled"`, `"disabled"`.

<a id="nestedatt--generic_rules--condition"></a>
### Nested Schema for `generic_rules.condition`

Required:

- `type` (String) Condition type. Valid values: `TYPE_USERGROUP`, `TYPE_PLATFORM`, `TYPE_TAG`, `TYPE_MACHINEGROUP`, `TYPE_MULTIURLDOMAIN`.
- `operator` (String) Condition operator. Valid values: `OPERATOR_EQ`, `OPERATOR_IN`, `OPERATOR_CONTAINS`, `OPERATOR_LTE`, `OPERATOR_GTE`, `OPERATOR_NOT`, `OPERATOR_RANGE`.
- `values` (List of String) Condition values (e.g., SID/OID strings for `TYPE_USERGROUP`, platform names for `TYPE_PLATFORM`).

Optional:

- `tag_source` (String) Tag source for `TYPE_TAG` conditions. Valid values: `NLS`, `CAS`, `EPA`, `ITM`, `ThirdPartyDevicePosture`. Must not be set for `TYPE_MULTIURLDOMAIN`.
- `tag_key` (String) Tag key for `TYPE_TAG` conditions. Must not be set for `TYPE_MULTIURLDOMAIN`.
- `metadata` (Map of String) Optional metadata as key-value pairs (e.g., display labels for user/group entries).

## Import

Import is supported using the session policy ID:

```shell
terraform import spa_session_policy.route_external 00000000-0000-0000-0000-000000000000
```
