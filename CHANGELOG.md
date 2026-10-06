# Changelog

All notable changes to the Citrix Secure Private Access (SPA) Terraform provider
are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!--
Maintainers: add new entries under "## [Unreleased]" as work merges. Before
tagging a release, rename that heading to "## [x.y.z] - YYYY-MM-DD". The release
workflow publishes the matching version section as the GitHub Release notes, so
keep the wording clear and user-facing.
-->

## [Unreleased]

### Added

- **`-ExtractLocals` flag on the resource-listing tool.** `./spa_manager.ps1 -List -ExtractLocals`
  (alias `-el`) hoists values that repeat throughout `spa_resources.tf` — resource-location names
  and UUIDs, and user/group directory tokens — into a `locals` block at the top of the file, and
  references them from the resources below. Updating a resource location or a user's tokens becomes
  a single edit instead of a find-and-replace across the whole file. The `user_metadata` local is
  derived from `users`, so a rule's `values` and its `metadata` display-name map can no longer drift
  apart. The rewrite is plan-neutral — the generated config yields the same `terraform plan` as
  without the flag — and anything the tool cannot reproduce exactly (for example
  `values = ["Everyone"]`, rules without `metadata`, or a resource-location UUID reported under two
  different names or in two different letter cases) is left inline, as is any resource location
  referenced only once, since hoisting it would buy no single edit point. Only the `locals`
  extraction itself is opt-in: the key-ordering and `${`/`%{` escaping fixes listed under **Fixed**
  below change the generated output on both paths, with or without the flag. See
  [`resource-listing-tool/README.md`](resource-listing-tool/README.md).

- **Pre-built application templates in [`templates/`](templates/README.md).**
  Ready-to-adapt Terraform configurations covering the **full SPA SaaS app
  catalog** — one template file per vendor (300+), grouped by SSO type under `templates/saml/` and
  `templates/form/`. Each template defines a `citrixspa_application` with
  `using_template = true` plus a `citrixspa_routing_domain` for every URL host
  it references — including placeholder hosts, so you replace each host in one
  place (wired together with `depends_on`), embeds the app icon and
  is annotated with the values you need to replace. SAML templates also carry
  the app's full set of SAML `custom_attributes` from the catalog — both the
  service-managed assertion mappings (for example `email = ns_user_email`) and
  admin-facing values (for example `UserID`) — so the generated app produces the
  same SAML assertion as picking the template in the SPA console, and any
  placeholder values are surfaced for you to complete.
  Routing-domain resource labels are namespaced per application (for example
  `rd_<app>_customer_fqdn`) so multiple templates can be combined into one
  configuration without colliding on Terraform resource names. Angle-bracketed
  placeholders (for example `<Customer FQDN>`) mark values that must be filled in
  before use.

- **SAML templates now preserve the `prefix_expr` flag on custom attributes.**
  When a catalog custom attribute set `prefix_expr`, the generated template now
  emits `prefix_expr = true`, so the app produces the same SAML assertion as
  selecting the template in the SPA console.

- **SAML templates now carry the catalog's relay state.** When a catalog entry
  defines a SAML relay state, the generated template now emits `sso.relay_state`
  (with the backend's surrounding quotes stripped; `NONE`/empty values are
  treated as unset), so the app produces the same SAML relay state as selecting
  the template in the SPA console.

- **`citrixspa_access_policy`: `TYPE_TAG` rules are now validated at plan time.**
  A rule with `type = "TYPE_TAG"` that omits (or sets to `""`) `tag_source` or
  `tag_key` now fails during `terraform plan` with a clear error pointing at the
  offending attribute, instead of being sent to the backend as an incomplete
  tag rule. Non-tag rule types are unaffected and may continue to omit these
  fields.

- **`citrixspa_routing_domain`: a routing domain deleted outside Terraform now
  explains itself.** Deleting an application also deletes the backend routing
  domains whose FQDN matches that application's `url`, `related_urls` or
  `destination` values. When a refresh finds a routing domain gone, the provider
  now emits a warning naming the FQDN, pointing at Note 3 in the
  `citrixspa_routing_domain` documentation, and describing how to resolve the
  drift — instead of silently treating the resource as gone and planning to
  recreate it with no explanation.

- **`citrixspa_application`: deleting an application now warns which routing
  domains may go with it.** The delete emits a warning listing the FQDNs derived
  from the application's `url`, `related_urls` and `destination` values, so the
  backend cleanup is visible before the next plan surfaces it as drift. The
  FQDNs are reduced the same way the service stores them — scheme, port, path
  and credentials removed, and lower-cased — so the warning names the routing
  domains that actually exist.

### Changed

- **Clarified the `using_template` / `template_name` documentation.** The
  `citrixspa_application` docs now explain that `using_template = true`
  provisions the application from the SPA service's built-in application catalog
  (maintained on the backend) and that any `sso`, `icon`, and `related_urls` you
  supply are applied on top of the catalog template, so the app can be created
  fully configured in a single apply — distinct from the reusable configuration
  files in the `templates/` directory.
- **`citrixspa_access_policy`: the `name` on an access rule
  (`access_rules[].name`) is now optional.** It was previously required, but the
  SPA service does not require a name for an access rule, so it can now be
  omitted. Existing configurations that set a name continue to work unchanged.

