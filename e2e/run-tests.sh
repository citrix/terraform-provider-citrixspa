#!/usr/bin/env bash
# =============================================================================
# SPA Terraform Provider — unit + acceptance tests  [scenario: tests]
# =============================================================================
# Runs the Go unit tests (always) and the acceptance tests (TF_ACC=1) against
# the dedicated Customer A E2E tenant. Included in the `all` e2e run.
#
# Acceptance tests create and destroy their own uniquely-named resources, so
# they are safe to run against the persistent Customer A tenant.
#
# Credentials: env (CITRIX_CUSTOMER_ID/CLIENT_ID/CLIENT_SECRET) or
# e2e/terraform.tfvars (tenant A keys).
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="${SCRIPT_DIR}"
PROVIDER_DIR="$(cd "${E2E_DIR}/.." && pwd)"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[tests]${NC} $1"; }
error() { echo -e "${RED}[tests]${NC} $1"; }

TFVARS="${E2E_DIR}/terraform.tfvars"
load_tfvar() { grep -E "^[[:space:]]*$1[[:space:]]*=" "$TFVARS" 2>/dev/null | grep -oE '"[^"]+"' | tr -d '"' | head -1; }
if [ -f "${TFVARS}" ]; then
  : "${CITRIX_CUSTOMER_ID:=$(load_tfvar citrix_customer_id)}"
  : "${CITRIX_CLIENT_ID:=$(load_tfvar citrix_client_id)}"
  : "${CITRIX_CLIENT_SECRET:=$(load_tfvar citrix_client_secret)}"
  : "${SPA_BASE_URL:=$(load_tfvar base_url)}"
fi
: "${CITRIX_CUSTOMER_ID:?set CITRIX_CUSTOMER_ID (Customer A)}"
: "${CITRIX_CLIENT_ID:?set CITRIX_CLIENT_ID (Customer A)}"
: "${CITRIX_CLIENT_SECRET:?set CITRIX_CLIENT_SECRET (Customer A)}"

cd "${PROVIDER_DIR}"

info "go build"
go build ./... || { error "build failed"; exit 1; }

info "go vet"
go vet ./... || { error "vet failed"; exit 1; }

info "Unit tests"
go test -v ./... || { error "unit tests failed"; exit 1; }

info "Acceptance tests (TF_ACC=1) against Customer A"
export TF_ACC=1
export CITRIX_CUSTOMER_ID CITRIX_CLIENT_ID CITRIX_CLIENT_SECRET
[ -n "${SPA_BASE_URL:-}" ] && export SPA_BASE_URL
# Restrict to acceptance tests (TestAcc*) so the unit tests aren't re-run here
# (they already ran in the previous step) — speeds up the tests scenario.
go test -v -count=1 -run '^TestAcc' ./... -timeout 120m || { error "acceptance tests failed"; exit 1; }

info "tests scenario PASSED (unit + acceptance)."
