.PHONY: proto test build docker-build docker-up docker-down docker-logs

proto:
	cd contract && buf lint && buf generate

test:
	go fmt ./... && go vet ./... && go test -v -race -count=1 ./...

build:
	go build -o bin/iam-server ./cmd && go build -o bin/grpc-healthcheck ./cmd/healthcheck

docker-build:
	docker compose build

docker-up:
	docker compose up -d --wait

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f iam
