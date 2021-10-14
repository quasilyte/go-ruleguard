set -e

# Need GCC and musl to enable -race.
apk update && apk add git gcc musl-dev

export GO111MODULE=on

git clone https://github.com/quasilyte/go-ruleguard.git /root/go-ruleguard
cd /root/go-ruleguard

go test -race -v ./ruleguard/... ./analyzer/...
go build -race -o go-ruleguard ./cmd/ruleguard

./go-ruleguard -rules /root/rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

go get -v -u github.com/quasilyte/ruleguard-rules-test
./go-ruleguard -rules /root/rules2.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

./go-ruleguard -disable 'testrules/boolExprSimplify' -rules /root/rules2.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected3.txt

./go-ruleguard -enable 'testrules/boolExprSimplify' -rules /root/rules2.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected4.txt

./go-ruleguard -e 'm.Match(`$f($*_, ($x), $*_)`)' /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected5.txt

# Run inside GOROOT.
export GO111MODULE=off
cd /usr/local/go
go get -v -u github.com/quasilyte/go-ruleguard/dsl
/root/go-ruleguard/go-ruleguard -rules /root/rules.go ./src/encoding/... &> actual.txt || true
diff -u actual.txt /root/expected6.txt

echo SUCCESS
