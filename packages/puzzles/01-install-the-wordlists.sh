#!/bin/sh
# Static assets and templates are embedded in the binary; only the
# wordlist directory is external (legacy copies it to /app/wordlists).
install -d -m 0755 "${DESTDIR}/app/wordlists"
install -m 0644 wordlists/*.wl "${DESTDIR}/app/wordlists/"
