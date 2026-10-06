# SPA Application Templates

Ready-to-adapt Terraform templates for Citrix Secure Private Access (SPA)
applications, generated from the SPA SaaS app catalog. There is one template
file per vendor, so you can use these as starting points for
`citrixspa_application` resources instead of writing each configuration from
scratch.

Files are grouped by SSO type:

- [`saml/`](saml/) — SAML SSO applications (one file per vendor).
- [`form/`](form/) — form-fill SSO applications.

> **These are starting points, not apply-as-is configurations.** Every template
> contains angle-bracketed `<placeholder>` values (tenant FQDNs, org URLs,
> customer IDs — for example `<Customer FQDN>`, `<customer-domain>`,
> `<customer_id>`) that you **must** replace before running `terraform apply`.
> SAML apps also need a signing certificate configured on the tenant, which is
> not part of a template.

## What's in each file

Each template defines, for one vendor:

- a **`citrixspa_application`** with `using_template = true`, the matching
  `template_name`, and a complete `sso` block; and
- one **`citrixspa_routing_domain`** for every URL host the app references (the
  app `url` host and each `related_urls` entry, including any that are still
  `<placeholder>` values), with the application `depends_on` those routing
  domains.

`using_template = true` provisions the app from the SPA service's built-in
catalog; the `sso`, `icon`, and `related_urls` you supply are applied on top of
it, so a single `terraform apply` produces a working single-sign-on app.

## How to use a template

Assuming you already have a working Terraform project (provider configured):

1. **Pick a template**, for example [`saml/salesforce.tf`](saml/salesforce.tf).
2. **Copy the whole file** — both the `citrixspa_application` block and its
   `citrixspa_routing_domain` blocks — into your own configuration.
3. **Replace every `<placeholder>`** with your tenant values (typically `url`,
   `audience`, `assertion_url`, and the `related_urls` entries). The SSO
   settings and icon are already filled in from the catalog.
4. **Replace placeholder routing-domain hosts.** Each URL host already has a
   matching `citrixspa_routing_domain` and `depends_on` entry; where a routing
   domain's `fqdn` is still a `<placeholder>`, replace it with the real host
   (matching the value you set in `related_urls`).
5. **Apply:** `terraform plan` to review, then `terraform apply`.

> **Combining several templates?** Routing-domain resource labels are namespaced
> per application (for example `rd_<app>_customer_fqdn`), so copying multiple
> templates into one configuration does not collide on Terraform resource names.
> Two apps can still resolve to the *same* host once you replace the placeholders
> (a shared CDN or Microsoft domain, or the common `<Customer FQDN>`); the SPA
> service rejects two routing domains with the same `fqdn`, so in that case keep
> a single `citrixspa_routing_domain` for that host and point each application's
> `depends_on` at it.

## Field notes

- **`saml_type`** — `SP`, `IDP`, or `SP_IDP`, matching how the vendor initiates SSO.
- **`name_id_source`** — one of `email`, `upn`, `name`, `guid_b64`, `sam`.
- **`name_id_format`** — one of `unspecified`, `emailAddress`, `persistent`,
  `transient`, `WindowsDomainQualifiedName`, `X509SubjectName`.
- **`sign_assertion`** — one of `ASSERTION`, `BOTH`, `NONE`, `RESPONSE`.
- **`state`** — templates set `complete` so the app is published on first apply.
  Change it to `incomplete` to create the app as a draft.
- **`related_urls`** — must not include `https://` or trailing slashes.
- **`icon`** — a raw base64-encoded image (PNG or JPEG), with no `data:...;base64,` prefix.

For the full attribute reference, see the
[`citrixspa_application` resource docs](../docs/resources/application.md).