- **`citrixspa_security_group`: the `data_in` and `data_out` fields inside
  `system` and `unpublished_app` are now optional and default to `"disabled"`.**
  They were previously required, but the SPA service treats them as optional with
  a `"disabled"` default, so they can now be omitted. Existing configurations that
  set these values continue to work unchanged.

- **`citrixspa_security_group`: `data_in` and `data_out` now validate to
  `enabled` or `disabled`.** An explicit empty string (or any other value) is
  now rejected at plan time with a clear "must be one of" error, moving
  invalid-value rejection from apply time to plan time. Omitting the field still
  defaults to `disabled`.

- **`citrixspa_terminate_machine_access`: the `name` field is now optional.** It
  was previously required, but the SPA service identifies a machine by
  `account_name`, `object_id` and `idp_type` and does not require a name.

- **`citrixspa_terminate_machine_access`: the `object_id` and `idp_type` fields
  are now required (breaking).** The SPA service (and its public API) requires
  both to identify the machine, so the provider now enforces them at plan time
  instead of allowing an apply-time failure. Configurations that already set
  these fields are unaffected.

- **`citrixspa_terminate_machine_access`: changing any attribute now forces
  resource replacement.** The SPA service has no update API for machine
  termination records, so the provider now recreates the record (delete +
  create) when a value changes, instead of accepting the plan and leaving the
  service unchanged.

- **`citrixspa_terminate_user_access`: the `email`, `domain_name` and `duration`
  fields are now optional.** They were previously required, but the SPA service
  does not require them (a user is identified by `account_name`, `object_id` and
  `idp_type`). Existing configurations that set these values continue to work
  unchanged.

- **`citrixspa_routing_domain`: the valid `type` values now match the SPA
  service.** `external_via_connector` has been retired by the service and is no
  longer accepted; `internal_via_gateway` and `external_fixed_ip` are now
  supported. The valid values are `internal`, `external`, `conflicting`,
  `internal_bypass_proxy`, `internal_via_gateway`, and `external_fixed_ip`.
  Update any configuration still using `external_via_connector`.

- **SAML templates emit the custom-attribute `format` in canonical lowercase.**
  Values such as `Unspecified` are now written as `unspecified` to match the
  service's canonical form.

- **`citrixspa_access_policy`: clarified the rule attribute descriptions.** The
  `type`, `operator`, `tag_source`, `tag_key`, and `metadata` descriptions in
  the schema now spell out their valid values and when each applies, so the
  generated documentation and editor hovers stay in sync with the schema.

### Deprecated

- **`citrixspa_access_policy`: `access_rules[].conditions[].user_and_groups` is
  deprecated and is no longer sent to the SPA API.** The SPA service is retiring
  the field; user and group scope belongs in `access_rules[].rules[]` as an entry
  with `type = "TYPE_USERGROUP"`, which is what the SPA Console writes. Existing
  configurations keep working: the provider translates any `user_and_groups`
  entries into an equivalent `TYPE_USERGROUP` rule on write and collapses that
  synthesized rule again on read, so state still matches the configuration.
  `terraform plan` emits a deprecation warning showing the exact rule to paste
  into `rules[]`. The attribute will be removed in a future release — migrate now.

  Two behaviour notes for the translated rule. Its tokens now reach the SPA
  service's user/group validation: an `EMAIL:/` token must match a directory
  email claim and `all_users` requires the corresponding tenant feature, so
  either can fail the apply. Other token forms are stored verbatim, but a key
  that is not a `SID:/`, `OID:/` or `EMAIL:/` identity token (or `all_users`)
  cannot be resolved by the SPA Console and makes its policy edit page fail to
  render. `terraform plan` warns about both cases before the apply.

- **`resource-listing-tool/spa_manager.ps1` no longer generates `user_and_groups`
  in the HCL it writes.** Discovered and imported policies now express user and
  group scope purely as `TYPE_USERGROUP` rules.

### Removed

- **`citrixspa_access_policy` and `citrixspa_access_policies` data sources: the
  `access_rules[].conditions[].user_and_groups` attribute has been removed.** Read
  user and group scope from `access_rules[].rules[]` entries with
  `type = "TYPE_USERGROUP"` instead.

  One migration caveat for policies that predate the rules model. A policy whose
  user and group scope still lives only in the retired condition field is now
  read back without it, so configuration generated from these data sources — by
  `resource-listing-tool/spa_manager.ps1`, for example — will not contain that
  scope, and applying the generated configuration clears it on the service. This
  cannot loosen an `ACCESS_ALLOW` policy: the service evaluates the retired field
  as an alternative to the condition's platform filter rather than as an
  additional restriction, so dropping it can only make a condition match fewer
  requests, and for the `PLATFORM_FILTER_ANY` conditions these data sources emit
  it makes no difference at all. It does matter for an `ACCESS_DENY` policy, and
  for a condition whose platform filter is `PLATFORM_FILTER_PC` or
  `PLATFORM_FILTER_MOBILE`. Such policies can no longer be created — the service
  rejects an access rule that carries no machine or user group rule — and opening
  one in the SPA Console migrates it automatically. If you generate configuration
  from a tenant that still has them, move their scope into a `TYPE_USERGROUP`
  rule before you apply.

