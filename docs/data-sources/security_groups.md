---
page_title: "spa_security_groups Data Source - spa"
description: |-
  Fetches a list of all SPA security groups.
---

# spa_security_groups (Data Source)

Fetches a list of all SPA security groups, including their clipboard isolation policies and associated applications.

## Example Usage

```terraform
data "spa_security_groups" "all" {}
```

## Schema

### Read-Only

- `security_groups` (Attributes List) List of security groups. (see [below for nested schema](#nestedatt--security_groups))

<a id="nestedatt--security_groups"></a>
### Nested Schema for `security_groups`

Read-Only:

- `id` (String) Security group identifier.
- `name` (String) Name of the security group.
- `app_ids` (Set of String) Set of application IDs associated with the security group.
- `system` (Attributes) Clipboard isolation policy for interactions between the virtual session and the local operating system. (see [below for nested schema](#nestedatt--security_groups--system))
- `unpublished_app` (Attributes) Clipboard isolation policy for interactions with unpublished (non-managed) applications running in the session. (see [below for nested schema](#nestedatt--security_groups--unpublished_app))
- `modified` (Number) Last modification timestamp of the security group (Unix epoch seconds).

<a id="nestedatt--security_groups--system"></a>
### Nested Schema for `security_groups.system`

Read-Only:

- `data_in` (String) Whether data can be pasted into the session from the system clipboard. Values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to the system clipboard. Values: `"enabled"`, `"disabled"`.

<a id="nestedatt--security_groups--unpublished_app"></a>
### Nested Schema for `security_groups.unpublished_app`

Read-Only:

- `data_in` (String) Whether data can be pasted into the session from unpublished applications. Values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to unpublished applications. Values: `"enabled"`, `"disabled"`.
