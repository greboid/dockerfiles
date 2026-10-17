# secret-env package

Reusable static Rust launcher at `/usr/bin/secret-env`, with no runtime package
dependencies. Uses the existing `packages/common` workflow; no applications are
wrapped automatically.

This initial package carries a source snapshot because the new upstream project
has no published release/tag yet. From `../../secret-env`, regenerate it with
`python3 scripts/stage-package.py`. Review and commit both repositories together;
`source/SHA256SUMS` is checked by the build. Do not edit the snapshot directly.
Once upstream is published, replace the snapshot with the normal `git-checkout`
pipeline using a real release tag and verified expected commit. No fictional
upstream URL or commit is used here.

Build with the organisation's patched melange (including `test/paths`):

```sh
melange keygen local.rsa
chmod 600 local.rsa
melange build melange.yaml --arch x86_64 --signing-key local.rsa --out-dir dist
melange test melange.yaml --arch x86_64 \
  --repository-append "$PWD/dist" --keyring-append "$PWD/local.rsa.pub"
```

Keys and build outputs are local only. Build-stage Python/C tools exercise the
black-box tests and do not enter the runtime APK. Package tests use a shell only
inside the test guest, not as a launcher dependency. Bump epoch for rebuilds of
the same version, following the other package repositories.

See the upstream project's `docs/packaging.md` for explicit-mode config packages
and the optional native-name wrapper pipeline. Repository creation and publishing
must be done from the host; nothing has been pushed.
