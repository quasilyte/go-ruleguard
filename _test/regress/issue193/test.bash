set -e

cd /root

GO111MODULE=on go get -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@master
go get -v github.com/quasilyte/go-ruleguard/dsl

# Try running with different rules file order.

export CGO_ENABLED=0

ruleguard -rules log-rule.go,worker-rule.go,string-rule.go . > actual.txt 2>&1 || true
diff -u actual.txt /root/expected.txt

ruleguard -rules log-rule.go,string-rule.go,worker-rule.go . > actual.txt 2>&1 || true
diff -u actual.txt /root/expected.txt

ruleguard -rules string-rule.go,log-rule.go,worker-rule.go . > actual.txt 2>&1 || true
diff -u actual.txt /root/expected.txt

ruleguard -rules worker-rule.go,string-rule.go,log-rule.go . > actual.txt 2>&1 || true
diff -u actual.txt /root/expected.txt

echo SUCCESS
