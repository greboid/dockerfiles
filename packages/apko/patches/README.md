# apko patch set

The patches carry hand-written changes only. Data derived from build output
(SBOM goldens under `internal/cli/testdata/golden/sboms` and
`pkg/sbom/generator/spdx/testdata/expected_image_sboms`, and the two
`want := "sha256:..."` digest pins in `internal/cli/publish_test.go`) is not
patched, so upstream regenerating it does not break the patch set. The build
does not run the Go tests; `scripts/verify.sh` regenerates that data and runs
them.

```sh
scripts/verify.sh                       # apply to the tag in melange.yaml, regenerate, go test ./...
scripts/verify.sh -w DIR --continue     # after fixing a patch that failed to apply in DIR
scripts/verify.sh -w DIR --export       # rewrite patches/ from the commits in DIR
```

Setting `APKO_UPDATE_GOLDENS=1` makes the SBOM golden tests write their output
instead of comparing (added by 005 and 006).

| patch | audited sites | change |
|---|---|---|
| `001-image-attribution.patch` | Default layer history comment empty (title+vendor annotation override kept); history `Author`/`CreatedBy` removed; image config author → `Greboid`; update `image_test.go` fixtures. |
| `002-etc-apko-config.patch` | Remove the `WriteEtcApkoConfig` call and function plus the now-unused `encoding/json` import: `/etc/apko.json` no longer enters the rootfs. |
| `003-accounts-gecos.patch` | apko-managed users get an empty GECOS field instead of `Account created by apko`. |
| `004-sbom-canonical-package.patch` | New `pkg/sbom/canonical` package: stdlib-only RFC 8785 canonical JSON + deterministic g5d.dev SPDX namespace scheme, with official RFC 8785 test vectors as Go tests. Byte-identical (modulo package clause) with the melange patch copy. |
| `005-sbom-g5d-identity.patch` | `Organization: Greboid` in index and per-arch image doc creators; both documents get `https://g5d.dev/containers/sbom/spdx/images/<sha256 of RFC 8785 canonical JSON of the doc minus documentNamespace>` computed just before rendering; `APKO_UPDATE_GOLDENS` hook in `spdx_test.go`. |
| `006-cli-test-policy-and-goldens.patch` | Replace golden-OCI-layout digest comparisons in `internal/cli/build_test.go` with direct output-policy assertions (config author, unbranded history, no `etc/apko.json` in layers); drop `/etc/apko.json` from `publish_test.go` required files; `APKO_UPDATE_GOLDENS` hook for the `internal/cli` SBOM goldens. Upstream OCI golden layouts are left in tree unreferenced to keep rebases small. |
| `007-index-head-405-fallback.patch` | Forgejo Alpine endpoint compatibility: `pkg/apk/apk/index.go` — the etag-probe HEAD on remote indexes now falls back to the ordinary GET path on 405/501 only (Forgejo answers `405` + `allow: GET, DELETE`); every other non-200 stays a hard error, and the fallback GET keeps the configured authenticator, signature verification and decompressed-size bound (it reuses `fetchAndParse`). `pkg/apk/apk/cache.go`: `cacheTransport.head` no longer caches a 405/501 HEAD response as if it were an etag probe result; it discovers the etag with one authenticated GET (body discarded unread) so the etag/disk cache keeps working. New `index_head405_test.go` covers 405+501 fallback, retained failures (403/404/500), auth on the fallback (positive + negative), RSA256-signed/unsigned/wrong-key index handling, and both cache paths (etag discovery counts, no repeated 405s, failing GET propagates). |
