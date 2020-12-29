set -e

apk update && apk add git

export GO111MODULE=on

git clone https://github.com/quasilyte/go-ruleguard.git /root/go-ruleguard
cd /root/go-ruleguard

go test -v ./ruleguard/... ./analyzer/...
go build -o go-ruleguard ./cmd/ruleguard
./go-ruleguard -rules /root/rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

go get -v -u github.com/quasilyte/ruleguard-rules-test
./go-ruleguard -rules /root/rules2.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

echo SUCCESS
