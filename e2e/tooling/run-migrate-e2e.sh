#!/usr/bin/env bash
# =============================================================================
# SPA Terraform Provider — migrate.ps1 two-tenant E2E  [Option C]
# =============================================================================
# Validates the customer-A -> customer-B migration flow end to end:
#
#   1. Build + install the provider.
#   2. Pre-clean tenant B to a guaranteed-empty state (guarded by a resource
#      count threshold so a populated/wrong tenant is not wiped by accident).
#   3. Seed a known set of resources into tenant A (the e2e/ harness).
#   4. Run migrate.ps1 (parameterized) to migrate A -> B.
#   5. Assert tenant B ends up with the migrated resources.
#   6. Tear down BOTH tenants (terraform destroy + API sweep).
#
# Requires TWO service principals / tenants, supplied via env:
#   Tenant A: CITRIX_CUSTOMER_ID   / CITRIX_CLIENT_ID   / CITRIX_CLIENT_SECRET
#   Tenant B: CITRIX_CUSTOMER_ID_B / CITRIX_CLIENT_ID_B / CITRIX_CLIENT_SECRET_B
#   Optional: SPA_BASE_URL (default https://api.cloud.com/accessSecurity)
#             SPA_TOKEN_URL (default https://api.cloud.com)
#             WIPE_ABORT_THRESHOLD (default 5) — abort the pre-clean if tenant B
#               has more than this many resources, unless CONFIRM_WIPE_TENANT_B=yes.
#
# NOTE: first draft. The assertion here is intentionally light (tenant B
# non-empty with the expected prefix after migrate).
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROVIDER_DIR="$(cd "${E2E_DIR}/.." && pwd)"
TOOL_DIR="${PROVIDER_DIR}/resource-listing-tool"
GOLDEN_DIR="${E2E_DIR}/customer-a"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[C-migrate]${NC} $1"; }
warn()  { echo -e "${YELLOW}[C-migrate]${NC} $1"; }
error() { echo -e "${RED}[C-migrate]${NC} $1"; }

# Fall back to e2e/terraform.tfvars for credentials when not supplied via env.
# Tenant A uses the standard keys; tenant B uses *_b keys in the same file.
TFVARS="${E2E_DIR}/terraform.tfvars"
load_tfvar() { grep -E "^[[:space:]]*$1[[:space:]]*=" "$TFVARS" 2>/dev/null | grep -oE '"[^"]+"' | tr -d '"' | head -1; }
if [ -f "${TFVARS}" ]; then
  : "${CITRIX_CUSTOMER_ID:=$(load_tfvar citrix_customer_id)}"
  : "${CITRIX_CLIENT_ID:=$(load_tfvar citrix_client_id)}"
  : "${CITRIX_CLIENT_SECRET:=$(load_tfvar citrix_client_secret)}"
  : "${CITRIX_CUSTOMER_ID_B:=$(load_tfvar citrix_customer_id_b)}"
  : "${CITRIX_CLIENT_ID_B:=$(load_tfvar citrix_client_id_b)}"
  : "${CITRIX_CLIENT_SECRET_B:=$(load_tfvar citrix_client_secret_b)}"
  : "${SPA_BASE_URL:=$(load_tfvar base_url)}"
  : "${SPA_TOKEN_URL:=$(load_tfvar token_url)}"
fi

: "${CITRIX_CUSTOMER_ID:?set CITRIX_CUSTOMER_ID (tenant A)}"
: "${CITRIX_CLIENT_ID:?set CITRIX_CLIENT_ID (tenant A)}"
: "${CITRIX_CLIENT_SECRET:?set CITRIX_CLIENT_SECRET (tenant A)}"
: "${CITRIX_CUSTOMER_ID_B:?set CITRIX_CUSTOMER_ID_B (tenant B)}"
: "${CITRIX_CLIENT_ID_B:?set CITRIX_CLIENT_ID_B (tenant B)}"
: "${CITRIX_CLIENT_SECRET_B:?set CITRIX_CLIENT_SECRET_B (tenant B)}"
BASE_URL="${SPA_BASE_URL:-https://api.cloud.com/accessSecurity}"
TOKEN_URL="${SPA_TOKEN_URL:-https://api.cloud.com}"
PREFIX="e2e"

