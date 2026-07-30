#!/usr/bin/env bash
# =============================================================================
# SPA Terraform Provider — Customer A GOLDEN check  [scenario: golden]
# =============================================================================
# Applies the committed golden configuration (customer-a/) to the dedicated
# Customer A tenant and asserts there is NO drift (the live tenant matches the
# reference config). Replaces the old ephemeral "lifecycle" scenario.
#
# Model: create -> validate (no drift) -> tear down. Set DESTROY_AFTER=1 to
# destroy the golden resources at the end (CI default, and recommended locally so
# a tenant isn't left populated).
#
# Credentials: env (CITRIX_CUSTOMER_ID/CLIENT_ID/CLIENT_SECRET) or
# e2e/terraform.tfvars (Customer A keys).
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="${SCRIPT_DIR}"
PROVIDER_DIR="$(cd "${E2E_DIR}/.." && pwd)"
GOLDEN_DIR="${E2E_DIR}/customer-a"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[golden]${NC} $1"; }
warn()  { echo -e "${YELLOW}[golden]${NC} $1"; }
error() { echo -e "${RED}[golden]${NC} $1"; }

# Load Customer A credentials from terraform.tfvars if not in the environment.
TFVARS="${E2E_DIR}/terraform.tfvars"
load_tfvar() { grep -E "^[[:space:]]*$1[[:space:]]*=" "$TFVARS" 2>/dev/null | grep -oE '"[^"]+"' | tr -d '"' | head -1; }
if [ -f "${TFVARS}" ]; then
  : "${CITRIX_CUSTOMER_ID:=$(load_tfvar citrix_customer_id)}"
  : "${CITRIX_CLIENT_ID:=$(load_tfvar citrix_client_id)}"
  : "${CITRIX_CLIENT_SECRET:=$(load_tfvar citrix_client_secret)}"
  : "${SPA_BASE_URL:=$(load_tfvar base_url)}"
  : "${SPA_TOKEN_URL:=$(load_tfvar token_url)}"
fi
: "${CITRIX_CUSTOMER_ID:?set CITRIX_CUSTOMER_ID (Customer A)}"
: "${CITRIX_CLIENT_ID:?set CITRIX_CLIENT_ID (Customer A)}"
: "${CITRIX_CLIENT_SECRET:?set CITRIX_CLIENT_SECRET (Customer A)}"
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
  if [ "${DESTROY_AFTER:-0}" = "1" ]; then
    info "DESTROY_AFTER=1 — destroying golden resources..."
    ( cd "${GOLDEN_DIR}" && terraform destroy -auto-approve -input=false ) || warn "golden destroy errored"
  fi
  exit "${rc}"
}
trap teardown EXIT

# --- build + install provider -------------------------------------------------
info "Building and installing the provider..."
( cd "${PROVIDER_DIR}" && make build && make install-local ) || { error "provider build/install failed"; exit 1; }

PLUGIN_DIR="${TF_PLUGIN_DIR:-$HOME/.terraform.d/plugins}"
cat > "${TERRAFORMRC}" <<EOF
provider_installation {
  filesystem_mirror { path = "${PLUGIN_DIR}" include = ["registry.terraform.io/citrix/citrixspa"] }
  direct { exclude = ["registry.terraform.io/citrix/citrixspa"] }
}
EOF

# --- apply golden config + assert no drift ------------------------------------
info "terraform init (customer-a)"
( cd "${GOLDEN_DIR}" && terraform init -reconfigure -input=false ) || exit 1

info "terraform apply (create golden resources)"
( cd "${GOLDEN_DIR}" && terraform apply -auto-approve -input=false ) || { error "golden apply failed"; exit 1; }

info "Drift check: terraform plan (expecting no changes)"
( cd "${GOLDEN_DIR}" && terraform plan -detailed-exitcode -input=false )
plan_rc=$?
if [ "${plan_rc}" -eq 2 ]; then
  error "GOLDEN DRIFT DETECTED: Customer A does not match customer-a/ config."
  exit 1
elif [ "${plan_rc}" -ne 0 ]; then
  error "terraform plan errored (exit ${plan_rc})."
  exit 1
fi

info "Golden check PASSED — Customer A matches the reference config (no drift)."
