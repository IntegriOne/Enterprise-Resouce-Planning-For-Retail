.PHONY: build test vet fmt run

build:
	GOWORK=off go build ./cmd ./cmd/healthcheck

test:
	GOWORK=off go test ./...

vet:
	GOWORK=off go vet ./...

fmt:
	go fmt ./...

run:
	GOWORK=off go run ./cmd
