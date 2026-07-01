---
page_title: "spa_security_group Data Source - spa"
description: |-
  Fetches a single SPA security group by ID.
---

# spa_security_group (Data Source)

Fetches a single SPA security group by its ID.

Security groups define clipboard isolation policies that control the flow of data (copy and paste) between virtual sessions and the user's local environment. Applications assigned to a security group inherit its clipboard policy.

## Example Usage

```terraform
data "spa_security_group" "example" {
  id = "00000000-0000-0000-0000-000000000000"
}
```

## Schema

### Required

- `id` (String) Security group identifier.

### Read-Only

- `name` (String) Name of the security group.
- `app_ids` (Set of String) Set of application IDs associated with the security group.
- `system` (Attributes) Clipboard isolation policy for interactions between the virtual session and the local operating system. (see [below for nested schema](#nestedatt--system))
- `unpublished_app` (Attributes) Clipboard isolation policy for interactions with unpublished (non-managed) applications running in the session. (see [below for nested schema](#nestedatt--unpublished_app))
- `modified` (Number) Last modification timestamp of the security group (Unix epoch seconds).

<a id="nestedatt--system"></a>
### Nested Schema for `system`

Read-Only:

- `data_in` (String) Whether data can be pasted into the session from the system clipboard. Values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to the system clipboard. Values: `"enabled"`, `"disabled"`.

<a id="nestedatt--unpublished_app"></a>
### Nested Schema for `unpublished_app`

Read-Only:

- `data_in` (String) Whether data can be pasted into the session from unpublished applications. Values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to unpublished applications. Values: `"enabled"`, `"disabled"`.
