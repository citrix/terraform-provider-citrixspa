---
page_title: "Upgrading to 1.1.0: spa_* to citrixspa_*"
subcategory: "Upgrade Guides"
description: |-
  Migrate existing configurations and Terraform state after the provider resource and data source types were renamed from spa_* to citrixspa_* in v1.1.0.
---

# Upgrading to 1.1.0

In v1.1.0 the provider's resource and data source types were renamed from
`spa_*` to `citrixspa_*`, and the provider local name changed from `spa` to
`citrixspa`, to match the published registry name `citrix/citrixspa`.

Upgrading requires two changes that must be done together: your **configuration
files** and your existing **Terraform state**. If you rename the types only in
your `.tf` files, Terraform will plan to destroy the old `spa_*` resources and
create new `citrixspa_*` ones. Rewriting the resource types (and their dependency
references) in state — Step 3 — is what keeps your existing resources in place.

## Step 1 — Update your configuration

In every `.tf` file, rename the provider local name and each resource /
data-source type:

- `provider "spa"` → `provider "citrixspa"`
- `required_providers { spa = { source = "citrix/citrixspa" } }` →
  `required_providers { citrixspa = { source = "citrix/citrixspa" } }`
- `resource "spa_application" "x"` → `resource "citrixspa_application" "x"`, and
  likewise for every other `spa_*` resource and `data "spa_*"` data source.

## Step 2 — Reinitialize with the new provider

Pull in the renamed provider version so Terraform uses the new types:

```sh
terraform init -upgrade
```

## Step 3 — Rewrite the resource types (and dependencies) in your Terraform state

Export the state, change every resource `type` **and** every `dependencies`
address from `spa_*` to `citrixspa_*`, and write it back. Use `-force` on the
push, because you are intentionally replacing the state at the same serial:

```sh
terraform state pull > state.json
#   apply exactly ONE of the options below to state.json
terraform state push -force state.json
```

Both the resource `type` fields and the `dependencies` arrays reference the old
prefix, so the rewrite must update **both** — otherwise Terraform keeps a stale
`spa_*` dependency in state until the next apply.

## Step 4 — Verify

Confirm no `spa_*` remains in the state, then plan:

```sh
grep '"spa_' state.json      # expect NO output
terraform plan               # expect: No changes. Your infrastructure matches the configuration.
```

(Data sources need only the Step 1 config rename; their state is refreshed on
every plan.)

## Choose one way to do the Step 3 rewrite

**Python is the recommended option** — it edits only the `type` and
`dependencies` fields, so it cannot accidentally change attribute values.

### Python (recommended)

```sh
python3 - state.json <<'PY'
import json, sys
path = sys.argv[1]
data = json.load(open(path))
def fix(s):
    return "citrixspa_" + s[len("spa_"):] if s.startswith("spa_") else s
for r in data["resources"]:
    r["type"] = fix(r["type"])
    for inst in r.get("instances", []):
        if inst.get("dependencies"):
            inst["dependencies"] = [fix(d) for d in inst["dependencies"]]
json.dump(data, open(path, "w"), indent=2)
PY
```

### jq

```sh
jq '
  def fix: if type == "string" and startswith("spa_")
           then "citrixspa_" + ltrimstr("spa_") else . end;
  (.resources[].type) |= fix
  | (.resources[].instances[]?.dependencies[]?) |= fix
' state.json > state.tmp && mv state.tmp state.json
```

### sed

```sh
sed -E -i \
  -e 's/("type"[[:space:]]*:[[:space:]]*")spa_/\1citrixspa_/g' \
  -e 's/^([[:space:]]*")spa_([A-Za-z_]+\.)/\1citrixspa_\2/' \
  state.json
```

### Manually (text editor)

Find every `"spa_` and replace it with `"citrixspa_`. These appear in the `type`
fields and in `dependencies` addresses; review each match before replacing —
attribute values almost never start with `spa_`. Save the file, then run the
`terraform state push -force` from Step 3.