# E2E Tests

End-to-end tests for the `citrixspa` Terraform provider. They build the provider
and exercise it against live Secure Private Access tenants through the real
Terraform CLI, covering the full resource lifecycle plus the discovery and
migration tooling.

> ⚠️ These tests create and destroy real resources. Run them only against
> dedicated, disposable test tenants. Local credential and state files
> (`terraform.tfvars`, Terraform state, `.terraformrc`) are git-ignored.

## Prerequisites

- Go 1.24+
- Terraform >= 1.0
- PowerShell 7+ (`pwsh`) — for the `tooling` and `migrate` scenarios
- Service-principal credentials for one or two dedicated test tenants

## Tenants

| Tenant | Role |
|--------|------|
| **Customer A** | Holds the standard reference ("golden") configuration in `customer-a/`. |
| **Customer B** | Kept empty; used only as the target of a migration. |

Only Customer A is required for `golden`, `tooling`, and `tests`. The `migrate`
scenario additionally needs Customer B.

## Scenarios

| Scenario | Script | What it checks |
|----------|--------|----------------|
| `golden` | `run-golden.sh` | Applies `customer-a/` and asserts there is no drift from the reference config. |
| `tooling` | `tooling/run-spa-manager-roundtrip.sh` | The discovery tool (`spa_manager.ps1`) reproduces the tenant's live state. |
| `migrate` | `tooling/run-migrate-e2e.sh` | Migrates Customer A → Customer B (with a guarded pre-clean of B). |
| `tests` | `run-tests.sh` | Go unit tests + acceptance tests (`TF_ACC`) against Customer A. |

## Run it manually

1. Copy the example vars file and fill in your test tenant(s):

   ```bash
   cd e2e
   cp terraform.tfvars.example terraform.tfvars
   # edit terraform.tfvars — Customer A keys always; Customer B (*_b) only for migrate
   ```

2. Run a scenario (each script builds/installs the provider and cleans up after):

   ```bash
   DESTROY_AFTER=1 ./run-golden.sh
   DESTROY_AFTER=1 ./tooling/run-spa-manager-roundtrip.sh
   CONFIRM_WIPE_TENANT_B=yes DESTROY_AFTER=1 ./tooling/run-migrate-e2e.sh
   ./run-tests.sh
   ```

Credentials can also be provided via environment variables instead of the vars
file: `CITRIX_CUSTOMER_ID` / `CITRIX_CLIENT_ID` / `CITRIX_CLIENT_SECRET` (Customer A)
and the `*_B` variants (Customer B).

### Environment options

| Variable | Effect |
|----------|--------|
| `DESTROY_AFTER=1` | Tear down created resources after the run (recommended). |
| `KEEP_RESOURCES=1` | `migrate`: skip teardown (debugging). |
| `CONFIRM_WIPE_TENANT_B=yes` | `migrate`: allow wiping Customer B when it holds more than the safety threshold of resources. |
| `WIPE_ABORT_THRESHOLD` | `migrate`: safety threshold before a wipe is refused (default 5). |

## Layout

| Path | Purpose |
|------|---------|
| `customer-a/` | Reference ("golden") configuration for Customer A |
| `run-golden.sh` | Golden drift check |
| `run-tests.sh` | Unit + acceptance |
| `tooling/run-spa-manager-roundtrip.sh` | Discovery round-trip |
| `tooling/run-migrate-e2e.sh` | Two-tenant migration |
| `tooling/run-migrate-cleanup.sh` | Manual cleanup after a `KEEP_RESOURCES=1` migrate run |
| `cleanup.sh` | Best-effort API sweep by prefix (fallback) |

The reference configuration is an intentionally small starter set (a web app, a
SaaS app, routing domains, and an access policy); more cases will be added.
