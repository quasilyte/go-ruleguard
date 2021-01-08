set -e

apk update && apk add git

export GO111MODULE=on

# Build go-ruleguard.
git clone https://github.com/quasilyte/go-ruleguard.git /root/go-ruleguard
cd /root/go-ruleguard
go build -o go-ruleguard ./cmd/ruleguard

cd /root
go mod init test

go mod tidy
cat go.mod | grep 'github.com/quasilyte/go-ruleguard/dsl'
echo 'OK: DSL is still in go.mod'

/root/go-ruleguard/go-ruleguard -rules /root/rules.go /root/src/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

/root/go-ruleguard/go-ruleguard -e 'm.Match(`$f($*_, ($x), $*_)`)' /root/src/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

echo SUCCESS