# Safety threshold: tenant B is wiped to a clean slate before the migration so
# the "B == A" check is meaningful. If B has MORE than this many resources we
# refuse to wipe it (it is probably a populated/wrong tenant) unless
# CONFIRM_WIPE_TENANT_B=yes is set to override.
WIPE_ABORT_THRESHOLD="${WIPE_ABORT_THRESHOLD:-5}"

# Count total resources in a tenant (apps + routing domains + security groups +
# access policies). Args: customerId clientId clientSecret
# Prints the total resource count on stdout and returns 0 on success. On ANY
# auth/API/parse failure it logs to stderr and returns non-zero WITHOUT printing
# a count, so callers must abort rather than treat a broken tenant as "empty"
# (important: this script can wipe tenant B).
count_tenant_resources() {
  local cid="$1" clid="$2" csec="$3" tok raw n sum=0 ep
  tok=$(curl -sS -X POST "${TOKEN_URL}/cctrustoauth2/${cid}/tokens/clients" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    --data-urlencode grant_type=client_credentials \
    --data-urlencode client_id="${clid}" \
    --data-urlencode client_secret="${csec}" \
    | python3 -c "import sys,json;print(json.load(sys.stdin)['access_token'])" 2>/dev/null) || true
  if [ -z "${tok}" ]; then
    error "count_tenant_resources: failed to acquire token for tenant ${cid}"
    return 1
  fi
  for ep in applications routingDomains securityGroup accessPolicy; do
    raw=$(curl -sS -f "${BASE_URL}/${ep}?limit=-1" -H "Authorization: CWSAuth bearer=${tok}" \
      -H "Citrix-CustomerId: ${cid}" -H "Accept: application/json") || {
      error "count_tenant_resources: API call failed for '${ep}' on tenant ${cid}"
      return 1
    }
    n=$(echo "${raw}" | python3 -c "
import sys, json
d = json.load(sys.stdin)
# Match cleanup.sh: the collection may be a bare array, an {'items':[...]}
# envelope, or a {'<name>':[...]} object — take the first list in the response.
if isinstance(d, dict):
    it = d.get('items')
    if it is None:
        lists = [v for v in d.values() if isinstance(v, list)]
        it = lists[0] if lists else []
else:
    it = d
print(len(it) if isinstance(it, list) else 0)
") || {
      error "count_tenant_resources: could not parse '${ep}' response for tenant ${cid}"
      return 1
    }
    sum=$((sum + n))
  done
  echo "${sum}"
}

# Discover one tenant with spa_manager and import its resources into a Terraform
# state, then copy that state to $4. Args: customerId clientId clientSecret outFile
# The tool dir is reset first so a previous tenant's config/state doesn't leak in.
discover_tenant_to_state() {
  local cid="$1" clid="$2" csec="$3" out="$4"
  rm -f "${TOOL_DIR}/spa_resources.tf" "${TOOL_DIR}/imports.tf" "${TOOL_DIR}/provider.tf" \
        "${TOOL_DIR}/management_summary.md" "${TOOL_DIR}/terraform.tfstate" \
        "${TOOL_DIR}/terraform.tfstate.backup" "${TOOL_DIR}/tfplan" \
        "${TOOL_DIR}/.terraform.lock.hcl"
  rm -rf "${TOOL_DIR}/.terraform"
  cat > "${TOOL_DIR}/terraform.tfvars" <<EOF
customer_id   = "${cid}"
client_id     = "${clid}"
client_secret = "${csec}"
base_url      = "${BASE_URL}"
EOF
  # Generate spa_resources.tf + import blocks for this tenant.
  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" pwsh ./spa_manager.ps1 -List ) || return 1
  [ -f "${TOOL_DIR}/spa_resources.tf" ] || { error "discovery generated no spa_resources.tf for ${cid}"; return 1; }
  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform init -reconfigure -input=false ) || return 1
  # apply realizes the import blocks -> populated state. Because the generated
  # config mirrors live state, this is import-only (0 add/change/destroy), so it
  # does not mutate the tenant.
  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform apply -auto-approve -input=false ) || return 1
  cp "${TOOL_DIR}/terraform.tfstate" "${out}"
}

