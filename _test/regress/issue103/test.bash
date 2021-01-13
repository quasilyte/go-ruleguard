set -e

apk update && apk add git

export GO111MODULE=on

git clone https://github.com/quasilyte/go-ruleguard.git /root/go-ruleguard &&
    cd /root/go-ruleguard && go build -o /root/ruleguard ./cmd/ruleguard && cd /root

go mod init test

go get -v -u github.com/quasilyte/go-ruleguard/dsl
go get -v -u github.com/sirupsen/logrus

./ruleguard -c 0 -rules /root/rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

./ruleguard -c 0 -rules /root/rules2.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected2.txt

echo SUCCESS
