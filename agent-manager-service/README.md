# Agent Manager Service

## Overview

The Agent Manager Service is a core component of the Agent Management Platform that handles agent deployment, management, and governing AI agents. It provides the backend API powering the control plane, including agent lifecycle management, JWT-based authentication, secret management, and integration with OpenChoreo for Kubernetes deployments.

## Folder Structure

```
agent-manager-service/
├── api/                        # HTTP API layer with HTTP handlers and routing
├── clients/                   # External service clients (OpenChoreo, Observer, API Platform, etc.)
├── config/                    # Configuration management
├── controllers/               # HTTP request controllers
├── db/                       # Database connection and utilities
├── db_migrations/            # Database schema migration files
├── db_types/                 # Custom database types
├── docs/                     # OpenAPI documentation
├── keys/                     # JWT signing keys (generated via make gen-keys)
├── middleware/               # HTTP middleware (auth, logging, recovery)
├── models/                   # Data models and entities
├── repositories/             # Data access layer
├── resources/                # Static resources (LLM provider templates, etc.)
├── scripts/                  # Development and build scripts
├── services/                 # Business logic layer
├── signals/                  # Graceful shutdown handling
├── tests/                    # Test files
├── utils/                    # Utility functions
├── wiring/                   # Dependency injection (Wire)
├── .air.toml                 # Air hot-reload configuration
├── .env.example              # Example environment variables
├── Dockerfile                # Production container build
├── Dockerfile.dev            # Development container with hot-reload
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── main.go                   # Application entry point
└── Makefile                  # Build automation
```

## Prerequisites

- **Go**: Version 1.25.0 or later
- **PostgreSQL**: Version 12 or later (16 recommended)
- **Make**: For build automation
- **air**: Hot-reload for Go — `go install github.com/air-verse/air@latest`
- **moq**: Mock generation — `go install github.com/matryer/moq@v0.5.3`
- **wire**: Dependency injection — install via `go install github.com/google/wire/cmd/wire@latest`
- **oapi-codegen**: OpenAPI client generation — `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`

## Local Development

### 1. Clone the Repository

```bash
git clone https://github.com/wso2/agent-manager.git
cd agent-manager/agent-manager-service
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Environment Variables

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Edit `.env` with your local settings. The key configuration sections are:

#### Server Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `SERVER_HOST` | Host address where the server runs | _(empty = all interfaces)_ |
| `SERVER_PORT` | Port number for the server | `9000` |
| `CORS_ALLOWED_ORIGIN` | Allowed CORS origin for the console | `http://localhost:3000` |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | `INFO` |
| `IS_LOCAL_DEV_ENV` | Enable local development mode | `true` |

#### Database Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `DB_HOST` | Database host address | `localhost` |
| `DB_PORT` | Database port number | `5432` |
| `DB_USER` | Username for database authentication | `agentmanager` |
| `DB_PASSWORD` | Password for database authentication | `agentmanager` |
| `DB_NAME` | Name of the database | `agentmanager` |

#### JWT Signing Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `JWT_SIGNING_PRIVATE_KEY_PATH` | Path to RSA private key for JWT signing | `keys/private.pem` |
| `JWT_SIGNING_PUBLIC_KEYS_CONFIG` | Path to JSON config file containing public keys | `keys/public-keys-config.json` |
| `JWT_SIGNING_ACTIVE_KEY_ID` | Key ID for active signing key | `key-1` |
| `JWT_SIGNING_DEFAULT_EXPIRY` | Default token expiry duration | `8760h` (1 year) |
| `JWT_SIGNING_ISSUER` | Issuer claim for JWT tokens | `agent-manager-service` |
| `JWT_SIGNING_DEFAULT_ENVIRONMENT` | Default environment for token claims | `default` |


#### Encryption Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `ENCRYPTION_KEY` | Hex-encoded 32-byte key for AES-256-GCM encryption of secrets at rest. Generate with `openssl rand -hex 32` | _(example key in .env.example)_ |

#### OpenChoreo Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `OPEN_CHOREO_BASE_URL` | OpenChoreo API base URL | `http://api.openchoreo.localhost:8195` |

