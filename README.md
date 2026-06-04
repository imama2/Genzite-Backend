# Genzite-Backend

AI-powered no-code website builder with Go backend and Python AI services.

## Architecture

- **Go Backend** (Port 8080): Main API server with Gin framework, JWT auth, PostgreSQL
- **Python gRPC Service** (Port 50051): AI Architect Agent for website generation
- **Python Worker**: Background job processor for build pipeline
- **PostgreSQL** (Port 5432): Main database
- **Redis** (Port 6379): Job queue and pub/sub
- **RabbitMQ** (Ports 5672/15672): Optional message queue

## Prerequisites

### For Local Development
- Go 1.23+
- Python 3.11+
- PostgreSQL 16
- Redis 7
- RabbitMQ 3.13 (optional)
- Protocol Buffers compiler (`protoc`)
- Poetry (Python dependency manager)

### For Docker
- Docker 20.10+
- Docker Compose 2.0+

## Setup

### 1. Generate Protobuf Files

Generate gRPC stubs for both Go and Python:

```bash
make proto
```

This creates:
- `gen/go/agents/*.pb.go` - Go gRPC stubs
- `python/grpc_server/generated/*.py` - Python gRPC stubs

### 2. Install Dependencies

**Go dependencies:**
```bash
go mod download
```

**Python dependencies:**
```bash
cd python
poetry install
```

### 3. Configure Environment

**Go service (.env):**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=genzite_db
DB_SSLMODE=disable

REDIS_URL=localhost:6379

JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_TOKEN_EXPIRY=86400

GRPC_ARCHITECT_URL=localhost:50051

MIDTRANS_SERVER_KEY=your-midtrans-key
```

**Python service (python/.env):**
```env
REDIS_HOST=localhost
REDIS_PORT=6379
GRPC_PORT=50051
LOG_LEVEL=INFO

OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
```

## Running the Backend

### Option 1: Local Development (Makefile)

**Terminal 1 - Start infrastructure:**
```bash
# Start PostgreSQL, Redis, RabbitMQ
docker compose up db redis rabbitmq -d
```

**Terminal 2 - Run Go backend:**
```bash
make run
```

**Terminal 3 - Run Python gRPC server:**
```bash
cd python
python -m grpc_server.server
```

**Terminal 4 - Run Python worker (optional):**
```bash
cd python
python workers/build_worker.py
```

**Run migrations:**
```bash
make migrate-cli
```

The Go backend will be available at http://localhost:8080

### Option 2: Full Docker Stack (Recommended)

Build and start all services:

```bash
docker compose up --build -d
```

This starts:
- ✅ `genzite-backend` - Go API server (http://localhost:8080)
- ✅ `genzite-python-grpc` - Python gRPC server (localhost:50051)
- ✅ `genzite-python-worker` - Background worker
- ✅ `genzite-postgres` - PostgreSQL database (localhost:5432)
- ✅ `genzite-redis` - Redis cache (localhost:6379)
- ✅ `genzite-rabbitmq` - RabbitMQ (http://localhost:15672, guest/guest)

**Check service status:**
```bash
docker compose ps
```

**View logs:**
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f app
docker compose logs -f python-grpc
docker compose logs -f python-worker
```

**Stop all services:**
```bash
docker compose down
```

**Stop and remove volumes:**
```bash
docker compose down -v
```

**Rebuild specific service:**
```bash
docker compose up --build app -d
```

## Testing the Backend

**Health check:**
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

**Register a user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
```

## Makefile Commands

```bash
make help                # Show all available commands
make run                 # Run server (air if available, else go run)
make build               # Build binary to ./bin/genzite-backend
make test                # Run unit tests
make test-coverage       # Generate coverage report
make proto               # Generate gRPC stubs
make docker-up           # Start full Docker stack
make docker-down         # Stop stack and remove volumes
make docker-rebuild      # Rebuild app container only
make docker-logs         # Follow app logs
make migrate-cli         # Run migrations via CLI
make fmt                 # Format code
make tidy                # Tidy go modules
make lint                # Run golangci-lint
```

## API Requests

Base URL: http://localhost:8080

### Health
- URL: http://localhost:8080/api/v1/health
- Request header: none
- Request body: none

### Auth
- URL: http://localhost:8080/api/v1/auth/register
- Request header: Content-Type: application/json
- Request body:
```json
{
	"email": "user@example.com",
	"password": "password123",
	"name": "User Name"
}
```

- URL: http://localhost:8080/api/v1/auth/login
- Request header: Content-Type: application/json
- Request body:
```json
{
	"email": "user@example.com",
	"password": "password123"
}
```

- URL: http://localhost:8080/api/v1/auth/me
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/auth/google/login
- Request header: none
- Request body: none

- URL: http://localhost:8080/api/v1/auth/google/callback?state=<state>&code=<code>
- Request header: none
- Request body: none

### Migrations (admin + migrations:* permission required)
- URL: http://localhost:8080/api/v1/migrations/up
- Request header: Authorization: Bearer <jwt>
- Request body: none
- Optional query: steps=<positive integer>

- URL: http://localhost:8080/api/v1/migrations/down?steps=<positive integer>
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/migrations/seed
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/migrations/version
- Request header: Authorization: Bearer <jwt>
- Request body: none

### Web Builder
- URL: http://localhost:8080/api/v1/api/v1/web-builder/sites
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"slug": "my-site",
	"config": {
		"title": "My Site",
		"name": "My Name",
		"headline": "Short headline",
		"bio": "Short bio",
		"avatar_url": "https://example.com/avatar.png",
		"links": [
			{"label": "GitHub", "url": "https://github.com/my"}
		]
	}
}
```

- URL: http://localhost:8080/api/v1/api/v1/web-builder/sites/:id/publish
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL (auto serve only): http://localhost:8080/api/v1/:slug
- Request header: none
- Request body: none

### Chatbot
- URL: http://localhost:8080/api/v1/api/v1/chatbot/generate
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"slug": "my-site",
	"prompt": "I am a designer focused on minimal portfolios",
	"name": "My Name",
	"headline": "Short headline",
	"bio": "Short bio",
	"avatar_url": "https://example.com/avatar.png",
	"links": [
		{"label": "GitHub", "url": "https://github.com/my"}
	]
}
```

### Payment
- URL: http://localhost:8080/api/v1/api/v1/payment/transactions
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"site_id": 1
}
```

- URL: http://localhost:8080/api/v1/api/v1/payment/webhook
- Request header: Content-Type: application/json
- Request body:
```json
{
	"order_id": "wb-1-1-20240502120000-abc123",
	"status_code": "200",
	"gross_amount": "99000",
	"signature_key": "<midtrans-signature>",
	"transaction_status": "settlement",
	"fraud_status": "accept"
}
```

### Template
- URL: http://localhost:8080/api/v1/api/v1/template/ping
- Request header: Authorization: Bearer <jwt>
- Request body: none