# Reliably wipe tenant B via the provider: discover -> import -> terraform
# destroy. (The raw API cannot delete in-use routing domains; the provider can.)
# Requires the provider built/installed and TERRAFORMRC generated first.
clean_tenant_b() {
  info "Pre-clean: wiping tenant B (${CITRIX_CUSTOMER_ID_B}) via terraform destroy..."
  cat > "${TOOL_DIR}/terraform.tfvars" <<EOF
customer_id   = "${CITRIX_CUSTOMER_ID_B}"
client_id     = "${CITRIX_CLIENT_ID_B}"
client_secret = "${CITRIX_CLIENT_SECRET_B}"
base_url      = "${BASE_URL}"
EOF
  rm -f "${TOOL_DIR}/spa_resources.tf" "${TOOL_DIR}/imports.tf" "${TOOL_DIR}/provider.tf" \
        "${TOOL_DIR}/management_summary.md" "${TOOL_DIR}/terraform.tfstate" \
        "${TOOL_DIR}/terraform.tfstate.backup" "${TOOL_DIR}/.terraform.lock.hcl" "${TOOL_DIR}/tfplan"
  rm -rf "${TOOL_DIR}/.terraform"
  ( cd "${TOOL_DIR}" && pwsh ./spa_manager.ps1 -List ) || { warn "pre-clean discovery failed"; return 1; }
  ( cd "${TOOL_DIR}" && terraform init -reconfigure -input=false \
      && terraform apply -auto-approve -input=false ) || { warn "pre-clean import failed"; return 1; }
  ( cd "${TOOL_DIR}" && terraform destroy -auto-approve -input=false ) || { warn "pre-clean destroy failed"; return 1; }
}

export TF_VAR_run_tag="${GITHUB_RUN_ID:-local-$(date +%s)}"
export TF_VAR_base_url="${BASE_URL}"
export TF_VAR_token_url="${TOKEN_URL}"

TERRAFORMRC="${E2E_DIR}/.terraformrc"
export TF_CLI_CONFIG_FILE="${TERRAFORMRC}"

sweep() {
  local cid="$1" clid="$2" csec="$3"
  ( cd "${E2E_DIR}" && PREFIX="${PREFIX}" \
      CITRIX_CUSTOMER_ID="${cid}" CITRIX_CLIENT_ID="${clid}" CITRIX_CLIENT_SECRET="${csec}" \
      SPA_BASE_URL="${BASE_URL}" SPA_TOKEN_URL="${TOKEN_URL}" ./cleanup.sh ) || warn "sweep on ${cid} errored"
}

teardown() {
  local rc=$?
  if [ "${KEEP_RESOURCES:-0}" = "1" ]; then
    warn "KEEP_RESOURCES=1 set — skipping teardown. Migrated resources remain in both tenants."
    warn "Clean up later with:  ./tooling/run-migrate-cleanup.sh   (or re-run without KEEP_RESOURCES)."
    exit "${rc}"
  fi
  # Reliable teardown = terraform destroy (the API sweep alone cannot delete
  # in-use routing domains). Tenant B lives in the spa_manager state under
  # resource-listing-tool/. Customer A is PERSISTENT (golden) and is left in
  # place unless DESTROY_AFTER=1. API sweep is a fallback.
  info "Teardown: destroying tenant B (spa_manager state)..."
  ( cd "${TOOL_DIR}" && terraform destroy -auto-approve -input=false ) || warn "tenant B terraform destroy errored"
  if [ "${DESTROY_AFTER:-0}" = "1" ]; then
    info "Teardown: destroying Customer A golden (customer-a/ state)..."
    ( cd "${GOLDEN_DIR}" \
        && TF_VAR_citrix_customer_id="${CITRIX_CUSTOMER_ID}" \
           TF_VAR_citrix_client_id="${CITRIX_CLIENT_ID}" \
           TF_VAR_citrix_client_secret="${CITRIX_CLIENT_SECRET}" \
           terraform destroy -auto-approve -input=false ) || warn "Customer A destroy errored"
    sweep "${CITRIX_CUSTOMER_ID}" "${CITRIX_CLIENT_ID}" "${CITRIX_CLIENT_SECRET}"
  fi
  info "Teardown: API sweep fallback on tenant B..."
  sweep "${CITRIX_CUSTOMER_ID_B}" "${CITRIX_CLIENT_ID_B}" "${CITRIX_CLIENT_SECRET_B}"
  exit "${rc}"
}
trap teardown EXIT

command -v pwsh >/dev/null 2>&1 || { error "pwsh (PowerShell 7) required"; exit 1; }

