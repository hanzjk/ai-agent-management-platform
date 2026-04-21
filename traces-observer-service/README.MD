# Traces Observer Service

## Overview

Traces Observer Service is a lightweight Go backend that provides an HTTP API for querying traces collected in the observability plane. It communicates with OpenSearch to retrieve and aggregate tracing data produced by agents (components) across the platform, and exposes endpoints consumed by the console.

## How It Works

1. The console sends a tracing query to the HTTP API exposed by this service.
2. The service validates and translates the request into an OpenSearch query.
3. The OpenSearch client executes the query against the configured cluster/indices.
4. Results are formatted and returned to the caller (with paging and basic aggregation info where applicable).

```
HTTP API → Request Handler → Service/Query Layer → OpenSearch Client → OpenSearch Cluster
```

## Folder Structure

```
traces-observer-service/
├── config/         # Configuration management
├── controllers/    # Request handlers
├── docs/           # API documentation
├── handlers/       # HTTP request handlers
├── middleware/      # HTTP middleware (auth, logging)
├── opensearch/     # OpenSearch integration and query builders
├── scripts/        # Development and build scripts
├── .air.toml       # Air hot-reload configuration
├── .env.example    # Example environment variables
├── Dockerfile      # Production container build
├── Dockerfile.dev  # Development container with hot-reload
├── go.mod          # Go module definition
├── go.sum          # Go module checksums
├── main.go         # Application entry point
└── Makefile        # Build automation
```

## Prerequisites

