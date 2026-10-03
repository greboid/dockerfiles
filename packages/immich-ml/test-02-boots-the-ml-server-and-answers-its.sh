#!/bin/sh
# Offline: /ping is served before any model is loaded, and the cache
# folder stays empty. IMMICH_PORT keeps the guest clear of whatever the
# host is running - the bubblewrap test guest shares its network.
mkdir -p /tmp/mlcache
env \
  PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PYTHONPATH=/usr/src \
  PATH=/opt/venv/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  DEVICE=cpu MACHINE_LEARNING_CACHE_FOLDER=/tmp/mlcache \
  TRANSFORMERS_CACHE=/tmp/mlcache HF_HOME=/tmp/mlcache/hf-cache \
  MACHINE_LEARNING_MODEL_ARENA=false IMMICH_PORT=33003 \
  LD_PRELOAD=/usr/lib/libmimalloc.so.2 HOME=/home/nonroot \
  /opt/venv/bin/python -m immich_ml >/tmp/imml.log 2>&1 &
pid=$!
ok=""
for _ in $(seq 1 30); do
  code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:33003/ping || true)
  [ "$code" = 200 ] && { ok=yes; break; }
  sleep 1
done
kill "$pid" 2>/dev/null || true
[ "$ok" = yes ] || { cat /tmp/imml.log; exit 1; }
