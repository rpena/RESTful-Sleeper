.PHONY: run test build compose-up compose-down

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build ./cmd/api

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down
