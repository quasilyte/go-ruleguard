GOPATH_DIR=`go env GOPATH`
RELEASE=v0.3.7
BUILD_COMMIT=`git rev-parse HEAD`

build:
	mkdir -p bin
	go build -o bin/ruleguard -ldflags "-X 'github.com/quasilyte/go-ruleguard/analyzer.Version=$(BUILD_COMMIT)'" ./cmd/ruleguard

build-release:
	mkdir -p bin
	go build -o bin/ruleguard -ldflags "-X 'github.com/quasilyte/go-ruleguard/analyzer.Version=$(RELEASE)'" ./cmd/ruleguard

test:
	go test -count 3 -coverpkg=./... -coverprofile=coverage.txt -covermode=atomic -race ./...
	cd rules && go test -v .
	@echo "everything is OK"

test-master:
	cd _test/install/gitclone && docker build --no-cache .
	cd _test/regress/issue103 && docker build --no-cache .
	cd _test/regress/issue193 && docker build --no-cache .
	@echo "everything is OK"

test-release:
	cd _test/install/binary_gopath && docker build --build-arg release=$(RELEASE) --no-cache .
	cd _test/install/binary_nogopath && docker build --build-arg release=$(RELEASE) --no-cache .
	@echo "everything is OK"

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_DIR)/bin v1.30.0
	$(GOPATH_DIR)/bin/golangci-lint run ./...
	go build -o go-ruleguard ./cmd/ruleguard
	./go-ruleguard -rules rules.go ./...
	@echo "everything is OK"

.PHONY: lint test test-master build build-release
