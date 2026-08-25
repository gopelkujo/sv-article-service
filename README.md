# sv-article-service

Production-oriented Go REST API for the Article Management service.

Articles are stored in a MySQL `article` database (`posts` table) and exposed through a clean-architecture microservice with CRUD endpoints, centralized validation, structured logging, and graceful shutdown.

## Architecture

```
                 ┌─────────────────────────────────────────┐
                 │              cmd/api (main)             │
                 │   load config · wire deps · HTTP server │
                 └───────────────────┬─────────────────────┘
                                     │
                 ┌───────────────────▼─────────────────────┐
                 │         handler (transport)             │
                 │   chi router · JSON envelope · status   │
                 └───────────────────┬─────────────────────┘
                                     │
                 ┌───────────────────▼─────────────────────┐
                 │         service (use cases)             │
                 │   orchestration · calls validator       │
                 └───────────────────┬─────────────────────┘
                                     │
              ┌──────────────────────┼──────────────────────┐
              │                      │                      │
   ┌──────────▼──────────┐ ┌─────────▼─────────┐ ┌──────────▼──────────┐
   │      validator      │ │  domain (entity + │ │ repository/mysql    │
   │  field-level rules  │ │   repo interface) │ │  SQL implementation │
   └─────────────────────┘ └───────────────────┘ └──────────┬──────────┘
                                                            │
                                                   ┌────────▼────────┐
                                                   │  MySQL (article)│
                                                   └─────────────────┘
```

Request flow: **Handler → Service → Repository**, with `context.Context` propagated end-to-end. Handlers never talk to the database.

## Prerequisites

| Tool | Version / notes |
|------|-----------------|
| Go | 1.24+ |
| MySQL | 8.0 (via Docker Compose or local) |
| Docker / Compose | Optional but recommended for MySQL |
| [golang-migrate](https://github.com/golang-migrate/migrate) CLI | Must be built **with MySQL driver** |
| [golangci-lint](https://golangci-lint.run/) | For `make lint` |
| Postman / Insomnia | Optional, for API exploration |

### Install migrate CLI (MySQL tag required)

```bash
make tools
# or:
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

## Quick start (< 5 minutes)

```bash
# 1. Clone and enter the repo
git clone <repo-url> sv-article-service
cd sv-article-service

# 2. Local env
cp .env.example .env

# 3. Start MySQL
docker compose up -d mysql
# wait until healthy, then:

# 4. Install migrate (once) and apply schema
make tools
make migrate-up

# 5. Run the API
make run
```

API listens on `http://127.0.0.1:8080` by default (`APP_PORT` in `.env`).

> **Port conflict:** if something else (e.g. nginx) already binds `8080`, set `APP_PORT=8081` in `.env`.

### One-command stack (MySQL + API)

```bash
cp .env.example .env
docker compose up -d --build
```

The API container uses `DB_HOST=mysql` automatically.

## Migrations

Schema lives in `migrations/` and is applied with golang-migrate — do **not** create tables manually.

```bash
make migrate-up      # apply all pending
make migrate-down    # roll back one step
make migrate-create name=add_index_on_status
```

The first migration creates `posts` with columns: `id`, `title`, `content`, `category`, `created_date`, `updated_date`, `status`.

## Tests

```bash
make test                 # unit tests (DB-backed tests skipped)
make test-integration     # includes live MySQL HTTP + repository CRUD
```

Integration tests require a reachable MySQL matching `.env` and set `RUN_DB_TESTS=1`.

## Lint

```bash
make lint                 # golangci-lint run ./...
```

Config: `.golangci.yml`.

## Postman

1. Import `postman/article-service.postman_collection.json`
2. Set collection variable `base_url` (default `http://127.0.0.1:8080`)
3. Run **Create Article** first — it stores `article_id` for follow-up requests

Details: [`postman/README.md`](postman/README.md).

## API summary

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (MySQL ping) |
| `POST` | `/article/` | Create article |
| `GET` | `/article/{limit}/{offset}` | List articles (paginated) |
| `GET` | `/article/{id}` | Get article by id |
| `PUT` / `PATCH` | `/article/{id}` | Update article |
| `DELETE` | `/article/{id}` | Delete article |

Full request/response shapes, status codes, and validation rules: **[`docs/API.md`](docs/API.md)**.

### Example create

```bash
curl -s -X POST "http://127.0.0.1:8080/article/" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "A complete guide to clean architecture",
    "content": "This article explains how to structure a Go microservice using clean architecture. Handlers stay thin, services own business rules, and repositories isolate persistence. Validation is centralized so create and update share the same rules for title, content, category, and status. The content field must be at least two hundred characters long.",
    "category": "tech",
    "status": "draft"
  }'
```

## Project layout

```
cmd/api/                  # process entrypoint (wiring only)
internal/
  config/                 # env loading
  domain/                 # entities + repository interface
  repository/mysql/       # MySQL adapter
  service/                # use cases
  handler/                # HTTP transport
  validator/              # shared validation rules
  middleware/             # request-id, logging, recover
migrations/               # golang-migrate SQL
postman/                  # Postman collection
docs/API.md               # endpoint documentation
```

## Makefile targets

| Target | Description |
|--------|-------------|
| `make run` | Start API |
| `make build` | Build `./bin/api` |
| `make test` | Unit tests |
| `make test-integration` | Unit + MySQL integration tests |
| `make lint` | golangci-lint |
| `make migrate-up` / `migrate-down` | Schema migrations |
| `make tools` | Install migrate CLI with MySQL driver |
| `make docker-up` | Start MySQL only |
| `make docker-up-all` | Start MySQL + API containers |