- **Go**: Version 1.25.0
- **Make**: For build automation
- **air** (optional): Hot-reload for Go — `go install github.com/air-verse/air@latest`
- **OpenSearch**: A running OpenSearch instance (provided by the platform's observability plane, or run locally)

## Local Development

### 1. Clone the Repository

```bash
git clone https://github.com/wso2/agent-manager.git
cd agent-manager/traces-observer-service
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

Edit `.env` with your settings:

| Key | Description | Default |
|-----|-------------|---------|
| `TRACES_OBSERVER_PORT` | HTTP server port | `9098` |
| `OPENSEARCH_ADDRESS` | OpenSearch cluster URL | `https://localhost:9200` |
| `OPENSEARCH_USERNAME` | OpenSearch username | `admin` |
| `OPENSEARCH_PASSWORD` | OpenSearch password | `admin` |
| `DEFAULT_SPAN_QUERY_LIMIT` | Default max spans per query | `1000` |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | `DEBUG` |
| `IS_LOCAL_DEV_ENV` | Enable local development mode (skips strict auth) | `true` |

**Optional — JWT validation (must match agent-manager-service values when auth is enabled):**

| Key | Description |
|-----|-------------|
| `KEY_MANAGER_ISSUER` | JWT token issuer for validation |
| `KEY_MANAGER_AUDIENCE` | JWT token audience for validation |
| `KEY_MANAGER_JWKS_URL` | JWKS endpoint URL for JWT verification |

### 4. Ensure OpenSearch Is Running

The service requires a running OpenSearch instance. If you've set up the full platform with `make setup` from the project root, OpenSearch runs inside the k3d cluster. Forward it to localhost:

```bash
# From the project root
make port-forward
```

This forwards OpenSearch to `localhost:9200` (among other services).

Alternatively, run OpenSearch locally with Docker:

```bash
docker run -d \
  --name opensearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "OPENSEARCH_INITIAL_ADMIN_PASSWORD=ThisIsTheOpenSearchPassword1" \
  opensearchproject/opensearch:2.11.0
```

Then set in your `.env`:
```
OPENSEARCH_ADDRESS=https://localhost:9200
OPENSEARCH_PASSWORD=ThisIsTheOpenSearchPassword1
```

### 5. Start Development Server

Using Make with hot-reload (recommended):

```bash
make run
```

Or build and run manually:

```bash
go build -o traces-observer-service .
./traces-observer-service
```

Or run directly:

```bash
go run .
```

The service starts on `http://localhost:9098` by default.

### 6. Verify the Service

```bash
curl http://localhost:9098/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-09T12:34:56Z"
}
```

### 7. Run Tests

```bash
make test
```

Run tests with coverage:

```bash
make test-cover
```

## Development Tools

| Command | Description |
|---------|-------------|
| `make run` | Start dev server with hot-reload (requires `air`) |
| `make build` | Build the Go binary |
| `make test` | Run unit tests |
| `make test-cover` | Run tests with coverage report |
| `make fmt` | Format code |
| `make lint` | Run linters (requires `golangci-lint`) |
| `make clean` | Clean build artifacts |
| `make help` | Show all available commands |

## Docker

Build the image:

```bash
make docker-build
```

Run the container:

```bash
make docker-run
```

Or manually:

```bash
docker build -t traces-observer-service .
docker run --rm -p 9098:9098 \
  -e OPENSEARCH_ADDRESS=http://host.docker.internal:9200 \
  -e OPENSEARCH_USERNAME=admin \
  -e OPENSEARCH_PASSWORD=ThisIsTheOpenSearchPassword1 \
  -e TRACES_OBSERVER_PORT=9098 \
  traces-observer-service
```

## API Reference

All endpoints use standard HTTP methods with query parameters.

### 1. List Traces — `GET /api/v1/traces`

Retrieves a list of trace overviews for a specific service with optional time filtering and pagination.

**Query Parameters:**

| Parameter | Required | Description | Default |
|-----------|----------|-------------|---------|
| `serviceName` | Yes | Name of the service to query traces for | — |
| `startTime` | No | Start time in RFC3339 format | — |
| `endTime` | No | End time in RFC3339 format | — |
| `limit` | No | Maximum number of traces to return | `10` |
| `offset` | No | Number of traces to skip for pagination | `0` |
| `sortOrder` | No | Sort order: `asc` or `desc` | `desc` |

**Example:**

```bash
curl 'http://localhost:9098/api/v1/traces?serviceName=sample-app&startTime=2025-11-03T00:00:00Z&endTime=2025-11-08T23:59:59Z&limit=10&offset=0'
```

**Response (200):**

```json
{
  "traces": [
    {
      "traceId": "5974d036b3d7709f2fc9f2b48461c176",
      "rootSpanId": "58f16238f09ae1b2",
      "rootSpanName": "LangGraph.workflow",
      "startTime": "2025-11-07T06:23:24.035086494Z",
      "endTime": "2025-11-07T06:23:27.545584559Z",
      "spanCount": 8
    }
  ],
  "totalCount": 1
}
```

### 2. Get Trace Spans — `GET /api/v1/trace`

Retrieves all spans for a specific trace ID and service.

**Query Parameters:**

| Parameter | Required | Description | Default |
|-----------|----------|-------------|---------|
| `traceId` | Yes | The trace ID to retrieve spans for | — |
| `serviceName` | Yes | Name of the service | — |
| `sortOrder` | No | Sort order: `asc` or `desc` | `asc` |
| `limit` | No | Maximum number of spans to return | `100` |

**Example:**

```bash
curl 'http://localhost:9098/api/v1/trace?traceId=21a29d5d24837ca724b8751494e70a95&serviceName=langchain-docker-app&sortOrder=asc&limit=100'
```

**Response (200):**

```json
{
  "spans": [
    {
      "traceId": "21a29d5d24837ca724b8751494e70a95",
      "spanId": "e2c22d3d4b7736bd",
      "name": "LangGraph.workflow",
      "service": "langchain-docker-app",
      "startTime": "2025-11-03T11:42:18.329535246Z",
      "endTime": "2025-11-03T11:42:20.55201954Z",
      "duration": 2222484,
      "kind": "SPAN_KIND_INTERNAL",
      "status": "0",
      "attributes": {
        "traceloop.entity.input": "{...}",
        "traceloop.entity.output": "{...}",
        "traceloop.workflow.name": "LangGraph"
      }
    }
  ],
  "totalCount": 2
}
```

### 3. Health Check — `GET /health`

```bash
curl http://localhost:9098/health
```

### Error Responses

| Status | Description |
|--------|-------------|
| `200 OK` | Success |
| `400 Bad Request` | Invalid parameters (missing required fields, invalid format) |
| `500 Internal Server Error` | Server/OpenSearch errors |

## Mock Trace Generation

Generate mock traces using OTLP HTTP for testing:

### Prerequisites

```bash
go install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen@latest
```

Ensure `$(go env GOPATH)/bin` is in your `PATH`.

### Generate Mock Traces

```bash
export AMP_AGENT_API_KEY="<your-api-key>"
make mock-traces
```

Override defaults on the command line:

```bash
make mock-traces SPANS_PER_TRACE=5 TRACE_COUNT=20
```

See the `Makefile` for all configurable variables (`OTLP_ENDPOINT`, `TRACE_SERVICE_NAME`, etc.).

## Running with the Full Platform

For a fully integrated local setup, see the [Local Setup Guide](../deployments/LOCAL-SETUP.md).
