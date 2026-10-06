#!/usr/bin/env python3
"""
DRAFT / EXPERIMENT — Customer A vs Customer B migration verifier (state diff).

Compares two Terraform state files (each produced by discovering a tenant with
spa_manager, so their resource addresses line up) and asserts that every
resource in A has a field-for-field equal twin in B — after:

  1. dropping tenant-specific / computed fields that legitimately differ
     (ids, timestamps, computed state), and
  2. remapping application-ID cross-references (access_policy.apps,
     security_group.app_ids) to application NAMES, since the IDs differ per
     tenant.

Usage:   compare-tenants.py A.tfstate B.tfstate
Exit 0 = A and B match (migration verified). Exit 1 = differences found.

Wired into the `migrate` E2E scenario (run-migrate-e2e.sh): after the migration,
both tenants are discovered into their own Terraform state and this comparator
gates the scenario's success. The SCRUB_KEYS / APP_ID_REFS lists below are what
may need extending as the golden config grows to new resource types/fields.
"""

import json
import sys
from collections import Counter


def suffix(rtype: str) -> str:
    """spa_application / citrixspa_application -> application (prefix-agnostic)."""
    return rtype.split("_", 1)[1] if "_" in rtype else rtype


# Resource suffix -> business key used to match a resource across tenants.
KEYS = {
    "application": "name",
    "routing_domain": "fqdn",
    "access_policy": "name",
    "security_group": "name",
    "session_policy": "name",
}

# Tenant-specific / computed keys dropped at ANY nesting depth before comparing.
# Observed against real discovered state:
#   id            -> per-tenant resource ids (top level AND nested, e.g.
#                    access_rules[].id)
#   customer      -> the tenant's own customer id embedded in application sso
#   created/modified -> timestamps
#   icon_url      -> per-tenant computed icon URL
#   state / policy_count / error -> computed / server-side
#   location_ids  -> per-tenant routing-location UUIDs
#   saml_sso_login_url / saml_cert_issuer_name -> server-computed SSO fields that
#                    embed the tenant's customer id and per-tenant APPID, so they
#                    legitimately differ between tenants (nested under sso).
SCRUB_KEYS = {
    "id",
    "customer",
    "created",
    "modified",
    "icon_url",
    "state",
    "policy_count",
    "error",
    "location_ids",
    "saml_sso_login_url",
    "saml_cert_issuer_name",
}

# Resource suffix -> fields holding application IDs, remapped to app names.
APP_ID_REFS = {
    "access_policy": ["apps"],
    "security_group": ["app_ids"],
}

# Resource suffix -> Terraform Set-typed fields. Sets serialize to JSON state as
# lists whose element order is NOT meaningful, so they are sorted before
# comparing to avoid false mismatches when A and B hold the same elements in a
# different order. (apps / app_ids are already sorted via APP_ID_REFS above.)
SET_FIELDS = {
    "application": ["related_urls", "keywords"],
}


def load(path):
    with open(path) as f:
        return json.load(f)


def app_id_to_name(state):
    """Build {application id -> name} for one tenant's state."""
    m = {}
    for r in state.get("resources", []):
        if suffix(r.get("type", "")) == "application":
            for inst in r.get("instances", []):
                a = inst.get("attributes", {})
                if a.get("id"):
                    m[a["id"]] = a.get("name")
    return m


def scrub(obj):
    """Recursively drop tenant-specific / computed keys at any nesting depth."""
    if isinstance(obj, dict):
        return {k: scrub(v) for k, v in obj.items() if k not in SCRUB_KEYS}
    if isinstance(obj, list):
        return [scrub(x) for x in obj]
    return obj


def normalize(sfx, attrs, id2name):
    a = dict(attrs)
    # Remap app-id cross-references to names BEFORE scrubbing ids.
    for field in APP_ID_REFS.get(sfx, ()):
        if isinstance(a.get(field), list):
            a[field] = sorted(id2name.get(x, f"<unknown-id:{x}>") for x in a[field])
    # Sort Set-typed fields: their state order is not meaningful.
    for field in SET_FIELDS.get(sfx, ()):
        if isinstance(a.get(field), list):
            a[field] = sorted(a[field], key=lambda x: json.dumps(x, sort_keys=True))
    return scrub(a)


def index(state):
    """Return ({(suffix, key): normalized_attrs}, problems).

    A resource with a missing business key, or a duplicate (suffix, key), is
    reported as a problem instead of being silently indexed as (suffix, None) or
    overwriting an earlier entry — either would hide a real mismatch and produce
    a false pass.
    """
    id2name = app_id_to_name(state)
    idx = {}
    problems = []
    for r in state.get("resources", []):
        sfx = suffix(r.get("type", ""))
        if sfx not in KEYS:
            continue
        for inst in r.get("instances", []):
            attrs = inst.get("attributes", {})
            key = attrs.get(KEYS[sfx])
            if key is None:
                problems.append(f"{sfx} resource has no '{KEYS[sfx]}' business key")
                continue
            if (sfx, key) in idx:
                problems.append(f"duplicate {sfx} with key '{key}'")
                continue
            idx[(sfx, key)] = normalize(sfx, attrs, id2name)
    return idx, problems


def main():
    if len(sys.argv) != 3:
        print("usage: compare-tenants.py A.tfstate B.tfstate", file=sys.stderr)
        sys.exit(2)

    A, problems = index(load(sys.argv[1]))
    B, pb = index(load(sys.argv[2]))
    problems = [f"[A] {p}" for p in problems] + [f"[B] {p}" for p in pb]

    # 1) per-type count parity
    ca = Counter(s for s, _ in A)
    cb = Counter(s for s, _ in B)
    for s in sorted(set(ca) | set(cb)):
        if ca[s] != cb[s]:
            problems.append(f"COUNT {s}: A={ca[s]} B={cb[s]}")

    # 2) every A resource must have an equal twin in B
    for (sfx, key), av in sorted(A.items()):
        bv = B.get((sfx, key))
        if bv is None:
            problems.append(f"MISSING in B: {sfx} '{key}'")
            continue
        if av != bv:
            lines = [
                f"    {f}: A={av.get(f)!r}  B={bv.get(f)!r}"
                for f in sorted(set(av) | set(bv))
                if av.get(f) != bv.get(f)
            ]
            problems.append(f"DIFF {sfx} '{key}':\n" + "\n".join(lines))

    # 3) anything extra in B
    for key in sorted(B):
        if key not in A:
            problems.append(f"EXTRA in B: {key[0]} '{key[1]}'")

    if problems:
        print("MIGRATION MISMATCH (A -> B):\n" + "\n".join(problems))
        sys.exit(1)
    print(f"OK — all {len(A)} comparable resources match A -> B.")
    sys.exit(0)


if __name__ == "__main__":
    main()
