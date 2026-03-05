APP ?= lectio
GOTOOLCHAIN ?= auto
DB_PATH ?= ./devdata/lectio.db
MIGRATIONS_DIR ?= migrations
MIGRATE_BIN ?= migrate
DATABASE_URL ?= sqlite3://$(abspath $(DB_PATH))

.PHONY: dev dev-api dev-web test migrate-up migrate-status

dev:
	@trap 'kill 0' INT TERM; \
	$(MAKE) dev-api & \
	$(MAKE) dev-web & \
	wait

dev-api:
	@mkdir -p devdata
	@GOTOOLCHAIN=$(GOTOOLCHAIN) go run ./cmd/lectio

dev-web:
	@cd web && npm run dev -- --host --port 5173

test:
	@GOTOOLCHAIN=$(GOTOOLCHAIN) go test ./...

migrate-up:
	@mkdir -p $(dir $(DB_PATH))
	@$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-status:
	@$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version
