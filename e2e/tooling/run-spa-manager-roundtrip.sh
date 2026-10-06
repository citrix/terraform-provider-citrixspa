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
clean_tool_dir() {
  rm -f "${TOOL_DIR}/spa_resources.tf" "${TOOL_DIR}/imports.tf" "${TOOL_DIR}/provider.tf" \
        "${TOOL_DIR}/management_summary.md" "${TOOL_DIR}/terraform.tfstate" \
        "${TOOL_DIR}/terraform.tfstate.backup" "${TOOL_DIR}/.terraform.lock.hcl" "${TOOL_DIR}/tfplan"
  rm -rf "${TOOL_DIR}/.terraform"
}

# discover_and_plan <label> <plan_file> [extra spa_manager args...]
#
# One full discovery cycle: clean, generate, init, plan, assert the plan is
# clean. Leaves the plan text in <plan_file> and the generated spa_resources.tf
# in place for the caller to inspect.
#
# The generated config ships import blocks, so the plan legitimately shows
# "N to import". Discovery is accurate when that plan has 0 add/change/destroy —
# every resource matches live state and only needs importing. (We therefore do
# NOT use -detailed-exitcode, which treats imports as changes.)
discover_and_plan() {
  local label="$1" plan_file="$2"
  shift 2

  info "[${label}] Cleaning tool directory before discovery..."
  clean_tool_dir

  info "[${label}] Running spa_manager.ps1 -List $*..."
  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" pwsh ./spa_manager.ps1 -List "$@" ) \
    || { error "[${label}] spa_manager.ps1 -List $* failed"; return 1; }

  if [ ! -f "${TOOL_DIR}/spa_resources.tf" ]; then
    error "[${label}] spa_manager.ps1 did not generate spa_resources.tf"
    return 1
  fi

  info "[${label}] Planning against generated config (expecting 0 add/change/destroy)..."
  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform init -reconfigure -input=false ) \
    || { error "[${label}] terraform init against generated config failed"; return 1; }

  ( cd "${TOOL_DIR}" && TF_CLI_CONFIG_FILE="${TERRAFORMRC}" terraform plan -input=false -no-color ) > "${plan_file}" 2>&1
  local plan_rc=$?
  cat "${plan_file}"
  if [ "${plan_rc}" -ne 0 ]; then
    error "[${label}] terraform plan errored (exit ${plan_rc})."
    return 1
  fi

  local summary add change destroy imports
  summary="$(grep -E '^(Plan:|No changes)' "${plan_file}" | tail -1)"
  get_count() { echo "${summary}" | grep -oE "[0-9]+ to $1" | grep -oE '[0-9]+' | head -1; }
  add="$(get_count add)";         add="${add:-0}"
  change="$(get_count change)";   change="${change:-0}"
  destroy="$(get_count destroy)"; destroy="${destroy:-0}"
  imports="$(get_count import)";  imports="${imports:-0}"

  # Guard against a trivially-clean plan: discovery must have found the golden
  # resources. Their attribute values retain the "e2e-golden" prefix.
  local seeded_marker="e2e-golden"
  if ! grep -q "${seeded_marker}" "${plan_file}"; then
    error "[${label}] FAILED: discovery did not include the seeded resources ('${seeded_marker}'). Empty/partial discovery?"
    return 1
  fi
  if [ "${imports}" -lt 1 ]; then
    error "[${label}] FAILED: no import blocks generated — nothing was discovered."
    return 1
  fi
  if [ "${add}" -ne 0 ] || [ "${change}" -ne 0 ] || [ "${destroy}" -ne 0 ]; then
    error "[${label}] FAILED: generated config shows drift (add=${add} change=${change} destroy=${destroy})."
    return 1
  fi

  info "[${label}] PASSED (imports=${imports}, 0 add/change/destroy, seeded resources present)."
  return 0
}

# --- pass 1: default literal output -------------------------------------------
plan_literal="$(mktemp)"
discover_and_plan "literal" "${plan_literal}" || { rm -f "${plan_literal}"; exit 1; }

# --- pass 2: -ExtractLocals ----------------------------------------------------
# Same tenant, same assertions, plus proof that the locals actually collapsed
# something. That extra proof matters because the failure mode of -ExtractLocals
# is QUIET: if a value stops being catalogued, the emitter falls back to the
# literal, which is still valid HCL and still plans clean. Only counting the
# collapse catches it.
plan_locals="$(mktemp)"
discover_and_plan "extract-locals" "${plan_locals}" -ExtractLocals \
  || { rm -f "${plan_literal}" "${plan_locals}"; exit 1; }