### Fixed

- **Resource-listing tool: `${` and `%{` in tenant data are no longer evaluated as Terraform
  templates.** A name or description containing `${...}` or `%{...}` — an application description,
  a group display name — was written into `spa_resources.tf` unescaped. The result is still valid
  HCL, so `terraform fmt` and `terraform validate` both pass, but Terraform interprets the text as
  an interpolation or directive at plan time instead of as literal text, silently changing the
  value or failing with an unhelpful parse error deep in the config. Both sigils are now escaped
  (`$${`, `%%{`) everywhere a tenant string is emitted — the literal list fallbacks used for rules
  `-ExtractLocals` cannot collapse, the access-rule and session-policy `type`/`operator`/
  `tag_source`/`tag_key` fields, application ids and platform filters, routing-domain overrides,
  security-group data-transfer settings, the browser mode, and every
  `citrixspa_terminate_machine_access` / `citrixspa_terminate_user_access` field — so any tenant
  string round-trips byte-for-byte.

- **Resource-listing tool: a missing `terraform` binary no longer looks like a generation failure.**
  `spa_manager.ps1` runs `terraform fmt` over the generated `spa_resources.tf` purely for
  cosmetics. When `terraform` was not on `PATH` the resulting error was reported as
  `Error writing spa_resources.tf` and generation returned failure, even though the file had
  already been written correctly. The tool now skips formatting with a warning and leaves the
  (still valid) unformatted HCL in place.

- **Resource-listing tool: `-List` now fails when it generates HCL that Terraform cannot parse.**
  `terraform fmt` rewrites formatting drift in place and still succeeds, so a non-zero exit from it
  means the generated `spa_resources.tf` is syntactically invalid, not merely untidy. That was
  previously reported and the run still finished successfully, leaving an unusable file to be
  discovered at the next `terraform plan`. `-List` now reports the parse error and exits with a
  failure. A missing `terraform` binary is still only a warning.

- **Resource-listing tool: generated map attributes no longer change key order
  between runs.** `spa_manager.ps1` built HCL maps such as a `TYPE_USERGROUP`
  rule's `metadata` from an unordered hashtable, so rediscovering an unchanged
  tenant could emit the same keys in a different order and show up as a spurious
  diff in version control. Key order now follows the order the API returns.

- **An unrelated API error no longer makes the provider believe a resource was
  deleted.** The provider decided that a resource had been removed outside
  Terraform by looking for the text `404` anywhere in the error message — which
  also contains the request's transaction ID and the raw service response. A
  failure that was not an HTTP 404 (a service error, an expired token, a
  validation error quoting one of your own values, such as an application URL
  like `error404.example.com`) could therefore be misread as a deletion: the
  resource was silently dropped from state and a recreate was planned. "Not
  found" is now determined from the HTTP status code, so any other failure is
  reported as an error and your state is left untouched. Genuine 404s are
  unchanged — a deleted resource is still detected as drift on refresh, and
  deleting an already-deleted resource still succeeds. Affects
  `citrixspa_application`, `citrixspa_access_policy`, `citrixspa_routing_domain`,
  `citrixspa_security_group` and `citrixspa_session_policy`.

- **`citrixspa_application`: a failed create can no longer adopt an unrelated
  existing application into your state.** When a create fails with an HTTP 500,
  the provider searches for an application the service may have partially
  created and, if it finds one matching the planned name and type, records it as
  the resource. That recovery was triggered by looking for the text `status 500`
  anywhere in the error message — which also quotes your own request back to
  you. A create rejected for an unrelated reason (for example an application URL
  containing the words `status 500`) could therefore start the search and bind a
  pre-existing, unrelated application to your configuration, so a later
  `terraform destroy` would delete it. Recovery is now triggered by the HTTP
  status code, so it runs only on a genuine service error. Recovery after a real
  HTTP 500 is unchanged.

- **`citrixspa_application`: the `icon` attribute now tolerates a data-URI
  prefix and wrapped base64.** An icon supplied with a `data:image/...;base64,`
  prefix or with the base64 payload split across lines is now normalized before
  it is sent to the SPA service and compared with semantic equality, so it
  applies cleanly and produces an empty plan on the next run instead of failing
  with "Provider produced inconsistent result after apply".

- **Application templates: icon data is now stored without whitespace, so every
  template applies cleanly.** A few catalog icons wrapped their base64 data
  across multiple lines. The SPA service stores the icon with whitespace
  stripped, so the value read back differed from the one sent and
  `terraform apply` failed with "Provider produced inconsistent result after
  apply". Template icons are now emitted as a single unbroken base64 string.

- **Application templates: icon base64 with missing padding is now
  canonicalized, so every template applies cleanly.** One catalog icon carried
  base64 whose length was not a multiple of four (the `=` padding was dropped).
  The SPA service pads and re-encodes such values, so the icon read back
  differed from the one sent and `terraform apply` failed with "Provider
  produced inconsistent result after apply". Template icons are now emitted in
  the same canonical form the service stores.

