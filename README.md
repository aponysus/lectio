# Lectio

Scaffold for the Lectio MVP described in `plan_v2.md`.

## Local Development

1. Copy `.env.example` to `.env` and set secrets.
2. Start API + web:

```bash
make dev
```

Or start each service separately:

```bash
make dev-api
make dev-web
```

## Testing

```bash
make test
```

If your local Go install has a patch-version toolchain mismatch, override the toolchain explicitly:

```bash
GOTOOLCHAIN=go1.25.5 make test
```

## Migrations

The scaffold expects `golang-migrate` CLI on `PATH`.

```bash
make migrate-status
make migrate-up
```
