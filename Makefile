APP_NAME := lectio
GO := GOTOOLCHAIN=go1.25.5 go

.PHONY: dev dev-api dev-web test migrate-up migrate-status build-api build-web

dev:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-api & \
	$(MAKE) --no-print-directory dev-web & \
	wait

dev-api:
	$(GO) run ./cmd/lectio serve

dev-web:
	cd web && npm run dev -- --host 0.0.0.0

test:
	$(GO) test ./...

migrate-up:
	$(GO) run ./cmd/lectio migrate up

migrate-status:
	$(GO) run ./cmd/lectio migrate status

build-api:
	$(GO) build ./cmd/lectio

build-web:
	cd web && npm run build
