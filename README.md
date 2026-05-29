# pipefy-integration

Backend service that registers clients and synchronises them with [Pipefy](https://www.pipefy.com/) cards via a GraphQL API.

Built with **Go 1.26**, following **Clean Architecture**, **Hexagonal Architecture (Ports & Adapters)**, **DDD**, and **SOLID** principles.

---

## Architecture

```
cmd/api/            → Composition root (main.go)
internal/
  domain/           → Entities, Value Objects, Port interfaces (no external deps)
  usecase/          → Application logic (CreateClient, ProcessWebhook, ProcessOutbox)
  adapter/
    inbound/http/   → Gin handlers, DTOs, middleware, router
    outbound/
      persistence/  → GORM models, mappers, repository implementations
      pipefy/       → Pipefy GraphQL adapters (fake + real)
pkg/                → Reusable packages (config, logger, apperror, database)
```

### Outbox Pattern

The distributed write problem (local DB + Pipefy API) is solved with the **Transactional Outbox Pattern**:

1. `POST /api/v1/clients` → persists the `client` row **and** an `outbox_events` row atomically in a single transaction.  
2. A background worker polls `outbox_events` every **5 seconds**, calls the Pipefy GraphQL API, and updates `clients.pipefy_card_id` on success.  
3. Failed events are retried up to **3 times**, then permanently marked as failed.

---

## Requirements

- Go 1.26+
- Docker & Docker Compose
- A Pipefy Personal Access Token (staging/prod only)

---

## Getting Started

```bash
# 1. Clone the repository
git clone https://github.com/Casagrande-Lucas/pipefy-integration
cd pipefy-integration

# 2. Generate config.yaml interactively
make setup

# 3. Start PostgreSQL
make docker-up

# 4. Run the API (auto-migrates tables on startup)
make run
```

### Makefile targets

| Target | Description |
|---|---|
| `make setup` | Interactive config.yaml generator |
| `make docker-up` | Start PostgreSQL via Docker Compose |
| `make docker-down` | Stop and remove containers |
| `make run` | Build and run the API |
| `make test` | Run all tests |
| `make lint` | Run golangci-lint |
| `make build` | Compile binary to `bin/api` |

---

## Configuration (`config.yaml`)

```yaml
app:
  name: pipefy-integration
  port: 8080
  environment: dev     # dev | staging | prod

database:
  host: localhost
  port: 5433
  name: pipefy_integration
  user: pipefy
  password: pipefy
  sslmode: disable

pipefy:
  token: ""            # Required for staging and prod
  pipe_id: ""          # Target pipe ID
  simulate: false      # true → log payload only (staging); false → send to Pipefy (prod)

logger:
  level: info          # debug | info | warn | error
  format: json         # json | console
```

### Environments

| Environment | Pipefy adapter | Token required | Mutations sent |
|---|---|---|---|
| `dev` | FakeAdapter (logs only) | No | No |
| `staging` | RealAdapter + `simulate: true` | Yes | No |
| `prod` | RealAdapter + `simulate: false` | Yes | Yes |

---

## API

### `POST /api/v1/clients`

Registers a new client and enqueues a Pipefy `createCard` event.

**Request**
```json
{
  "cliente_nome": "Lucas Casagrande",
  "cliente_email": "lucas@example.com",
  "tipo_solicitacao": "consultoria",
  "valor_patrimonio": 250000
}
```

**Response `201 Created`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "nome": "Lucas Casagrande",
  "email": "lucas@example.com",
  "tipo_solicitacao": "consultoria",
  "valor_patrimonio": 250000,
  "prioridade": "prioridade_alta",
  "status": "aguardando_analise",
  "pipefy_card_id": "",
  "criado_em": "2026-05-29T00:00:00Z",
  "atualizado_em": "2026-05-29T00:00:00Z"
}
```

> `pipefy_card_id` is empty on creation and filled asynchronously by the outbox worker.

**Priority rules**

| `valor_patrimonio` | `prioridade` |
|---|---|
| ≥ 200 000 | `prioridade_alta` |
| < 200 000 | `prioridade_normal` |

---

### `POST /api/v1/webhook`

Receives a Pipefy `card.field.update` event and updates the client status.  
Processing is **idempotent** — duplicate events are silently ignored.

**Header (optional)**
```
X-Pipefy-Webhook-UUID: <uuid>
```

**Request**
```json
{
  "action": "card.field.update",
  "data": {
    "card": { "id": "card-42", "pipe_id": "pipe-1" },
    "field": { "field_id": "status", "new_value": "processado" }
  }
}
```

**Response `204 No Content`**

---

### `GET /health`

```json
{ "status": "ok" }
```

---

## Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/usecase/...
go test ./internal/adapter/inbound/http/handler/...
go test ./internal/domain/...
```

Tests use hand-written stubs and `net/http/httptest` — no external test framework or database required.

---

## Project Structure (detailed)

```
cmd/
  api/main.go                   Composition root
  setup/main.go                 Interactive config generator
internal/
  domain/
    entity/                     Client, OutboxEvent, ProcessedEvent
    valueobject/                Email, ID, Priority, Status, OutboxEventStatus
    port/repository/            Repository interfaces + Transactor[R]
    port/service/               PipefyService interface
  usecase/
    client/create_client.go     Atomic client + outbox event creation
    outbox/process_outbox.go    Background worker — dispatches Pipefy mutations
    webhook/process_webhook.go  Idempotent webhook processing
  adapter/
    inbound/http/
      dto/                      Request/response structs
      handler/                  ClientHandler, WebhookHandler
      middleware/               Logger (Zap), Recovery
      router.go                 Gin engine setup
    outbound/
      persistence/
        model/                  GORM models
        mapper/                 Domain ↔ model mapping
        repository/             GORM implementations
      pipefy/
        fake_adapter.go         Dev — logs, no HTTP
        real_adapter.go         Staging/prod — GraphQL mutations
pkg/
  apperror/                     Typed errors with HTTP status codes
  config/                       Viper config loader + validation
  database/                     GORM connect, migrate, GORMTransactor[R]
  logger/                       Zap factory
```