#### Optional Configuration

| Key | Description |
|-----|-------------|
| `OBSERVER_URL` | Observer service URL for observability data |
| `IDP_TOKEN_URL` | Thunder IDP OAuth2 token endpoint |
| `IDP_CLIENT_ID` | OAuth2 client ID for service-to-service auth |
| `IDP_CLIENT_SECRET` | OAuth2 client secret |
| `SECRET_MANAGER_PROVIDER` | Secret manager provider (`openbao`) |
| `OPENBAO_URL` | OpenBao/Vault URL |
| `OPENBAO_TOKEN` | OpenBao/Vault access token |
| `DEFAULT_GATEWAY_PORT` | Default AI gateway port |
| `LLM_TEMPLATE_DEFINITIONS_PATH` | Path to LLM provider template definitions |

See `.env.example` for the full list of available configuration options.

### 4. Set Up Database

Start a PostgreSQL instance. The easiest way is with Docker:

```bash
docker run -d \
  --name agent-manager-db \
  -e POSTGRES_DB=agentmanager \
  -e POSTGRES_USER=agentmanager \
  -e POSTGRES_PASSWORD=agentmanager \
  -p 5432:5432 \
  postgres:16-alpine
```

Or if you have PostgreSQL installed locally, create the database:

```bash
createdb -U postgres agentmanager
```

### 5. Generate JWT Signing Keys

Generate RSA key pairs for JWT token signing:

```bash
make gen-keys
```

**Generated Artifacts (default key-1):**
- `keys/private.pem` — Private signing key
- `keys/public.pem` — Public verification key
- `keys/public-keys-config.json` — Public keys configuration with key ID "key-1"

**Key Rotation (generating additional keys):**
```bash
./scripts/gen_keys.sh key-2
```

This produces `keys/private-key-2.pem` and `keys/public-key-2.pem`. Update `keys/public-keys-config.json` to include key-2, then set:
```bash
JWT_SIGNING_PRIVATE_KEY_PATH=./keys/private-key-2.pem
JWT_SIGNING_ACTIVE_KEY_ID=key-2
```

### 6. Run Database Migrations

```bash
make dev-migrate
```

### 7. Start Development Server

Using Make:

```bash
make run
```

The service will start on `http://localhost:9000` by default with hot-reloading enabled.

### 8. Verify the Service

```bash
curl http://localhost:9000/health
```

### 9. Run Tests

Run tests against the local database:

```bash
make test
```

Run tests with an isolated test database:

```bash
make dev-test
```

## Development Tools

| Command | Description |
|---------|-------------|
| `make run` | Start dev server with hot-reload |
| `make test` | Run tests |
| `make dev-test` | Run tests with isolated test database |
| `make fmt` | Format code |
| `make lint` | Run linters |
| `make wire` | Generate Wire dependency injection code |
| `make codegen` | Run go generate (Wire + models) |
| `make spec` | Generate types/client from OpenAPI spec |
| `make gen-keys` | Generate JWT signing keys |
| `make dev-migrate` | Run database migrations |
| `make gen-evaluators-dev` | Generate builtin evaluator catalog (dev mode) |
| `make help` | Show all available commands |

## API Documentation

### OpenAPI Specification

The API is documented using OpenAPI 3.0 specification in `docs/api_v1_openapi.yaml`.

### Agent Token Authentication

The service provides JWT-based authentication for external agents:

- **Token Generation**: `POST /api/v1/orgs/{orgName}/projects/{projName}/agents/{agentName}/token`
  - Generate a signed JWT token for an agent
  - Optional parameters: `environment` (query), `expires_in` (body)
  - Returns a Bearer token with configurable expiry

- **JWKS Endpoint**: `GET /auth/external/jwks.json`
  - Public endpoint for retrieving JSON Web Key Set
  - Used by clients to verify JWT signatures
  - No authentication required

## Running with Docker Compose

For a fully integrated local setup (database + service + console), see the [Local Setup Guide](../deployments/LOCAL-SETUP.md).
