# apko patch set at v1.4.4

Applied series: numbered unified diffs against the v1.4.4 tag, applied in
lexical filename order with `git apply` by `tooling/tools.sh`. Build fails
loudly if any patch fails to apply.

| patch | audited sites | change |
|---|---|---|
| `001-image-attribution.patch` | Default layer history comment empty (title+vendor annotation override kept); history `Author`/`CreatedBy` removed; image config author → `Greboid`; update `image_test.go` fixtures. |
| `002-etc-apko-config.patch` | Remove the `WriteEtcApkoConfig` call and function plus the now-unused `encoding/json` import: `/etc/apko.json` no longer enters the rootfs. |
| `003-accounts-gecos.patch` | apko-managed users get an empty GECOS field instead of `Account created by apko`. |
| `004-sbom-canonical-package.patch` | New `pkg/sbom/canonical` package: stdlib-only RFC 8785 canonical JSON + deterministic g5d.dev SPDX namespace scheme, with official RFC 8785 test vectors as Go tests. Byte-identical (modulo package clause) with the melange patch copy. |
| `005-sbom-g5d-identity.patch` | `Organization: Greboid` in index and per-arch image doc creators; both documents get `https://g5d.dev/containers/sbom/spdx/images/<sha256 of RFC 8785 canonical JSON of the doc minus documentNamespace>` computed just before rendering; regenerate generator SBOM goldens. |
| `006-cli-test-policy-and-goldens.patch` | Replace golden-OCI-layout digest comparisons in `internal/cli/build_test.go` with direct output-policy assertions (config author, unbranded history, no `etc/apko.json` in layers); drop `/etc/apko.json` from `publish_test.go` required files and bump its two intentional-digest pins; regenerate the `internal/cli` SBOM goldens (image digests changed deliberately). Upstream OCI golden layouts are left in tree unreferenced to keep rebases small. |
| `007-index-head-405-fallback.patch` | Forgejo Alpine endpoint compatibility: `pkg/apk/apk/index.go` — the etag-probe HEAD on remote indexes now falls back to the ordinary GET path on 405/501 only (Forgejo answers `405` + `allow: GET, DELETE`); every other non-200 stays a hard error, and the fallback GET keeps the configured authenticator, signature verification and decompressed-size bound (it reuses `fetchAndParse`). `pkg/apk/apk/cache.go`: `cacheTransport.head` no longer caches a 405/501 HEAD response as if it were an etag probe result; it discovers the etag with one authenticated GET (body discarded unread) so the etag/disk cache keeps working. New `index_head405_test.go` covers 405+501 fallback, retained failures (403/404/500), auth on the fallback (positive + negative), RSA256-signed/unsigned/wrong-key index handling, and both cache paths (etag discovery counts, no repeated 405s, failing GET propagates). |

Every step compiles and the upstream test suites pass after the series.
Rebase notes: the SBOM goldens and digest pins listed above are
content-derived and must be regenerated/re-bumped with the rebased code.
