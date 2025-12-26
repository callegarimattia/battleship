test:
	go test -v -race -coverprofile=coverage.out ./...
	grep -v "/mocks/" coverage.out > coverage.final.out
	go tool cover -func=coverage.final.out
	rm coverage.out coverage.final.out
lint:
	golangci-lint run
fmt:
	golangci-lint fmt
generate:
	mockery

.PHONY: test lint fmt generate
