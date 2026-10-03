#!/bin/sh
cd machine-learning
# The six heavyweight deps are Alpine packages (system site); the rest
# of the lock export installs into the venv, with onnx and ml-dtypes
# built from source against the system protobuf/numpy.
#
# onnx defaults to statically linking protobuf (ONNX_USE_PROTOBUF_SHARED_LIBS
# is OFF, so CMake's FindProtobuf runs with Protobuf_USE_STATIC_LIBS=ON),
# but Alpine's protobuf-dev ships no libprotobuf.a - only the shared
# library and protoc. And since protoc is present, onnx skips its
# bundled-protobuf fallback and configure dies on the missing
# protobuf::libprotobuf target. Link the system shared protobuf instead;
# the package pins protobuf=31.1-r1 at runtime. Scikit-build-core reads
# CMAKE_ARGS from the environment, which - unlike pip's --config-settings
# - also reaches packages installed from a requirements file.
export CMAKE_ARGS="-DONNX_USE_PROTOBUF_SHARED_LIBS=ON"
uv export --frozen --no-dev --extra cpu --no-emit-project --no-hashes \
  --format requirements-txt -o /tmp/all-reqs.txt
grep -vE '^(onnxruntime|opencv-python|opencv-python-headless|numpy|shapely|watchfiles|websockets)==' \
  /tmp/all-reqs.txt > /tmp/reqs.txt
/opt/venv/bin/pip install --no-cache-dir -r /tmp/reqs.txt \
  --no-binary onnx --no-binary ml-dtypes
