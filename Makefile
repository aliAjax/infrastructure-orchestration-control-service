.PHONY: build test run migrate lint clean

build:
	go build ./...

test:
	go test ./...

run:
	go run ./cmd/server

migrate:
	go run ./cmd/server -migrate-only=true

clean:
	rm -rf bin
