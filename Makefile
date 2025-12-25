test:
	go test -v -race -cover ./...
lint:
	golangci-lint run
fmt:
	golangci-lint fmt

.PHONY: test lint fmt
