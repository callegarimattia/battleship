build:
	go build -o bin/battleship ./cmd/server

run:
	go run ./cmd/server

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

docker-run:
	docker build -t battleship .
	docker run -p 8080:8080 battleship

all: fmt lint generate test 

.PHONY: build test lint fmt generate docker-build docker-run