- **`citrixspa_application`: `template_name` no longer causes a perpetual diff
  when omitted.** For applications that do not set `template_name`, Terraform
  previously planned the attribute as "known after apply" on every run,
  producing a spurious in-place update. The attribute now retains its stored
  value, so an unchanged configuration yields an empty plan.

- **Application templates: descriptions are now sanitized to the values the SPA
  service accepts.** Vendor descriptions carried from the catalog could contain
  characters the application API rejects (for example `&`, `/`, `(`, `)`, and
  typographic quotes), causing `terraform apply` to fail with an
  "Application Description not valid" error. Descriptions now keep letters,
  digits, spaces, and basic punctuation (`&` is rendered as "and"), so every
  template applies without manual editing.

- **Application templates: routing-domain resource names are now unique within
  a template.** When an app referenced two hosts that reduced to the same
  identifier (for example `teams.microsoft.com` and `*.teams.microsoft.com`),
  the generator emitted two `citrixspa_routing_domain` blocks with the same
  resource name, which Terraform rejected as a duplicate. Colliding names are
  now suffixed to stay unique.

- **Application templates: corrected `related_urls` and removed bogus routing
  domains for 17 vendors.** For some catalog entries (for example Bugsnag,
  Buildkite, Duo) the generator emitted `related_urls = ["uniqueItems",
  "items", "default", "type"]` and a matching `citrixspa_routing_domain` for
  each of those junk values. Those templates now correctly produce
  `related_urls = ["<Customer FQDN>"]` with no spurious routing domains.

- **Application templates: cleaned up catalog text.** Vendor descriptions and
  URLs no longer contain literal `\u2019`/`\u00a0` escape sequences, missing
  spaces after sentence-ending periods, or trailing whitespace in URLs.

- **`citrixspa_application`: `template_name` no longer causes a perpetual
  diff.** The attribute is now computed, so when an application was created from
  a backend catalog template the `template_name` returned by the API matches
  state instead of showing a persistent change on every `plan`.

- **`citrixspa_application`: omitting `using_template` no longer causes a
  perpetual diff.** The attribute now defaults to `false` when unset (it is
  computed), so a configuration that leaves `using_template` out matches the
  value returned by the API on read instead of showing a persistent change on
  every `plan`.
- **`citrixspa_routing_domain`: creating a routing domain no longer fails when
  the service cannot read it back immediately.** The read that runs right after
  a successful create treated a "not found" response as an out-of-band deletion,
  discarded the new resource and failed the apply with `Missing Resource State
  After Create`, leaving a routing domain that existed in the service but not in
  the state. That read is now distinguished from a refresh: the created routing
  domain stays in state, and the provider emits a warning explaining that its
  computed values are refreshed by the next `terraform plan`.

- **`citrixspa_access_policy`: mixing explicit and omitted access-rule priorities
  no longer sends duplicate priority values to the backend.** A rule that omits
  `access_rules[].priority` previously defaulted to its 1-based list position,
  which could collide with a priority explicitly set on another rule in the same
  policy (for example an explicit `priority = 2` next to an omitted rule at list
  index 2, both sent as `2`). Omitted priorities are now assigned above the
  highest explicit priority in list order, so a filled-in priority never
  duplicates an explicit one. Policies where every rule omits its priority are
  unaffected (still numbered `1, 2, 3, …`), as are policies that set every
  priority explicitly.

