set -e

release=$1

apk update && apk add git

export GO111MODULE=on

mkdir /root/test
cd /root/test
cp /root/rules.go rules.go

rm -rf $GOPATH
unset GOPATH

GOBIN=/root/test go install github.com/quasilyte/go-ruleguard/cmd/ruleguard@$release

go mod init test
go get -v -u github.com/quasilyte/go-ruleguard/dsl@master
go get -v -u github.com/quasilyte/ruleguard-rules-test@master
go get -v -u github.com/quasilyte/ruleguard-rules-test/sub2@master

./ruleguard -rules rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

./ruleguard -e 'm.Match(`$f($*_, ($x), $*_)`)' /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

echo SUCCESS
