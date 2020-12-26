set -e

release=$1

apk update && apk add git

cd /root

wget "https://github.com/quasilyte/go-ruleguard/releases/download/$release/ruleguard-linux-amd64.zip"
unzip ruleguard-linux-amd64.zip

go get -v -u github.com/quasilyte/go-ruleguard/dsl

./ruleguard -rules /root/rules.go /root/target.go &> actual.txt || true
diff -u actual.txt /root/expected.txt

echo SUCCESS
