# AGENTS.md — Citrix SPA Terraform Provider

Contributor reference for the Citrix Secure Private Access Terraform provider.

## Project Overview

| Field | Value |
|-------|-------|
| Name | Citrix SPA Terraform Provider |
| Language | Go 1.24 (toolchain 1.24.6) |
| Framework | Terraform Plugin Framework v1.16.1 |
| Binary | `terraform-provider-citrixspa` |
| Registry | `registry.terraform.io/citrix/citrixspa` |
| API Base | `https://api.cloud.com/accessSecurity` |
| Rate limiting | `golang.org/x/time/rate` |
| Encryption | `golang.org/x/crypto` (AES-256-GCM + PBKDF2) |
| UUID generation | `github.com/google/uuid` v1.6.0 |
| Build / Release | GoReleaser |

## Build & Test

```bash
make build       # Produces ./terraform-provider-citrixspa
make install     # Installs to ~/.terraform.d/plugins/registry.terraform.io/citrix/citrixspa/0.1.0/<os>_<arch>/
make test        # Unit tests
make testacc     # Acceptance tests
make fmt         # go fmt + terraform fmt
make lint        # Run golangci-lint
make docs        # Regenerate registry documentation (not used, docs are hand-written, not auto-generated)
make check       # fmt + lint + test (all checks)
make release     # goreleaser release --rm-dist
```

### Local Testing

```bash
./test-local.sh                      # Direct token auth
./test-local.sh --service-principal  # Service principal auth
./test-local.sh --sp --apply         # Apply changes
```

The script builds, installs locally, generates `.terraformrc`, and runs `terraform init/validate/plan`.

### Testing

- **Unit tests**: `make test` or `go test -v ./...` — no credentials needed
- **Acceptance tests**: `make testacc` — requires env vars: `CITRIX_CUSTOMER_ID`, `CITRIX_CLIENT_ID`, `CITRIX_CLIENT_SECRET` (`TF_ACC=1` is already set by the Makefile target)
- **Single test / filtered run example**:

```bash
TF_CLI_ARGS_apply="-parallelism=1" TF_CLI_ARGS_plan="-parallelism=1" TF_ACC=1 go test ./internal/provider/ -run 'TestAccSecurityGroup*'
```

- **Manual apply with debug logging example**:

```bash
TF_LOG=DEBUG TF_LOG_PATH=terraform-debug.log terraform apply -refresh=false -target='spa_application.app'
```

- Test configs: `test-local/` (direct token), `test-local-sp/` (service principal)

## Key Components

| Component | File | Responsibility |
|-----------|------|----------------|
| Provider | `internal/provider/provider.go` | Schema, configure, resource/datasource registration |
| API Client | `internal/provider/client.go` | `SPAClient` interface, `APIClient` with rate limiting and HTTP handling |
| Auth | `internal/provider/auth.go` | OAuth2 service principal flow, `AuthenticatedClient` with token refresh |
| Token Cache | `internal/provider/token_persistence.go` | AES-256-GCM encrypted disk cache (`~/.terraform.d/spa-cache/`) |
| Resource Listing | `resource-listing-tool/spa_manager.ps1` | PowerShell discovery & Terraform generation |

## Code Layout

```
internal/provider/
├── provider.go              Provider setup, schema, rate limiting, resource registration
├── auth.go                  OAuth2 client credentials authentication
├── client.go                SPAClient interface + APIClient implementation
├── token_persistence.go     Encrypted disk-based token cache (AES-256-GCM)
├── sso_type.go              Application SSO model and conversion helpers
├── resource_application.go  Application resource (web/saas/ztna)
├── resource_access_policy.go  Access policy resource
├── resource_routing_domain.go  Routing domain resource (keyed by FQDN)
├── resource_security_group.go  Security group resource
├── resource_certificate.go  Certificate resource
├── resource_browser_mode.go  Browser mode resource
├── resource_terminate_*.go  Access termination resources
├── data_source_*.go         Data source implementations
└── *_test.go                Co-located tests
```

## Documentation

Resource and data source documentation lives in `docs/` and is **hand-written** (not auto-generated). Each resource has a corresponding markdown file in `docs/resources/` and each data source in `docs/data-sources/`. These files are published to the Terraform Registry and serve as the primary user-facing reference.

## Code Conventions

- All provider source lives in `internal/provider/`
- One file per resource/data source: `resource_<name>.go`, `data_source_<name>.go`
- Tests co-located: `resource_<name>_test.go`, `data_source_<name>_test.go`
- Resource struct pattern:
  ```go
  type FooResource struct { client SPAClient }
  func NewFooResource() resource.Resource { return &FooResource{} }
  ```
- Computed fields: `Computed: true` + `UseStateForUnknown()` plan modifier
- Application SSO uses `SSOModel` struct with `ssoFromAPI()`/`ssoModelToObject()`/`ssoObjectToModel()` conversion helpers
- API list operations have `*Detailed()` variants for batch detail fetching
- All API calls go through `makeRequest()` which handles rate limiting, token refresh, headers, logging
- Routing domains use FQDN as identifier — URL-encoded in API paths via `encodeFQDN()`
- Auth logic in `auth.go`, token persistence in `token_persistence.go`

