# Include .env file
include .env
export $(shell sed 's/=.*//' .env)

# Define the Go script filename for package replacement
SCRIPT = pkg/module_name_changer/main.go

# Define the new module name variable for package replacement
MODULE_NAME ?= pkg/package/name

# Commands
run:
	@go run cmd/server/main.go

swagger:
	@swag init -g ./cmd/server/main.go

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) status

up: 
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) up

down:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) down

delete-all:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) reset

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) reset
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_CONNECTION) goose -dir=$(MIGRATION_PATH) up

seed:
	@go run migrations/seed/main.go

# Package replacement commands
.PHONY: change-mod-name

# Run the Go script for package replacement with the new package name
change-mod-name:
	@go run $(SCRIPT) $(MODULE_NAME)

# make change-mod-name MODULE_NAME=new/package/name

reset-seed:
	@make delete-all
	@make up
	@make seed