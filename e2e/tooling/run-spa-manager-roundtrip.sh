#!/usr/bin/env bash
# =============================================================================
# SPA Terraform Provider — spa_manager.ps1 discovery round-trip  [Option C]
# =============================================================================
# Validates that the PowerShell discovery/generation tool faithfully reproduces
# the live tenant state:
#
#   1. Build + install the provider.
#   2. Seed a known set of resources into the tenant (the e2e/ harness).
#   3. Run `spa_manager.ps1 -List` to discover them and generate spa_resources.tf
#      + import blocks in resource-listing-tool/.
#   4. `terraform plan` against the generated config and assert NO changes
#      (exit 0) — i.e. discovery captured reality accurately.
#   5. Tear everything down (harness destroy + API sweep).
#
# Credentials come from env (CI): CITRIX_CUSTOMER_ID / CITRIX_CLIENT_ID /
# CITRIX_CLIENT_SECRET, plus optional SPA_BASE_URL / SPA_TOKEN_URL.
#
# NOTE: first draft. The exact plan-clean assertion may need tuning against the
# real spa_manager.ps1 output (some computed fields can produce benign diffs);
# treat a non-empty plan as a signal to investigate, not necessarily a hard bug.
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROVIDER_DIR="$(cd "${E2E_DIR}/.." && pwd)"
TOOL_DIR="${PROVIDER_DIR}/resource-listing-tool"
GOLDEN_DIR="${E2E_DIR}/customer-a"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[C-roundtrip]${NC} $1"; }
warn()  { echo -e "${YELLOW}[C-roundtrip]${NC} $1"; }
error() { echo -e "${RED}[C-roundtrip]${NC} $1"; }

# Fall back to e2e/terraform.tfvars for credentials when not supplied via env
# (keeps local runs consistent with run-e2e.sh).
TFVARS="${E2E_DIR}/terraform.tfvars"
# Anchor to `key =` so e.g. `citrix_customer_id` does not also match
# `citrix_customer_id_b` (which would load the wrong tenant's credentials).
load_tfvar() { grep -E "^[[:space:]]*$1[[:space:]]*=" "$TFVARS" 2>/dev/null | grep -oE '"[^"]+"' | tr -d '"' | head -1; }
if [ -f "${TFVARS}" ]; then
  : "${CITRIX_CUSTOMER_ID:=$(load_tfvar citrix_customer_id)}"
  : "${CITRIX_CLIENT_ID:=$(load_tfvar citrix_client_id)}"
  : "${CITRIX_CLIENT_SECRET:=$(load_tfvar citrix_client_secret)}"
  : "${SPA_BASE_URL:=$(load_tfvar base_url)}"
  : "${SPA_TOKEN_URL:=$(load_tfvar token_url)}"
fi

: "${CITRIX_CUSTOMER_ID:?set CITRIX_CUSTOMER_ID}"
: "${CITRIX_CLIENT_ID:?set CITRIX_CLIENT_ID}"
: "${CITRIX_CLIENT_SECRET:?set CITRIX_CLIENT_SECRET}"
BASE_URL="${SPA_BASE_URL:-https://api.cloud.com/accessSecurity}"
TOKEN_URL="${SPA_TOKEN_URL:-https://api.cloud.com}"

export TF_VAR_citrix_customer_id="${CITRIX_CUSTOMER_ID}"
export TF_VAR_citrix_client_id="${CITRIX_CLIENT_ID}"
export TF_VAR_citrix_client_secret="${CITRIX_CLIENT_SECRET}"
export TF_VAR_base_url="${BASE_URL}"
export TF_VAR_token_url="${TOKEN_URL}"

TERRAFORMRC="${E2E_DIR}/.terraformrc"
export TF_CLI_CONFIG_FILE="${TERRAFORMRC}"

teardown() {
  local rc=$?
  # Customer A is persistent (golden). By default leave it in place; set
  # DESTROY_AFTER=1 to remove the golden resources (e.g. against a shared tenant).
  if [ "${DESTROY_AFTER:-0}" = "1" ]; then
    info "Teardown: destroying golden resources and sweeping orphans..."
    ( cd "${GOLDEN_DIR}" && terraform destroy -auto-approve -input=false ) || warn "golden destroy errored"
    ( cd "${E2E_DIR}" && PREFIX="e2e" CITRIX_CUSTOMER_ID="${CITRIX_CUSTOMER_ID}" \
        CITRIX_CLIENT_ID="${CITRIX_CLIENT_ID}" CITRIX_CLIENT_SECRET="${CITRIX_CLIENT_SECRET}" \
        SPA_BASE_URL="${BASE_URL}" SPA_TOKEN_URL="${TOKEN_URL}" ./cleanup.sh ) || warn "sweep errored"
  fi
  exit "${rc}"
}
trap teardown EXIT

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

