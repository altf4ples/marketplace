.DEFAULT_GOAL := help
-include .env

.PHONY: help up down down-v logs test test-race fmt vet build check compose-config migrate-up migrate-down smoke

help:
	@echo "make up             - build and start all services"
	@echo "make down           - stop services, preserve volumes"
	@echo "make down-v         - stop services and delete volumes"
	@echo "make logs           - follow API logs"
	@echo "make fmt            - format Go files"
	@echo "make test           - run unit tests"
	@echo "make test-race      - run tests with race detector"
	@echo "make vet            - run go vet"
	@echo "make build          - build API binary"
	@echo "make check          - fmt check, test, vet, build"
	@echo "make compose-config - validate Compose config"
	@echo "make smoke          - run end-to-end API smoke test (stack must be running)"

up:
	docker compose up --build -d

down:
	docker compose down

down-v:
	docker compose down -v

logs:
	docker compose logs -f api

fmt:
	gofmt -w $$(find cmd internal -name '*.go')

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -trimpath -o bin/api ./cmd/api

check:
	@test -z "$$(gofmt -l $$(find cmd internal -name '*.go'))" || (gofmt -l $$(find cmd internal -name '*.go'); exit 1)
	go test -race -cover ./...
	go vet ./...
	CGO_ENABLED=0 go build ./...

compose-config:
	docker compose config

smoke:
	python3 scripts/smoke.py

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate -path=/migrations -database="$(DATABASE_URL)" down 1
