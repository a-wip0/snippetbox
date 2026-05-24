include .env
export

MIGRATIONS_DIR=./db/migrations
MIGRATE=migrate
DB_URL=mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

.PHONY: migrate-up migrate-down migrate-force migrate-create migrate-install

## Install golang-migrate CLI
migrate-install:
	curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
	sudo mv migrate /usr/local/bin/

## Run all up migrations
migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

## Rollback latest migration
migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

## Force migration version (useful if migration failed midway)
migrate-force:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force

## Create a new migration file (requires -name=...)
migrate-create:
ifndef name
	$(error name is required. Use like: make migrate-create name=create_users)
endif
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)