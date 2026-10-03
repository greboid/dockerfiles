#!/bin/sh
export PATH="/go/bin:/usr/local/go/bin:$PATH"
test -x /go/bin/go-licenses.bin
test -s /usr/share/licenses/go-licenses-builder/LICENSE
mkdir -p /tmp/license-smoke
cd /tmp/license-smoke
go mod init example.com/license-smoke
printf 'package main\nimport "fmt"\nfunc main(){fmt.Println("ok")}\n' > main.go
cp /usr/local/go/LICENSE LICENSE
go-licenses report . > /tmp/licenses.csv
test "$(wc -l < /tmp/licenses.csv)" -eq 1
grep '^example.com/license-smoke,.*BSD-3-Clause$' /tmp/licenses.csv
# Exercise a different stdlib source prefix without network/toolchain downloads.
mkdir -p /tmp/alternate-go
for item in /usr/local/go/*; do ln -s "$item" /tmp/alternate-go/; done
GOROOT=/tmp/alternate-go go-licenses report . > /tmp/alternate-licenses.csv
cmp /tmp/licenses.csv /tmp/alternate-licenses.csv