- **`citrixspa_access_policy`: inserting or reordering access rules no longer
  reassigns rule identities or swaps `access_rules[].rules[].metadata` between
  rules, and no longer fails with "Provider produced inconsistent result after
  apply".** Backend-owned computed values (`access_rules[].id` and an omitted
  `access_rules[].rules[].metadata` map) were previously reconciled by list
  position. Adding a rule at the front (or reordering rules) shifted every
  following rule to a new index, so each existing rule inherited the id and
  metadata of whatever rule now sat at its position: existing rules had their
  backend ids reassigned, one rule's display-name/token mapping was associated
  with another (which could break the policy's edit page in the SPA Console),
  and the genuinely new rule was planned as `null` while the apply returned a
  real id and an empty map (`.access_rules[N].id: was null, but now ...`;
  `.rules[0].metadata: was null, but now cty.MapValEmpty`). Rules are now matched
  to their prior state by content and name rather than by index, so each existing
  rule keeps its own id and metadata across an insert or reorder, only a
  genuinely new rule's id and omitted metadata are planned as "known after
  apply", and re-applying an unchanged configuration stays idempotent.
  Configurations that set `id` or `metadata` explicitly are unaffected. Rules
  whose `values` reference an unresolved resource output (unknown at plan time)
  are matched by their unique `name` when reordered, and are never matched by
  position, so their backend ids are not swapped either. Named rules that are
  reordered and content-edited in the same change now keep their ids via their
  unique name instead of falling back to position, and unnamed rules with any
  unknown signature field (for example `active`, `access`, or a condition wired
  to an unresolved output) are likewise never reconciled by position. An omitted
  (Console-owned) rule `name`/`priority` is now preserved across an ordinary
  content edit: it is recovered from the prior rule matched by its reconciled id
  rather than by content signature, so editing a rule no longer clears its name
  or resets its priority. The content signature also includes the rule
  `description`, so two rules that differ only by their description stay
  identity-stable when reordered. Rule matching is now collision-safe: rules that
  differ only in how their `values`, `metadata`, `user_and_groups`, or
  enhanced-security maps are split (for example `["a,b"]` versus `["a","b"]`, or
  `{"a" = "b,c=d"}` versus `{"a" = "b", "c" = "d"}`) no longer produce the same
  internal signature, so their ids/name/priority can no longer be swapped on
  reorder. When an entire `access_rules` list is unknown at plan time (any rule
  field wired to an unresolved output), existing rules now keep their backend
  `id` and their `access_rules[].rules[].metadata` on the next apply — recovered
  from prior state by content, or by their unique `name` when the rule's own
  `values` are the unresolved output — instead of the backend assigning fresh
  ids and the metadata being dropped. This holds even when such rules are
  reordered in the same change: a reordered rule whose value is unknown at plan
  time keeps its own metadata (matched by name and position within the prior
  rule) rather than losing it or inheriting another rule's. When a sub-rule
  supplies `metadata` explicitly, the prior-state value that is preserved for an
  omitted same-key sibling is now the remaining prior entry rather than the one
  the explicit sub-rule claimed, so reordering a set of same-key sub-rules where
  one keeps its value and one omits it no longer collapses both onto the same
  metadata. Inserting a new, differently named rule alongside an existing rule
  whose name is omitted but whose content is identical no longer lets the new
  rule inherit the existing rule's backend `id`. An explicitly configured
  `access_rules[].id` is now authoritative: rules are matched to prior state by
  that id first, so reordering and editing two id-pinned rules together no longer
  lets the positional fallback pair a rule with the wrong prior and adopt its
  name, priority, or metadata. A sub-rule whose `metadata` map is set in
  configuration but wired to a value that is only known after apply (so the whole
  map is unknown at plan time) now keeps the requested new value instead of
  silently reverting to the prior-state metadata, which previously also caused a
  "Provider produced inconsistent result after apply" error; only a truly omitted
  (null) `metadata` map still adopts prior state. A rule whose `id` is pinned in
  configuration no longer lets a second rule with the same `name` and content
  re-claim that already-matched prior, so a backend rule id is never sent twice
  in one update. Renaming and reordering two rules in the same change no longer
  swaps their ids or metadata: a renamed rule that no longer matches prior state
  by name is now recovered by its content before the positional fallback runs.
  An `access_rules[].id` that is set in configuration but wired to a value only
  known after apply (unknown at plan time) now keeps the requested value instead
  of reverting to the prior rule's id at its old list position, which previously
  also produced a "Provider produced inconsistent result after apply" error.
  Reordering a set of same-key sub-rules where one keeps its `metadata`, one
  changes it to a new value, and one omits it no longer lets the omitted sibling
  adopt the changed sibling's old value. Two otherwise-identical rules that carry
  distinct, explicitly-configured `metadata` are now matched to prior state by a
  metadata-aware signature, so reordering them no longer swaps their ids or
  metadata.

- **`citrixspa_access_policy`: reordering two otherwise-identical access rules
  whose distinguishing `metadata` is unknown at plan time no longer fails with
  "Provider produced inconsistent final plan".** When a rule's `metadata` is
  wired to a value that is only known after apply (for example another resource's
  output), the metadata-aware match that separates two otherwise-identical rules
  can only run at apply, so at plan time their `access_rules[].id` values were
  pinned by list position and then re-assigned by metadata at apply — swapping
  between the plan and the final plan. The provider now defers the `id` of such
  an ambiguous rule to "known after apply" whenever the metadata-independent
  content and name it shares matches more than one candidate — a sibling in the
  same plan (a reorder) or a prior-state rule being deleted while this one
  survives — so apply assigns the metadata-correct id without a plan/apply
  mismatch. Rules with known or omitted `metadata`, and rules that are unique by
  content or name, are unaffected and keep their ids planned as before.

- **No more spurious drift on Console-owned fields `citrixspa_application.using_template`
  and `citrixspa_access_policy.priority` (including `access_rules[].priority`).**
  These fields are owned/normalized by the SPA Console on save, so they are now
  `Optional + Computed`. Omitting them from configuration no longer fails with
  "Provider produced inconsistent result after apply" (e.g. `.priority: was
  null, but now ...` or `.using_template: was null, but now cty.False`), and a
  Console-side edit that does not change their meaning no longer produces a
  phantom `terraform plan` diff. The policy-level `priority` retains its prior
  value across plans; the nested `access_rules[].priority` and `access_rules[].name`
  are now preserved from prior state as well by matching each rule to its
  prior-state counterpart by content, so an unrelated policy update no longer
  clears a rule name or resets a rule priority to its list position (matching
  stays correct even after the rules are reordered). Configurations that set
  these fields explicitly continue to work unchanged.

- **`citrixspa_access_policy`: an omitted `priority` is now assigned by the SPA
  service instead of being forced to `0`.** Previously, leaving `priority` unset
  sent a literal `0`, which suppressed the service's automatic priority
  assignment and could make multiple policies collide on the same value. The
  request now omits the field entirely when it is not set, so the service
  assigns the next available priority; an explicit value (including `0`) is still
  sent verbatim.
- **`citrixspa_access_policy`: an omitted `access_rules[].priority` is no longer
  serialized as `0`.** Every rule with an unset priority previously serialized as
  `0`, which suppressed the service's ordering; omitted rules now receive a
  distinct, order-reflecting priority, assigned above the highest explicit
  priority in the policy (so all-omitted rules become `1, 2, 3, …`).
- **`citrixspa_application`: orphaned-application recovery is no longer blocked
  by the Console-owned `using_template` field.** During recovery after a partial
  create, a candidate whose `using_template` was set by the Console is no longer
  rejected when the configuration omitted that field, preventing a spurious
  failure (and potential duplicate application). When `using_template` is set
  explicitly in configuration, it is still honored as a discriminator so a
  contradicting candidate is not adopted.

- **`citrixspa_access_policy`: setting any `enhanced_security_settings` key no
  longer drifts forever with "Provider produced inconsistent result after apply
  ... new element '_browserV1' has appeared".** When a restrictions block carries
  any enhanced-security setting, the backend injects a `_browserV1` key into its
  response even though the configuration never set it. The provider now drops
  such backend-injected keys from state (unless the configuration set them
  explicitly), so applies succeed and `terraform plan` is idempotent.

- **`citrixspa_access_policy`: setting the `enhanced_security_settings` key
  `insecure_content_allowed_for_urls_v1` to its default `"disabled"` no longer
  fails with "Provider produced inconsistent result after apply ... element
  'insecure_content_allowed_for_urls_v1' has vanished".** The backend drops this
  key from its response at the default value (as it does for the other toggle
  keys), so the provider now treats `"disabled"` as its reconcilable default and
  restores it when the API omits it.

- **`citrixspa_access_policy`: rule reconciliation on read is now matched by
  rule identity for both conditions and restrictions.** If the backend ever
  returns access rules in a different order than configured, conditions
  (`user_and_groups` null-vs-empty preservation) and restrictions are now both
  reconciled against the same identity-matched prior rule instead of the rule at
  the same list position, and neither name-based nor positional fallback pairs
  two rules whose known IDs differ. This prevents stale values being applied to
  the wrong rule after an out-of-band change.

- **`citrixspa_access_policy`: `enhanced_security_settings` toggle values are now
  validated at plan time.** The keys `clipboardV1`, `downloadV1`, `printingV1`,
  `uploadV1`, `keyLoggingV1`, `screenCaptureV1`, `watermarkV1`, and
  `insecure_content_allowed_for_urls_v1` now reject any value other than
  `enabled` or `disabled` with a clear error, moving invalid-value rejection from
  apply time to plan time. Other keys are left unconstrained.

- **`citrixspa_access_policy`: setting `enhanced_security_settings` keys to their
  default value (e.g. `downloadV1`, `uploadV1`, `clipboardV1`, `printingV1` =
  `"enabled"`) no longer fails with "Provider produced inconsistent result after
  apply".** The API omits keys whose value equals the backend default from its
  create/read response, and omits the entire `restrictions` block when *every*
  setting is at its default. The provider now restores whatever the API dropped —
  including rebuilding the whole `restrictions` block when it is omitted — so
  `apply` succeeds and re-planning is idempotent. Restored keys with a known
  default are set to that default value (not the previous state value), so
  genuine out-of-band drift is still detected on refresh, and prior rules are
  matched by identity (ID/name) rather than list position. (After a fresh
  `terraform import` of such a policy, the first plan may show a one-time diff
  that re-adds the default-valued keys, which self-heals on the next apply.)
  This also covers a `restrictions` block that only sets `redirect_sbs` (with no
  `enhanced_security_settings`): the omitted block is rebuilt while preserving a
  null `enhanced_security_settings` map, so it no longer vanishes on apply.

- **`citrixspa_access_policy`: a `restrictions` block that sets a non-default
  `redirect_sbs = true` (with no `enhanced_security_settings`) can now be
  created.** The API keeps the block but returns no `enhanced_security_settings`
  keys; the provider previously forced an empty map, failing with "Provider
  produced inconsistent result after apply ... `enhanced_security_settings`: was
  null, but now `cty.MapValEmpty`". The null map is now preserved.

- **`citrixspa_access_policy`: out-of-band removal of a non-default `restrictions`
  block is now detected as drift instead of being masked.** The API omits the
  `restrictions` block both when every setting is at its default *and* when the
  block is deleted entirely; reconstruction from prior state is now gated on the
  prior block being semantically all-default, so removing a non-default block
  (e.g. `redirect_sbs = true`) surfaces on refresh and is re-applied rather than
  silently reconstructed. The all-default check now uses the empirically
  confirmed backend defaults for every `enhanced_security_settings` key the API
  can omit (adding `keyLoggingV1` and `screenCaptureV1` = `"enabled"`), and
  treats any remaining unrecognized key as non-default so its drift is surfaced
  rather than assumed to be at default.

- **`citrixspa_access_policy`: importing a policy whose `restrictions` block the
  API still returns (e.g. `redirect_sbs = true`, or any non-default
  `enhanced_security_settings` key) but with no `enhanced_security_settings` keys
  no longer shows a spurious diff.** On a fresh import there is no prior state to
  reconcile against, so an omitted `enhanced_security_settings` field is now
  recorded as `null` (matching an unset optional map) instead of an empty `{}`,
  which Terraform treated as different. Note this does not apply to an all-default
  block (only `redirect_sbs = false` and no `enhanced_security_settings`): the API
  omits that block entirely, so a fresh import records no `restrictions` block and
  the first plan shows the same one-time, self-healing diff described above for
  default-valued policies.

- **`citrixspa_access_policy`: out-of-band clearing of a non-default
  `proxyTrafficV1` (or any `enhanced_security_settings` key with no reconcilable
  backend default) is now detected as drift instead of being masked.** Such a key
  is only ever omitted from the API response when its value was cleared out of
  band; the provider previously re-added the prior value, hiding the change. It is
  now left absent so Terraform surfaces the drift and re-applies the configured
  value.

- **`citrixspa_access_policy`: default `enhanced_security_settings` are now
  restored to the correct rule when a policy has two access rules that share the
  same name.** Rule names are not required to be unique, and on create the
  planned rules carry unknown IDs, so rules were matched to the API response by
  name; two identically-named rules both matched the first one, which could
  restore one rule's omitted default settings onto the other. Each prior rule is
  now claimed at most once, so identically-named rules are paired one-to-one in
  order.

- **`citrixspa_terminate_machine_access` and `citrixspa_terminate_user_access`:
  a create no longer fails if the new record is not yet visible in the list API
  immediately after creation.** The post-create re-read now falls back to the
  create response on a lookup miss, and user-configured optional values are
  preserved rather than overwritten by the re-read.

- **`citrixspa_terminate_machine_access` and `citrixspa_terminate_user_access`
  no longer fail with "Provider produced inconsistent result after apply".**
  The create API only echoes back the `id` and `object_id`, so fields such as
  `account_name`, `idp_type` and `duration` were being reset to empty values in
  state right after creation. The provider now re-reads the full record after
  creating it, so state matches the configuration and server-assigned defaults
  (for example the revocation `duration`) are populated correctly.

- **Omitting a newly-optional computed field no longer clears its server value
  on update.** For `citrixspa_security_group` (`system`/`unpublished_app`
  `data_in` and `data_out`) and `citrixspa_terminate_user_access` (`email`,
  `domain_name`, `duration`), the provider now preserves the prior state value
  when the field is left unset in the configuration, so updating other
  attributes no longer wipes or resets these values on the SPA service.

- **Token cache disk writes are now atomic**, and the provider skips disk
  caching (instead of writing to a relative path under the current module
  directory) when the home directory can't be resolved.

- **`citrixspa_access_policy`: an access rule's `restrictions.redirect_sbs` now
  defaults to `false` when omitted.** The attribute is now `Optional` +
  `Computed` with a static `false` default, so a `restrictions` block that sets
  only `enhanced_security_settings` (leaving `redirect_sbs` unset) plans and
  applies cleanly instead of failing with "Provider produced inconsistent result
  after apply" (planned `null`, but the backend/read returns `false`).

- **`citrixspa_access_policy`: an access rule's `restrictions.redirect_sbs` can
  now be set back to `false`.** The value was previously omitted from the
  request body when `false`, so the backend retained its prior `true` and
  `apply` failed with "Provider produced inconsistent result after apply"
  (`was cty.False, but now cty.True`). It is now always sent, so toggling
  `redirect_sbs` off takes effect.

- **`citrixspa_access_policy`: plan-time validation no longer errors when
  `access_rules` (or a nested `rules` block) is computed from another resource
  or an unresolved expression.** The `TYPE_TAG` validation previously failed
  with a value-conversion error when any of those collections was still unknown
  at plan time; it now defers to apply in that case, so referencing computed
  values in `access_rules` plans cleanly.

- **`citrixspa_application`: omitting a previously-set `description` or
  `category` now retains the existing value, and an explicit empty string is
  rejected at plan time.** The backend enforces a minimum length on both fields,
  rejects an empty value with `400 ... is too short`, and retains the prior
  value when the field is omitted — so a set `description` or `category` cannot
  be cleared once configured. Omitting either field now keeps the prior value
  (no drift, no `apply` error), and setting `description = ""` or
  `category = ""` now fails during `terraform plan` with a clear
  "string length must be at least 1" error instead of an apply-time
  `400 ... is too short` or an inconsistent-result error. `url` is required and
  continues to be sent on every request.

- **`citrixspa_application`: creating an application without a `category` no
  longer fails with `400 ... category is too short`.** `category` was being
  sent as an empty string even when unset; because the backend enforces a
  minimum length, an empty value is now omitted from the request body instead
  of sent.

- **`citrixspa_application`: creating an application with an explicit
  `sso = { type = "nosso" }` no longer fails with "Provider produced
  inconsistent result after apply" (`.sso: was cty.ObjectVal(...), but now
  null`).** For some application types (notably `web`, whose SSO defaults to
  `nosso`), the backend does not echo the configured SSO back on the read that
  immediately follows a create or update, so the provider previously stored
  `null` and contradicted the plan. The configured SSO is now preserved when
  the post-create or post-update read omits it, and any `custom_attributes` on
  a preserved SSO are retained (previously they could be dropped). On update,
  known server-computed SAML fields (`saml_sso_login_url`,
  `saml_cert_issuer_name`, `customer`) are also preserved instead of being reset
  to `null`, which previously could trigger the same inconsistent-result error.
  This most commonly affected configuration generated by discovering an existing
  tenant and re-applying it to another.

- **`citrixspa_application`: `sso` is no longer read back as `null` by
  `terraform import` or by a refresh that runs in the brief window right after
  the application was created.** The backend can transiently omit the SSO
  object on the read that immediately follows a write, so an import or refresh
  in that window previously recorded a configured SSO as `null` and reported
  spurious drift. The provider now re-reads a bounded number of times until the
  SSO object is returned; applications that genuinely have no SSO are
  unaffected.

- **`citrixspa_application`: boolean fields can now be set to `false`.** Setting
  `agentless_access`, `hidden`, `mobile_security`, or `sbs_only_launch` to
  `false` on an application whose backend value is currently `true` is now
  correctly persisted. Previously the `false` value was dropped from the request
  body, so the backend kept the old `true` and `apply` failed with
  `produced an unexpected new value: ... was cty.False, but now cty.True`.

- **`citrixspa_access_policy`: `null` and `""` are now treated as equivalent for
  optional strings.** Configuration produced by
  `terraform plan -generate-config-out` (which emits `null` for
  `description`, `access_rules[].description`, `rules[].tag_source`, and
  `rules[].tag_key`) now applies cleanly. Previously the provider returned `""`
  for these unset fields after apply, causing "Provider produced inconsistent
  result after apply" (`was null, but now cty.StringVal("")`) and a permanent
  phantom diff. Omitting one of these fields (or setting it back to `null`) now
  also clears any previously-set value to `""` instead of retaining the old
  value.

- **`citrixspa_access_policy` now reconciles an emptied `description`.** Setting
  `description = ""` (or reverting out-of-band drift back to an empty value) is
  now correctly applied; previously an empty description was dropped from the
  update request, so the change was never persisted and `apply` failed with
  `produced an unexpected new value: .description`.

- **`citrixspa_access_policy`: applying an access rule condition with an empty
  `user_and_groups = {}` no longer fails with "Provider produced inconsistent
  result after apply" (`was cty.MapValEmpty(cty.String), but now null`).** An
  empty `user_and_groups` map is now preserved as written instead of being
  silently converted to `null`, so `create` and `update` of policies that use
  empty user/group conditions succeed. (After a fresh `terraform import` the
  first plan may still show a one-time `{}`-vs-`null` diff, which self-heals on
  the next apply.)

- **Resource-discovery tool (`spa_manager.ps1`): access-policy conditions no
  longer produce a spurious diff after import.** Generated configuration for an
  access policy's `conditions` block previously always emitted
  `user_and_groups = {}`, which the provider reports as `null` when empty, causing
  a needless in-place update on the first `terraform plan` after importing the
  generated resources. The tool now only emits condition fields that are actually
  set, so discovered access policies with conditions round-trip cleanly.

### Security

- **Token cache key now includes the client secret.** The on-disk
  service-principal token cache (`~/.terraform.d/spa-cache/`) previously derived
  its AES-256-GCM key only from the customer ID and client ID, so it could be
  decrypted by anyone who knew those public values. The key now also uses the
  client secret. Existing cache files are re-created on the next authentication.

- **Redacted the PKCS#12 certificate and password from provider debug logging
  to prevent secrets from being written to logs.** The provider now replaces the
  values of the `certificate` and `certificatePassword` fields (matched by exact
  JSON field name) with `[REDACTED]` markers in the request debug log — the only
  path that carries them — while preserving all other fields for troubleshooting.

## [1.1.0]

### Added

- **Automatic handling of Citrix Cloud API rate limits.** Requests that are
  throttled by the API (HTTP 429) are now retried automatically, honoring the
  server's `Retry-After` hint with a capped backoff, and the number of
  concurrent in-flight changes is bounded to stay within the API's limits. Large
  `plan` and `apply` operations are noticeably more reliable under load.

### Changed

- **BREAKING — resources and data sources renamed from `spa_*` to `citrixspa_*`.**
  Every resource and data source type is now prefixed `citrixspa_` (for example
  `spa_application` → `citrixspa_application`), and the provider's local name is
  now `citrixspa`, matching the published registry name `citrix/citrixspa`.
  Existing configurations must be updated and their state migrated — see the
  [Upgrading to 1.1.0 guide](docs/guides/upgrading-to-1.1.0.md).
- **Quality — automated end-to-end testing.** An end-to-end test suite
  (reference-configuration drift check, resource-discovery round-trip,
  tenant-to-tenant migration, and unit/acceptance tests) now runs in CI before
  every release. This is an internal quality improvement with no change to
  provider behavior or configuration.

### Fixed

- **More resilient service-principal authentication.** The OAuth2 token endpoint
  is now retried on transient failures (HTTP 429 and 5xx) with backoff, so a
  brief rate limit or server blip while acquiring a token no longer fails the
  whole operation. Invalid-credential responses (401/403) still fail fast.

## [1.0.1] - 2026-07-07

- First public release on the Terraform Registry as
  [`citrix/citrixspa`](https://registry.terraform.io/providers/citrix/citrixspa).
