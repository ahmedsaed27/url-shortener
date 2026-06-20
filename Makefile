ifneq (,$(wildcard .env))
include .env
export
endif

APP_MAIN=./cmd/api
MIGRATE_CMD=migrate -path migrations -database "$(DATABASE_URL)"

.PHONY: run worker dev test tidy docker-up docker-down migrate-up migrate-down migrate-create

run:
	go run $(APP_MAIN)

worker:
	go run ./cmd/worker

dev:
	air

test:
	go test ./...

tidy:
	go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down

name ?= create_table_name
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