# --- tenant B safety pre-check (fail fast, before the build) ------------------
info "Checking tenant B (${CITRIX_CUSTOMER_ID_B}) before migration..."
if ! B_EXISTING="$(count_tenant_resources "${CITRIX_CUSTOMER_ID_B}" "${CITRIX_CLIENT_ID_B}" "${CITRIX_CLIENT_SECRET_B}")"; then
  error "ABORT: could not determine tenant B resource count (auth/API/parse failure)."
  error "Refusing to continue because the pre-clean safety check cannot be trusted."
  exit 1
fi
info "Tenant B currently has ${B_EXISTING} resource(s); safety threshold is ${WIPE_ABORT_THRESHOLD}."
if [ "${B_EXISTING}" -gt "${WIPE_ABORT_THRESHOLD}" ] && [ "${CONFIRM_WIPE_TENANT_B:-}" != "yes" ]; then
  error "ABORT: tenant B has ${B_EXISTING} resources (> ${WIPE_ABORT_THRESHOLD})."
  error "This likely isn't a clean test tenant — refusing to wipe it before migration."
  error "Set CONFIRM_WIPE_TENANT_B=yes to wipe intentionally, or point CITRIX_*_B at a dedicated tenant."
  exit 1
fi

# --- build + install provider -------------------------------------------------
info "Building and installing provider..."
( cd "${PROVIDER_DIR}" && make build && make install-local ) || { error "provider build/install failed"; exit 1; }

PLUGIN_DIR="${TF_PLUGIN_DIR:-$HOME/.terraform.d/plugins}"
cat > "${TERRAFORMRC}" <<EOF
provider_installation {
  filesystem_mirror { path = "${PLUGIN_DIR}" include = ["registry.terraform.io/citrix/citrixspa"] }
  direct { exclude = ["registry.terraform.io/citrix/citrixspa"] }
}
EOF

# migrate.ps1 (via spa_manager.ps1) pins a specific provider version in its
# generated provider.tf. Mirror the freshly built binary under that version too
# so the filesystem_mirror can satisfy the tool's version constraint.
TOOL_VERSION="$(grep -oE 'version[[:space:]]*=[[:space:]]*"[0-9]+\.[0-9]+\.[0-9]+"' "${TOOL_DIR}/spa_manager.ps1" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"
# Fail fast: without the pinned version we can't mirror the local build, and the
# generated .terraformrc excludes direct downloads for citrix/citrixspa, so a
# later `terraform init` would otherwise fail in a much harder-to-diagnose way.
if [ -z "${TOOL_VERSION}" ]; then
  error "could not parse the pinned provider version from spa_manager.ps1; cannot mirror the local provider build"
  exit 1
fi
ARCH="$(go env GOOS)_$(go env GOARCH)"
DEST="${PLUGIN_DIR}/registry.terraform.io/citrix/citrixspa/${TOOL_VERSION}/${ARCH}"
info "Mirroring provider build as v${TOOL_VERSION} for spa_manager (${DEST})..."
mkdir -p "${DEST}"
cp "${PROVIDER_DIR}/terraform-provider-citrixspa" "${DEST}/"

# --- pre-clean tenant B (guaranteed clean slate) ------------------------------
if [ "${B_EXISTING}" -gt 0 ]; then
  # Pre-clean is safety-critical: migrating into a non-empty tenant B produces
  # collisions and false compare results, so abort unless B is confirmed clean.
  clean_tenant_b || { error "ABORT: pre-clean of tenant B failed; refusing to migrate into a dirty tenant."; exit 1; }
  if ! B_LEFT="$(count_tenant_resources "${CITRIX_CUSTOMER_ID_B}" "${CITRIX_CLIENT_ID_B}" "${CITRIX_CLIENT_SECRET_B}")"; then
    error "ABORT: could not recount tenant B after pre-clean (auth/API error); cannot confirm it is clean."
    exit 1
  fi
  if [ "${B_LEFT}" -ne 0 ]; then
    error "ABORT: tenant B still has ${B_LEFT} resource(s) after pre-clean; refusing to migrate into a dirty tenant."
    exit 1
  fi
  info "Tenant B is clean."
else
  info "Tenant B already empty; no pre-clean needed."
fi

