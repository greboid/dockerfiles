#!/bin/sh
mkdir -p "${DESTDIR}/opt" "${DESTDIR}/usr/src"
cp -a /opt/venv "${DESTDIR}/opt/venv"
cp -a machine-learning/immich_ml "${DESTDIR}/usr/src/immich_ml"
cp -a machine-learning/scripts/healthcheck.py "${DESTDIR}/usr/src/healthcheck.py"