generated="${TOOL_DIR}/spa_resources.tf"
locals_failed=0
fail_locals() { error "[extract-locals] FAILED: $1"; locals_failed=1; }

# 1. The flag did something at all.
grep -q '^locals {' "${generated}" || fail_locals "no locals block in the generated config"

# 2. The shared values actually collapsed. These UUIDs and tokens are seeded by
#    e2e/customer-a/main.tf specifically so there is repetition to remove; see
#    the "-ExtractLocals tooling pass" locals there. Each UUID is used by
#    several routing domains/applications, so after collapsing it must appear
#    exactly once — inside the locals block.
# Count occurrences, not matching lines: concat()/merge() put two references on
# a single line, so `grep -c` would undercount them.
count_occurrences() { grep -o "$1" "$2" | wc -l; }

for uuid in "00000000-0000-0000-0000-0000000000aa" "00000000-0000-0000-0000-0000000000bb"; do
  n="$(count_occurrences "${uuid}" "${generated}")"
  [ "${n}" -eq 1 ] || fail_locals "resource location ${uuid} appears ${n} times, expected exactly 1 (inside locals)"
done

n="$(count_occurrences 'local\.resource_locations\.' "${generated}")"
[ "${n}" -ge 5 ] || fail_locals "only ${n} local.resource_locations references, expected >= 5"

n="$(count_occurrences 'local\.users\.' "${generated}")"
[ "${n}" -ge 4 ] || fail_locals "only ${n} local.users references, expected >= 4"

n="$(count_occurrences 'local\.user_metadata\.' "${generated}")"
[ "${n}" -ge 4 ] || fail_locals "only ${n} local.user_metadata references, expected >= 4"

# The >= bounds above are deliberately loose, and loose enough to hide a real
# regression: "Golden Shared Group" is seeded into exactly four rules (three
# access policies plus the session policy) and "Golden Second Group" into one,
# so if the session-policy emitter stopped collapsing, local.users. would drop
# from 5 to 4 and still pass. Pin the shared identity's own count exactly.
# The key is Get-SharedLocalName's sanitisation of the display name; a rename
# in e2e/customer-a/main.tf has to be mirrored here.
for prefix in 'local\.users' 'local\.user_metadata'; do
  n="$(count_occurrences "${prefix}\.Golden_Shared_Group" "${generated}")"
  [ "${n}" -eq 4 ] || fail_locals "${prefix}.Golden_Shared_Group appears ${n} times, expected exactly 4 (3 access policies + 1 session policy)"
done

# The two-identity rule is the only shape that produces concat()/merge().
grep -q 'concat(local\.users\.' "${generated}" || fail_locals "no concat(local.users...) — the two-identity rule did not collapse"
grep -q 'merge(local\.user_metadata\.' "${generated}" || fail_locals "no merge(local.user_metadata...) — the two-identity rule did not collapse"

# 3. The literal fallback is still intact. `values = ["Everyone"]` carries no
#    metadata and must never be hoisted; if it is, the extractability rules have
#    been loosened too far.
grep -q '"Everyone"' "${generated}" || fail_locals "no literal \"Everyone\" rule left — the fallback path regressed"

# 4. Both passes must produce the same plan. Once the golden tenant is
#    repetitive, any difference between the literal and refactored plans is a
#    real bug, not noise.
#
#    Terraform refreshes resources concurrently, so the interleaved
#    "Preparing import... / Refreshing state..." progress lines come out in a
#    different order on every run — that IS noise, and it is dropped before the
#    comparison. Everything else (the planned actions and the Plan: summary) is
#    deterministic and is compared verbatim.
strip_plan_progress() {
  grep -vE ': (Preparing import\.\.\.|Refreshing state\.\.\.|Reading\.\.\.|Read complete after )' "$1"
}
plan_diff="$(mktemp)"
strip_plan_progress "${plan_literal}" > "${plan_literal}.norm"
strip_plan_progress "${plan_locals}"  > "${plan_locals}.norm"
if ! diff -u "${plan_literal}.norm" "${plan_locals}.norm" > "${plan_diff}" 2>&1; then
  error "[extract-locals] FAILED: plan differs between the literal and -ExtractLocals passes:"
  head -60 "${plan_diff}"
  locals_failed=1
fi

rm -f "${plan_literal}" "${plan_locals}" "${plan_diff}" \
      "${plan_literal}.norm" "${plan_locals}.norm"

if [ "${locals_failed}" -ne 0 ]; then
  exit 1
fi

info "spa_manager discovery round-trip PASSED for both the literal and -ExtractLocals passes. Teardown will run next."
