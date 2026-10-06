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
  local residual=0
  if [ "$n" -gt 0 ]; then
    export ENDPOINT="$endpoint" ENCODE="$encode"
    local results_file
    results_file=$(mktemp)
    # del_one prints "<http_code> <id>"; keep the raw codes to check residuals.
    xargs -P 8 -I{} bash -c 'del_one "{}"' < "$ids_file" | tee "$results_file" | sort | uniq -c | sort -rn
    # A DELETE is clean only on 2xx (deleted) or 404 (already gone); curl prints
    # 000 on transport failure. Anything else means the resource may remain.
    residual=$(awk '$1 !~ /^2[0-9][0-9]$/ && $1 != "404" { c++ } END { print c+0 }' "$results_file")
    rm -f "$results_file"
  fi
  rm -f "$ids_file"
  if [ "$residual" -gt 0 ]; then
    echo "== ${endpoint}: NOT clean — ${residual} of ${n} target(s) failed to delete" >&2
    return 1
  fi
  return 0
}

# Routing domains cannot be deleted while enabled — the API returns HTTP 400
# ("notDisabled"). Mirror the provider's teardown: disable each matching domain
# with a PUT (using the exact Content-Type/nosniff headers the provider sends,
# otherwise the gateway replies 415) and then DELETE it.
purge_routing_domains() {
  local raw
  if ! raw=$(ccurl -sS -f "${BURL}/routingDomains?limit=-1" \
      -H "Authorization: CWSAuth bearer=${TOK}" \
      -H "Citrix-CustomerId: ${CID}" \
      -H "Accept: application/json"); then
    echo "== routingDomains: WARNING: list request failed (API error) — NOT a clean signal; skipping" >&2
    return 0
  fi

  printf '%s' "$raw" | ccurl_env python3 -c '
import sys, os, json, urllib.parse, urllib.request

prefix = os.environ["PREFIX"]
burl   = os.environ["BURL"]
cid    = os.environ["CID"]
tok    = os.environ["TOK"]
insecure = os.environ.get("INSECURE", "0") == "1"

d = json.load(sys.stdin)
if isinstance(d, dict):
    items = d.get("items")
    if items is None:
        lists = [v for v in d.values() if isinstance(v, list)]
        items = lists[0] if lists else []
else:
    items = d

# Match domains where some DNS label is exactly $PREFIX or begins with
# "$PREFIX-". Golden resources are named "{prefix}-{suffix}", so PREFIX "e2e"
# matches labels "e2e-golden-web..." (and "api.e2e-golden-web..."), but NOT an
# unrelated label that merely starts with the prefix chars (e.g. "e2etools").
# This anchors the destructive sweep at a real label boundary. Splitting on "."
# also handles a leading "*." and matches the prefix in any label position.
def matches_prefix(fqdn, prefix):
    return any(label == prefix or label.startswith(prefix + "-")
               for label in fqdn.split("."))

targets = [a for a in items if isinstance(a, dict)
           and matches_prefix(str(a.get("fqdn", "")), prefix)]
print("== routingDomains: %d %r resources to delete" % (len(targets), prefix))

ctx = None
if insecure:
    import ssl
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE

def req(method, fqdn, body=None):
    path = burl + "/routingDomains/" + urllib.parse.quote(fqdn, safe="")
    data = json.dumps(body).encode() if body is not None else None
    r = urllib.request.Request(path, data=data, method=method)
    r.add_header("Authorization", "CWSAuth bearer=" + tok)
    r.add_header("Citrix-CustomerId", cid)
    r.add_header("Accept", "application/json")
    if body is not None:
        r.add_header("Content-Type", "application/json; charset=utf-8")
        r.add_header("X-Content-Type-Options", "nosniff")
    try:
        with urllib.request.urlopen(r, context=ctx) as resp:
            return resp.status
    except urllib.error.HTTPError as e:
        return e.code
    except Exception as e:
        # Transport failure (DNS/TCP/TLS): return a string sentinel that ok()
        # treats as a residual, so it is never mistaken for a success status.
        return "ERROR: " + str(e)

def ok(status):
    # 2xx = deleted/disabled; 404 = already gone. Anything else (including the
    # transport-error string sentinel) means the resource may still be present.
    return isinstance(status, int) and (200 <= status < 300 or status == 404)

residual = 0
for a in targets:
    fqdn = a["fqdn"]
    if a.get("flag") != "disabled":
        body = {
            "fqdn":        fqdn,
            "type":        a.get("type", ""),
            "appType":     a.get("appType", ""),
            "comment":     a.get("comment", ""),
            "flag":        "disabled",
            "ip":          a.get("ip", False),
            "locationIds": a.get("locationIds", []),
        }
        ds = req("PUT", fqdn, body)
        print("  disable %s -> %s" % (fqdn, ds))
    xs = req("DELETE", fqdn)
    print("  delete  %s -> %s" % (fqdn, xs))
    if not ok(xs):
        residual += 1

if residual:
    print("== routingDomains: NOT clean — %d of %d target(s) failed to delete"
          % (residual, len(targets)), file=sys.stderr)
    sys.exit(1)
'
}

# ccurl_env: run a command with the credentials/config the python helper needs.
ccurl_env() { PREFIX="$PREFIX" BURL="$BURL" CID="$CID" TOK="$TOK" INSECURE="${INSECURE:-0}" "$@"; }

# Delete in reverse dependency order. Security groups and session policies use
# the SINGULAR endpoint. Every sweep fails loudly (non-zero) if a DELETE left a
# residual; run them all best-effort and aggregate, so one endpoint's orphan
# doesn't hide another's and the tenant is never reported clean while resources
# remain.
sweep_status=0
purge "securityGroup"  "id"   0 || sweep_status=1
purge "sessionPolicy"  "id"   0 || sweep_status=1
purge "accessPolicy"   "id"   0 || sweep_status=1
purge "applications"   "id"   0 || sweep_status=1
purge_routing_domains            || sweep_status=1

if [ "$sweep_status" -ne 0 ]; then
  echo "cleanup: NOT clean — one or more sweeps left residuals (see errors above)" >&2
  exit 1
fi

echo "cleanup: done"
