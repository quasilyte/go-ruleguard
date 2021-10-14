set -e

release=$1

apk update && apk add git

export GO111MODULE=off

cd /root

wget "https://github.com/quasilyte/go-ruleguard/releases/download/$release/ruleguard-linux-amd64.zip"
unzip ruleguard-linux-amd64.zip

go get -v -u github.com/quasilyte/go-ruleguard/dsl
go get -v -u github.com/quasilyte/ruleguard-rules-test/...

./ruleguard -rules /root/rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

./ruleguard -e 'm.Match(`$f($*_, ($x), $*_)`)' /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

CGO_ENABLED=0 ./ruleguard -rules /root/rules.go /usr/local/go/src/encoding/... &> actual.txt || true
diff -u actual.txt /root/expected3.txt

echo SUCCESS
