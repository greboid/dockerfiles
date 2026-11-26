# runcage

melange package for [runcage](https://git.mouse-lake.ts.net/greboid/runcage), an OCI runtime wrapper for Docker that decrypts SOPS documents into per-container tmpfs secrets. Installs `/usr/bin/runcage` and the `docker.service` drop-in importing the `age-key` credential.

The runtime is not registered with Docker by this package; add it to `/etc/docker/daemon.json`:

```json
{ "runtimes": { "runcage": { "path": "/usr/bin/runcage" } } }
```
