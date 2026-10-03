#!/bin/sh
# Run after: tooling/single-package.py go-licenses-builder
#            tooling/single-image.sh golang
#            docker load -i dist/images/golang.tar
# Usage: sh tooling/test-golang-image.sh [golang:local-amd64]
set -eu
image=${1:-golang:local-amd64}
test "$(docker image inspect -f '{{json .Config.Cmd}}' "$image")" = '["/bin/sh"]'
case "$(docker image inspect -f '{{json .Config.Entrypoint}}' "$image")" in
  null|'[]') ;;
  *) echo 'unexpected entrypoint' >&2; exit 1 ;;
esac
test "$(docker run --rm "$image" go version)" = 'go version go1.27.1 linux/amd64'
# No command: stdin is interpreted by the default shell CMD.
printf 'test "$(id -u)" = 0\n' | docker run --rm -i "$image"
docker run --rm -i "$image" /bin/sh -eux <<'EOF'
test "$(id -u)" = 0
test "$LANG" = en_US.UTF-8
test "$GOPATH" = /go
test "$GOFLAGS" = -buildvcs=false
test "$GOTELEMETRY" = off
test "$GOAMD64" = v1
test "$SOURCE_DATE_EPOCH" = 0
test "$(go env GOROOT)" = /usr/local/go
test "$(git config --system advice.detachedHead)" = false
for directory in /go /go/src /go/bin; do
  test "$(stat -c %a "$directory")" = 777
done
for command in gcc ld git patch apk go go-licenses; do command -v "$command"; done
test -s /usr/local/go/src/fmt/print.go
test -s /usr/local/go/LICENSE
test -s /usr/share/licenses/go-licenses-builder/LICENSE
test ! -e /usr/local/go/src/cmd
test ! -e /go/pkg/mod
test -w /etc/apk/repositories
grep -q 'https://dl-cdn.alpinelinux.org/alpine/v3.24/main' /etc/apk/repositories
find /etc/apk/keys -name '*.pub' | grep .
# The default repositories must work for downstream apk installs; local
# build-only dist/ repositories must not leak into this file.
! grep -q '^dist' /etc/apk/repositories
apk add --no-cache bash=5.3.9-r1
mkdir /tmp/smoke
cd /tmp/smoke
go mod init example.com/smoke
printf 'package main\nimport "fmt"\nfunc main(){fmt.Println("pure-go-ok")}\n' > main.go
CGO_ENABLED=0 go build -o pure .
test "$(./pure)" = pure-go-ok
cp /usr/local/go/LICENSE LICENSE
go-licenses report . > /tmp/licenses.csv
test "$(wc -l < /tmp/licenses.csv)" -eq 1
grep '^example.com/smoke,.*BSD-3-Clause$' /tmp/licenses.csv
mkdir /tmp/alternate-go
for item in /usr/local/go/*; do ln -s "$item" /tmp/alternate-go/; done
GOROOT=/tmp/alternate-go go-licenses report . > /tmp/alternate-licenses.csv
cmp /tmp/licenses.csv /tmp/alternate-licenses.csv
printf 'package main\n/* int answer(void) { return 42; } */\nimport "C"\nimport "fmt"\nfunc main(){fmt.Println(C.answer())}\n' > main.go
CGO_ENABLED=1 go build -o cgo .
test "$(./cgo)" = 42
EOF
# Test a real external Dockerfile consumer and COPY --from path contract.
context=$(mktemp -d)
smoke_image="golang-stage-smoke-$$:local"
trap 'rm -rf "$context"; docker image rm "$smoke_image" >/dev/null 2>&1 || true' EXIT HUP INT TERM
printf 'package main\nimport "fmt"\nfunc main(){fmt.Println("stage-ok")}\n' > "$context/main.go"
printf 'module example.com/stage\n\ngo 1.27.1\n' > "$context/go.mod"
cat > "$context/Dockerfile" <<'EOF'
ARG BUILDER
FROM ${BUILDER} AS build
WORKDIR /go/src/smoke
COPY . .
RUN CGO_ENABLED=0 go build -o /out/smoke .
FROM scratch
COPY --from=build /out/smoke /smoke
ENTRYPOINT ["/smoke"]
EOF
docker build --network none --pull=false --build-arg "BUILDER=$image" -t "$smoke_image" "$context"
test "$(docker run --rm --network none "$smoke_image")" = stage-ok
