set -e

apk update && apk add git

cd /root

GO111MODULE=on go get -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@master
go get -v github.com/quasilyte/go-ruleguard/dsl
go get -v github.com/quasilyte/uber-rules
go get -v github.com/delivery-club/delivery-club-rules@testBundleMerge
# Try running with different rules file order.

export CGO_ENABLED=0

ruleguard -rules rules.go . > actual.txt 2>&1 || true
diff -u actual.txt /root/expected.txt

echo SUCCESS
