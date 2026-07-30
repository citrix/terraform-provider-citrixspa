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
