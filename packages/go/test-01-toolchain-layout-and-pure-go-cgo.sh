#!/bin/sh
export PATH="/go/bin:/usr/local/go/bin:$PATH"
test "$(go version)" = 'go version go1.27.1 linux/amd64'
test "$(go env GOROOT)" = /usr/local/go
test -s /usr/local/go/src/fmt/print.go
test -s /usr/local/go/LICENSE
test -s /usr/share/licenses/go/LICENSE
test ! -e /usr/local/go/src/cmd
test "$(stat -c %a /go)" = 777
test "$(stat -c %a /go/bin)" = 777
test "$(stat -c %a /go/src)" = 777
mkdir -p /tmp/go-smoke
cd /tmp/go-smoke
go mod init example.com/smoke
printf 'package main\nimport "fmt"\nfunc main(){fmt.Println("pure-go-ok")}\n' > main.go
CGO_ENABLED=0 go build -o pure .
test "$(./pure)" = pure-go-ok
printf 'package main\n/* int answer(void) { return 42; } */\nimport "C"\nimport "fmt"\nfunc main(){fmt.Println(C.answer())}\n' > main.go
CGO_ENABLED=1 go build -o cgo .
test "$(./cgo)" = 42
