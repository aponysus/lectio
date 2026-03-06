APP_NAME := lectio

.PHONY: dev dev-api dev-web test migrate-up migrate-status build-api build-web

dev:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-api & \
	$(MAKE) --no-print-directory dev-web & \
	wait

dev-api:
	go run ./cmd/lectio serve

dev-web:
	cd web && npm run dev -- --host 0.0.0.0

test:
	go test ./...

migrate-up:
	go run ./cmd/lectio migrate up

migrate-status:
	go run ./cmd/lectio migrate status

build-api:
	go build ./cmd/lectio

build-web:
	cd web && npm run build
