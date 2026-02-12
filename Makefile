DB_DSN ?= postgres://ap_final:ap_final@127.0.0.1:5433/ap_final?sslmode=disable
MIGRATIONS_DIR := ./migrations

.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql