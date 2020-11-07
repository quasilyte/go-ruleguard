GOPATH_DIR=`go env GOPATH`

test:
	go test -count 3 -coverprofile=coverage.txt -covermode=atomic -race -v ./analyzer/...
	go test -count 3 -race -v ./ruleguard/...
	@echo "everything is OK"

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_DIR)/bin v1.30.0
	$(GOPATH_DIR)/bin/golangci-lint run ./analyzer/...
	$(GOPATH_DIR)/bin/golangci-lint run ./ruleguard/...
	@echo "everything is OK"

.PHONY: lint test