## Authentication

Two mutually exclusive methods — configure only one:

| Method | Config Fields | Environment Variables |
|--------|---------------|-----------------------|
| Direct token | `auth_token` | `CITRIX_AUTH_TOKEN` |
| Service principal | `client_id` + `client_secret` | `CITRIX_CLIENT_ID` + `CITRIX_CLIENT_SECRET` |

Always required: `customer_id` / `CITRIX_CUSTOMER_ID`

Service principal OAuth2 endpoint: `POST {token_url}/cctrustoauth2/{customerId}/tokens/clients`

### Schema Rules

- Mark API-assigned fields (IDs, timestamps, counts) `Computed: true`
- Use `UseStateForUnknown()` plan modifier on computed fields to suppress spurious diffs
- Use `Set` (not `List`) for unordered collections (e.g., `related_urls`) to avoid order diffs
- Sensitive fields (`auth_token`, `client_secret`): add `Sensitive: true`
- Always call through `SPAClient` interface, never instantiate `APIClient` directly in resource code
- List APIs and individual GET APIs may return different JSON shapes — use separate model types (e.g., `ApplicationListItem` vs `Application`)

## Resources (managed)

- `spa_application` — Web, SaaS, ZTNA applications
- `spa_access_policy` — Access policies with rules/conditions
- `spa_routing_domain` — Routing domains (keyed by FQDN)
- `spa_security_group` — Security groups
- `spa_certificate` — SSL certificates
- `spa_browser_mode` — Browser mode configuration
- `spa_terminate_machine_access` — Machine access termination
- `spa_terminate_user_access` — User access termination

## Data Sources (read-only)

- `spa_application` / `spa_applications`
- `spa_access_policy` / `spa_access_policies`
- `spa_routing_domain` / `spa_routing_domains`
- `spa_security_group` / `spa_security_groups`
- `spa_certificates`
- `spa_browser_mode`
- `spa_hybrid_config`
- `spa_last_activity`
- `spa_terminate_machine_access` / `spa_terminate_user_access`

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `CITRIX_CUSTOMER_ID` | Customer ID for API |
| `CITRIX_AUTH_TOKEN` | Direct bearer token |
| `CITRIX_CLIENT_ID` | Service principal client ID |
| `CITRIX_CLIENT_SECRET` | Service principal secret |
| `SPA_BASE_URL` | Override API base URL |
| `TF_ACC` | Enable acceptance tests (set automatically by `make testacc`) |
| `TF_LOG` | Terraform log level (DEBUG) |
| `TF_LOG_PATH` | Write logs to file |
| `TF_CLI_CONFIG_FILE` | Point to local `.terraformrc` |

## API Conventions

- Header: `Citrix-CustomerId` (capital C, capital I)
- Auth header: `Authorization: CWSAuth bearer={token}`
- Transaction tracking: `Citrix-TransactionId` (UUID per request)
- Rate limit: 15 req/sec (configurable); API limit ~1000 req/min (per source comment in `provider.go`)
- 429 retry: `makeRequest()` automatically retries up to 3 times on `429 Too Many Requests`, respecting the `Retry-After` header (capped at 30s per wait)
- Routing domain paths: FQDN must be URL-encoded via `url.PathEscape()`

## Adding a New Resource

1. Create `internal/provider/resource_foo.go` with `NewFooResource()`
2. Define `FooResourceModel` struct with `tfsdk` tags matching schema
3. Implement `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, `ImportState`
4. Add `GetFoo`, `CreateFoo`, `UpdateFoo`, `DeleteFoo` methods to `SPAClient` interface in `client.go` and implement on `APIClient`
5. Register `NewFooResource` in `provider.go` `Resources()` slice
6. Add acceptance tests in `resource_foo_test.go`
7. Write resource documentation in `docs/resources/foo.md` (documentation is hand-written, not auto-generated)

## Common Pitfalls

| Problem | Root Cause | Fix |
|---------|-----------|-----|
| Spurious plan diff on application SSO field | API returns equivalent but differently-structured JSON | SSO model handles conversion via `ssoFromAPI()` |
| 404 on routing domain | FQDN not URL-encoded | Always use `encodeFQDN()` |
| 400 on routing domain create | Routing domain with that FQDN already exists | Check existing routing domains before creating |
| Spurious diff on computed fields | `state`, `policyCount`, timestamps in state | Mark `Computed: true` + `UseStateForUnknown()` |
| Rate limit 429 errors | Too many requests to API | Automatic retry (up to 3 attempts with `Retry-After`); `fetch_details_on_list = false` could reduce calls |
| 401 mid-run | Token expired | Automatic — `EnsureValidToken()` in `auth.go` |
| List vs detail field mismatch | Different JSON shape from list vs GET endpoints | Use `ApplicationListItem` vs `Application` types |

## Security

- `*.tfvars` are gitignored (`!*.tfvars.example` excluded from ignore)
- Never commit credentials; use environment variables in CI
- Token cache stored at `~/.terraform.d/spa-cache/` with `0600` file permissions
- Tokens encrypted with AES-256-GCM + PBKDF2 key derivation (100,000 iterations)
- Rotate tokens: delete `~/.terraform.d/spa-cache/`
