.PHONY: proto test vet build

proto:
	cd contract && buf lint && buf generate

test:
	go fmt ./... && go vet ./... && go test -v -race -count=1 ./...

vet:
	go vet ./...

build:
	go build -o bin/enterprise-foundation ./cmd
	go build -o bin/grpc-healthcheck ./cmd/healthcheck
