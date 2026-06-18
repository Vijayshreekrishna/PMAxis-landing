.PHONY: run build clean test

SERVICES := discovery ingestion processor storage api

run:
	docker-compose -f deployments/docker-compose.yml up --build

build: build-services build-cli

build-services:
	go build ./services/...

build-cli:
	go build -o bin/pmaxis ./cmd/pmaxis/main.go

clean:
	docker-compose -f deployments/docker-compose.yml down -v
	find . -name "go.sum" -delete

test:
	go test ./...
