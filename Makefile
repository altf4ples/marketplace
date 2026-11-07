.DEFAULT_GOAL := help

include .env

pg-up:
	@docker compose up -d postgres

pg-down:
	@docker compose down -v postgres

mg-create:
	@mkdir -p migrations
	@test -n "$(name)" || (echo "Укажите name=..." && exit 1)
	@docker run --rm \
		-u "$(shell id -u):$(shell id -g)" \
		-v "$(PWD)/migrations:/migrations" \
		migrate/migrate:v4.15.2 \
		create -ext sql -dir /migrations -seq $(name)

mg-up:
	@docker compose up migrate

mg-down:
	@docker compose run --rm \
		migrate \
		-path=/migrations \
		-database="$(DATABASE_URL)" \
		down -all

mg-status:
	@docker compose run --rm \
		migrate \
		-path=/migrations \
		-database="$(DATABASE_URL)" \
		version

help:
	@echo "pg-up          — поднять Postgres"
	@echo "pg-down        — остановить Postgres"
	@echo "migrate-create name=... — создать миграцию"
	@echo "migrate-up     — применить миграции"
	@echo "migrate-down   — откатить все миграции"
	@echo "migrate-status — текущая версия"

.PHONY: pg-up pg-down migrate-create migrate-up migrate-down migrate-status help