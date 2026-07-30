#!/usr/bin/env bash
# =============================================================================
# SPA Terraform Provider — E2E orphan sweeper
# =============================================================================
# Best-effort deletion of every E2E resource (security groups, access policies,
# applications, routing domains) whose name/fqdn starts with $PREFIX, via the
# SPA API. Idempotent — safe to re-run until 0 remain. Used as a safety net
# when `terraform destroy` fails or leaves orphans.
#
# Credentials are read from environment variables first (CI), then fall back to
# terraform.tfvars (local):
#   CITRIX_CUSTOMER_ID / CITRIX_CLIENT_ID / CITRIX_CLIENT_SECRET
#   SPA_BASE_URL (default https://api.cloud.com/accessSecurity)
#   SPA_TOKEN_URL (default https://api.cloud.com)
#   PREFIX (default "e2e")
#
# TLS verification is ON by default. Set INSECURE=1 to skip it (dev endpoints).
# =============================================================================
set -uo pipefail
cd "$(dirname "$0")" || exit 1

PREFIX="${PREFIX:-e2e}"

# --- resolve credentials: env first, then terraform.tfvars --------------------
# Anchor to `key =` so e.g. `citrix_customer_id` does not also match
# `citrix_customer_id_b` (which would load the wrong tenant's credentials).
read_tfvar() { grep -E "^[[:space:]]*$1[[:space:]]*=" terraform.tfvars 2>/dev/null | grep -oE '"[^"]+"' | tr -d '"' | head -1; }

CID="${CITRIX_CUSTOMER_ID:-$(read_tfvar citrix_customer_id)}"
CLID="${CITRIX_CLIENT_ID:-$(read_tfvar citrix_client_id)}"
CSEC="${CITRIX_CLIENT_SECRET:-$(read_tfvar citrix_client_secret)}"
BURL="${SPA_BASE_URL:-$(read_tfvar base_url)}"; BURL="${BURL:-https://api.cloud.com/accessSecurity}"
TURL="${SPA_TOKEN_URL:-$(read_tfvar token_url)}"; TURL="${TURL:-https://api.cloud.com}"

if [ -z "${CID}" ] || [ -z "${CLID}" ] || [ -z "${CSEC}" ]; then
  echo "cleanup: missing credentials (set CITRIX_CUSTOMER_ID/CLIENT_ID/CLIENT_SECRET or terraform.tfvars)" >&2
  exit 1
fi

TOK=$(curl -sS -X POST "${TURL}/cctrustoauth2/${CID}/tokens/clients" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=client_credentials" \
  --data-urlencode "client_id=${CLID}" \
  --data-urlencode "client_secret=${CSEC}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['access_token'])" 2>/dev/null || true)
# Fail closed: without a token every DELETE would 401 and the sweep would
# silently no-op, giving a false "clean" signal.
if [ -z "${TOK}" ]; then
  echo "cleanup: failed to acquire access token (auth/API error); aborting" >&2
  exit 1
fi
echo "cleanup: token acquired (len ${#TOK}); sweeping resources with prefix '${PREFIX}'"

export INSECURE="${INSECURE:-0}"
export BURL CID TOK PREFIX

ccurl() { if [ "${INSECURE:-0}" = "1" ]; then curl -k "$@"; else curl "$@"; fi; }
export -f ccurl

del_one() {
  local id="$1" path_id="$1"
  if [ "${ENCODE:-0}" = "1" ]; then
    path_id=$(python3 -c "import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1],safe=''))" "$id")
  fi
  ccurl -sS -X DELETE "${BURL}/${ENDPOINT}/${path_id}" \
    -H "Authorization: CWSAuth bearer=${TOK}" \
    -H "Citrix-CustomerId: ${CID}" \
    -o /dev/null -w "%{http_code} ${id}\n"
}
export -f del_one

# purge <endpoint> <id-field> [encode]
purge() {
  local endpoint="$1" id_field="$2" encode="${3:-0}" ids_file
  ids_file=$(mktemp)

  # Parse the list response defensively: the collection may be a bare JSON array,
  # a {"items":[...]} envelope, or a {"<name>":[...]} object — extract the first
  # list of objects, keep those whose name/fqdn starts with $PREFIX.
  # Fetch the list; on API failure emit a clear warning instead of a false
  # "0 to delete" (best-effort teardown continues with the other endpoints).
  local raw
  if ! raw=$(ccurl -sS -f "${BURL}/${endpoint}?limit=-1" \
      -H "Authorization: CWSAuth bearer=${TOK}" \
      -H "Citrix-CustomerId: ${CID}" \
      -H "Accept: application/json"); then
    echo "== ${endpoint}: WARNING: list request failed (API error) — NOT a clean signal; skipping" >&2
    rm -f "$ids_file"; return 0
  fi

  if ! printf '%s' "$raw" | ID_FIELD="${id_field}" python3 -c '
import sys, json, os
p = os.environ["PREFIX"]; f = os.environ["ID_FIELD"]
d = json.load(sys.stdin)
if isinstance(d, dict):
    items = d.get("items")
    if items is None:
        lists = [v for v in d.values() if isinstance(v, list)]
        items = lists[0] if lists else []
else:
    items = d
out = []
for a in items:
    if not isinstance(a, dict):
        continue
    name = str(a.get("name") or a.get("fqdn") or "")
    if name.startswith(p) and f in a:
        out.append(str(a[f]))
print("\n".join(out))
' > "$ids_file"; then
    echo "== ${endpoint}: WARNING: could not parse list response — NOT a clean signal; skipping" >&2
    rm -f "$ids_file"; return 0
  fi

  local n
  n=$(grep -c . "$ids_file" || true)
  echo "== ${endpoint}: ${n} '${PREFIX}' resources to delete"
  if [ "$n" -gt 0 ]; then
    export ENDPOINT="$endpoint" ENCODE="$encode"
    xargs -P 8 -I{} bash -c 'del_one "{}"' < "$ids_file" | sort | uniq -c | sort -rn
  fi
  rm -f "$ids_file"
}

# Delete in reverse dependency order. This is a best-effort backstop: the API
# rejects deletion of in-use routing domains (HTTP 400), so the reliable
# teardown is `terraform destroy`. Security groups use the SINGULAR endpoint.
purge "securityGroup"  "id"   0
purge "accessPolicy"   "id"   0
purge "applications"   "id"   0
purge "routingDomains" "fqdn" 1

echo "cleanup: done"