# --- ensure Customer A has the golden config ----------------------------------
info "Applying Customer A golden config (customer-a/)..."
( cd "${GOLDEN_DIR}" \
    && TF_VAR_citrix_customer_id="${CITRIX_CUSTOMER_ID}" \
       TF_VAR_citrix_client_id="${CITRIX_CLIENT_ID}" \
       TF_VAR_citrix_client_secret="${CITRIX_CLIENT_SECRET}" \
       terraform init -reconfigure -input=false \
    && TF_VAR_citrix_customer_id="${CITRIX_CUSTOMER_ID}" \
       TF_VAR_citrix_client_id="${CITRIX_CLIENT_ID}" \
       TF_VAR_citrix_client_secret="${CITRIX_CLIENT_SECRET}" \
       terraform apply -auto-approve -input=false ) \
  || { error "Customer A golden apply failed"; exit 1; }

# --- run migrate.ps1 A -> B ---------------------------------------------------
# Tenant B was pre-cleaned above, so migrate.ps1's own (fragile) reset is
# disabled here. `-Param:\$false` is deliberate and verified: bash turns `\$`
# into a literal `$`, so PowerShell's -File colon syntax receives `-Param:$false`
# and binds a real [bool] $false. (The quoted form `'$false'` is byte-identical
# and also works; the escaped form matches PowerShell conventions.)
info "Running migrate.ps1 (tenant A -> tenant B)..."
( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" pwsh ./migrate.ps1 \
    -CustomerA "${CITRIX_CUSTOMER_ID}" -ClientIdA "${CITRIX_CLIENT_ID}" -ClientSecretA "${CITRIX_CLIENT_SECRET}" \
    -CustomerB "${CITRIX_CUSTOMER_ID_B}" -ClientIdB "${CITRIX_CLIENT_ID_B}" -ClientSecretB "${CITRIX_CLIENT_SECRET_B}" \
    -BaseUrl "${BASE_URL}" -ResetCustomerBBeforeMigration:\$false ) \
  || { error "migrate.ps1 failed"; exit 1; }

# --- verify A -> B fidelity via state-to-state comparison ---------------------
# Discover BOTH tenants into their own Terraform state (same generator, so the
# resource attributes line up), then compare A vs B with compare-tenants.py.
# The comparator ignores tenant-specific fields (ids/timestamps/computed) and
# remaps app-id cross-references to names, so it asserts every Customer A
# resource migrated to Customer B field-for-field — not merely that B is
# non-empty.
A_STATE="$(mktemp -t stateA.XXXXXX.tfstate)"
B_STATE="$(mktemp -t stateB.XXXXXX.tfstate)"

info "Discovering Customer A into a Terraform state..."
discover_tenant_to_state "${CITRIX_CUSTOMER_ID}" "${CITRIX_CLIENT_ID}" "${CITRIX_CLIENT_SECRET}" "${A_STATE}" \
  || { error "failed to discover Customer A into state"; exit 1; }

# Customer B's discovery LIST can lag right after the migration apply
# (read-after-write list-indexing eventual consistency), surfacing a just-created
# resource as "missing". Re-discover B and re-compare with exponential backoff
# before declaring a mismatch: 1 initial attempt + 2 retries.
compare_attempts=3
compare_backoff=15
compare_attempt=1
while true; do
  info "Discovering Customer B into a Terraform state (attempt ${compare_attempt}/${compare_attempts})..."
  discover_tenant_to_state "${CITRIX_CUSTOMER_ID_B}" "${CITRIX_CLIENT_ID_B}" "${CITRIX_CLIENT_SECRET_B}" "${B_STATE}" \
    || { error "failed to discover Customer B into state"; exit 1; }

  info "Comparing Customer A vs Customer B resources (state-to-state diff)..."
  if python3 "${SCRIPT_DIR}/compare-tenants.py" "${A_STATE}" "${B_STATE}"; then
    break
  fi

  if [ "${compare_attempt}" -ge "${compare_attempts}" ]; then
    error "migrate E2E FAILED: Customer B does not match Customer A after ${compare_attempts} attempts (see diff above)."
    exit 1
  fi

  warn "Compare mismatch on attempt ${compare_attempt}; Customer B may still be indexing. Retrying in ${compare_backoff}s..."
  sleep "${compare_backoff}"
  compare_attempt=$((compare_attempt + 1))
  compare_backoff=$((compare_backoff * 2))
done
rm -f "${A_STATE}" "${B_STATE}"

info "migrate.ps1 E2E PASSED. Teardown will sweep both tenants next."