# spa_manager.ps1 pins a specific provider version in its generated provider.tf.
# Mirror the freshly built binary under that version too, so the filesystem_mirror
# can satisfy the tool's version constraint with our local build.
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

# --- ensure Customer A has the golden config ----------------------------------
info "Applying Customer A golden config (customer-a/)..."
( cd "${GOLDEN_DIR}" && terraform init -reconfigure -input=false && terraform apply -auto-approve -input=false ) \
  || { error "golden apply failed"; exit 1; }

# --- run discovery tool -------------------------------------------------------
info "Writing resource-listing-tool/terraform.tfvars and running spa_manager.ps1 -List..."
cat > "${TOOL_DIR}/terraform.tfvars" <<EOF
customer_id   = "${CITRIX_CUSTOMER_ID}"
client_id     = "${CITRIX_CLIENT_ID}"
client_secret = "${CITRIX_CLIENT_SECRET}"
base_url      = "${BASE_URL}"
EOF

if ! command -v pwsh >/dev/null 2>&1; then
  error "pwsh (PowerShell 7) is required for spa_manager.ps1"
  exit 1
fi

# Clean any state/generated files from a previous run — leftovers cause
# spa_manager to short-circuit discovery and emit an empty config.
info "Cleaning tool directory before discovery..."
rm -f "${TOOL_DIR}/spa_resources.tf" "${TOOL_DIR}/imports.tf" "${TOOL_DIR}/provider.tf" \
      "${TOOL_DIR}/management_summary.md" "${TOOL_DIR}/terraform.tfstate" \
      "${TOOL_DIR}/terraform.tfstate.backup" "${TOOL_DIR}/.terraform.lock.hcl" "${TOOL_DIR}/tfplan"
rm -rf "${TOOL_DIR}/.terraform"

( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" pwsh ./spa_manager.ps1 -List ) \
  || { error "spa_manager.ps1 -List failed"; exit 1; }

if [ ! -f "${TOOL_DIR}/spa_resources.tf" ]; then
  error "spa_manager.ps1 did not generate spa_resources.tf"
  exit 1
fi

# --- assert the generated config reproduces reality (no drift) ----------------
# The generated config ships import blocks, so a first plan legitimately shows
# "N to import". Discovery is accurate when that plan has 0 add/change/destroy —
# every resource matches live state and only needs importing. (We therefore do
# NOT use -detailed-exitcode, which treats imports as changes.)
info "Planning against generated config (expecting 0 add/change/destroy)..."
( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform init -reconfigure -input=false ) \
  || { error "terraform init against generated config failed"; exit 1; }

plan_out="$(mktemp)"
( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform plan -input=false -no-color ) > "${plan_out}" 2>&1
plan_rc=$?
cat "${plan_out}"
if [ "${plan_rc}" -ne 0 ]; then
  error "terraform plan errored (exit ${plan_rc})."
  exit 1
fi

summary="$(grep -E '^(Plan:|No changes)' "${plan_out}" | tail -1)"
get_count() { echo "${summary}" | grep -oE "[0-9]+ to $1" | grep -oE '[0-9]+' | head -1; }
add="$(get_count add)";     add="${add:-0}"
change="$(get_count change)"; change="${change:-0}"
destroy="$(get_count destroy)"; destroy="${destroy:-0}"
imports="$(get_count import)"; imports="${imports:-0}"

# Guard against a trivially-clean plan: discovery must have found the golden
# resources. Their attribute values retain the "e2e-golden" prefix.
seeded_marker="e2e-golden"
if ! grep -q "${seeded_marker}" "${plan_out}"; then
  error "Round-trip FAILED: discovery did not include the seeded resources ('${seeded_marker}'). Empty/partial discovery?"
  rm -f "${plan_out}"
  exit 1
fi
if [ "${imports}" -lt 1 ]; then
  error "Round-trip FAILED: no import blocks generated — nothing was discovered."
  rm -f "${plan_out}"
  exit 1
fi
rm -f "${plan_out}"

if [ "${add}" -ne 0 ] || [ "${change}" -ne 0 ] || [ "${destroy}" -ne 0 ]; then
  error "Round-trip FAILED: generated config shows drift (add=${add} change=${change} destroy=${destroy})."
  exit 1
fi

info "spa_manager discovery round-trip PASSED (imports=${imports}, 0 add/change/destroy, seeded resources present). Teardown will run next."
