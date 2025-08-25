include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONTY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONTY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DSN) up

.PHONTY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DSN) down $(filter-out $@,$(MAKECMDGOALS))