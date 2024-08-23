set -e

release=$1

apk update && apk add git

export GO111MODULE=on

mkdir /root/test
cd /root/test
cp /root/rules.go rules.go

rm -rf $GOPATH
unset GOPATH

wget "https://github.com/quasilyte/go-ruleguard/releases/download/$release/ruleguard-linux-amd64.zip"
unzip ruleguard-linux-amd64.zip

go mod init test
go get -v -u github.com/quasilyte/go-ruleguard/dsl@master
go get -v -u github.com/quasilyte/ruleguard-rules-test@master
go get -v -u github.com/quasilyte/ruleguard-rules-test/sub2@master

./ruleguard -rules rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

./ruleguard -e 'm.Match(`$f($*_, ($x), $*_)`)' /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

echo SUCCESS
