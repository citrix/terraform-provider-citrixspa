---
page_title: "spa_security_group Resource - spa"
description: |-
  Resource for creating and managing SPA security groups.
---

# spa_security_group (Resource)

Resource for creating and managing SPA security groups. A security group defines clipboard isolation policies that control the flow of data (copy and paste) between virtual sessions and the user's local environment.

Each security group defines two independent clipboard policies:
- **`system`** — governs clipboard interactions between the session and the local operating system.
- **`unpublished_app`** — governs clipboard interactions with unpublished (non-managed) applications running inside the session.

~> **Note** An application can belong to at most one security group at a time. Only fully provisioned (`complete` state) web or SaaS applications are eligible for assignment.

### Application Dependencies

Security groups reference applications by ID via the `app_ids` attribute. When managing both in the same configuration, reference the application resources directly so Terraform automatically creates them in the correct order:

```terraform
resource "spa_security_group" "example" {
  name = "Restricted Group"

  app_ids = [
    spa_application.web_app.id,
    spa_application.saas_app.id,
  ]

  # ...
}
```

## Example Usage

```terraform
resource "spa_security_group" "restricted_clipboard" {
  name = "Restricted Clipboard Group"

  app_ids = [
    spa_application.my_web_application.id,
    spa_application.saas_app.id,
  ]

  system = {
    data_in  = "enabled"
    data_out = "disabled"
  }

  unpublished_app = {
    data_in  = "disabled"
    data_out = "disabled"
  }
}
```

## Schema

### Required

- `name` (String) Name of the security group.
- `app_ids` (Set of String) Set of application IDs to associate with this security group. Only `web` and `saas` applications in `complete` state are eligible. Each application can belong to at most one security group. Cannot be an empty array.
- `system` (Attributes) Clipboard isolation policy for interactions between the virtual session and the local operating system. (see [below for nested schema](#nestedatt--system))
- `unpublished_app` (Attributes) Clipboard isolation policy for interactions with unpublished (non-managed) applications running in the session. (see [below for nested schema](#nestedatt--unpublished_app))

### Read-Only

- `id` (String) GUID identifier of the security group.
- `modified` (Number) Last modification timestamp of the security group (Unix epoch seconds).

<a id="nestedatt--system"></a>
### Nested Schema for `system`

Required:

- `data_in` (String) Whether data can be pasted into the session from the system clipboard. Valid values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to the system clipboard. Valid values: `"enabled"`, `"disabled"`.

<a id="nestedatt--unpublished_app"></a>
### Nested Schema for `unpublished_app`

Required:

- `data_in` (String) Whether data can be pasted into the session from unpublished applications. Valid values: `"enabled"`, `"disabled"`.
- `data_out` (String) Whether data can be copied out of the session to unpublished applications. Valid values: `"enabled"`, `"disabled"`.

## Import

Import is supported using the security group ID:

```shell
terraform import spa_security_group.restricted_clipboard 00000000-0000-0000-0000-000000000000
```
