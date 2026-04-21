# Local Development Setup Guide

This guide walks you through setting up the entire Agent Management Platform locally for development. By the end, you'll have all platform services running on your machine.

## Architecture Overview

The local development environment consists of:

| Service | Description | Local URL |
|---------|-------------|-----------|
| **Console** | React web UI | http://localhost:3000 |
| **Agent Manager Service** | Go backend API | http://localhost:9000 |
| **PostgreSQL** | Database | localhost:5432 |
| **Traces Observer Service** | Trace query API | http://localhost:9098 |
| **OpenChoreo** | Kubernetes runtime (k3d) | Various ports via port-forward |

The core services (Console, Agent Manager Service, PostgreSQL) run via **Docker Compose**. OpenChoreo and its extensions (observability, secrets, IDP, gateway) run in a **k3d** Kubernetes cluster.

```
┌──────────────────────────────────────────────────────────────┐
│  Docker Compose (deployments/docker-compose.yml)             │
│                                                              │
│  ┌─────────────┐  ┌──────────────────┐  ┌───────────────┐   │
│  │  PostgreSQL  │  │ Agent Manager    │  │   Console     │   │
│  │  :5432       │←─│ Service :9000    │←─│   :3000       │   │
│  └─────────────┘  └──────────────────┘  └───────────────┘   │
│                           │                                  │
└───────────────────────────│──────────────────────────────────┘
                            │
┌───────────────────────────│──────────────────────────────────┐
│  k3d Cluster              ↓                                  │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────┐  │
│  │ OpenChoreo   │  │ OpenSearch  │  │ Traces Observer    │  │
│  │ Control Plane│  │ :9200       │  │ Service :9098      │  │
│  └──────────────┘  └─────────────┘  └────────────────────┘  │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────┐  │
│  │ Thunder IDP  │  │ OpenBao     │  │ AI Gateway         │  │
│  │ :8090        │  │ :8200       │  │ :8084              │  │
│  └──────────────┘  └─────────────┘  └────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## Prerequisites

Install the following tools before starting:

| Tool | Version | Installation |
|------|---------|-------------|
| **Docker** | v26+ | [docker.com](https://www.docker.com/) or via Colima |
| **Colima** | v0.9.0 | `brew install colima` |
| **k3d** | v5.8+ | `brew install k3d` |
| **kubectl** | v1.32+ | `brew install kubectl` |
| **Helm** | v3.12+ | `brew install helm` |
| **Node.js** | 20.19.0+ or 22.12.0+ | `brew install node` or [nvm](https://github.com/nvm-sh/nvm) |
| **Go** | 1.25.0+ | `brew install go` or [go.dev](https://go.dev/dl/) |
| **Rush** | 5.157.0 | `npm install -g @microsoft/rush@5.157.0` |
| **Make** | Any | Pre-installed on macOS |

### Verify Prerequisites

```bash
docker --version
colima version
k3d version
kubectl version --client
helm version
node --version    # Should be >=20.19.0 or >=22.12.0
go version        # Should be >=1.25.0
rush --version
```

## Setup Options

There are two ways to set up the local environment:

- **Option A: Full Platform Setup (Recommended)** — automated one-command setup of everything
- **Option B: Core Services Only** — just the console, API, and database via Docker Compose

---

## Option A: Full Platform Setup

This sets up everything: Colima VM, k3d cluster, OpenChoreo, and all platform services.

### Step 1: Run the Full Setup

From the **project root**:

```bash
make setup
```

This single command runs the following in sequence:

1. **`setup-colima`** — Starts a Colima VM with 4 CPUs, 8 GB RAM, and Rosetta for x86_64 compatibility
2. **`setup-k3d`** — Creates a k3d Kubernetes cluster named `openchoreo-local-setup`
3. **`setup-openchoreo`** — Installs OpenChoreo planes (control, data, workflow, observability) and AMP extensions (Thunder IDP, OpenBao secrets, AI gateway, evaluation workflows)
4. **`setup-platform`** — Generates JWT keys, builds Docker images, starts Docker Compose services (PostgreSQL, Agent Manager Service, Console)
5. **`setup-console-local`** — Installs console dependencies and builds the monorepo

> **Note:** The full setup can take 15-30 minutes on the first run, depending on your internet speed and machine. Subsequent runs are much faster since it skips already-running components.

### Step 2: Forward Kubernetes Services

In a **separate terminal**, run:

```bash
make port-forward
```

Keep this terminal open. It forwards these services from the k3d cluster to localhost:

| Service | Local Port |
|---------|-----------|
| OpenSearch | 9200 |
| Traces Observer Service | 9098 |
| Observer Service API | 8085 |
| Thunder IDP | 8090 |
| Observability Gateway (HTTP) | 22893 |
| Observability Gateway (HTTPS) | 22894 |
| OpenBao (Data Plane) | 8200 |
| OpenBao (Workflow Plane) | 8201 |
| OpenChoreo API | 8195 |
| AI Gateway Runtime | 8084 |

Press `Ctrl+C` to stop port forwarding.

### Step 3: Run Database Migrations

```bash
make dev-migrate
```

### Step 4: Verify Everything Is Running

```bash
# Check Docker Compose services
cd deployments && docker compose ps && cd ..

# Check the console
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000

# Check the API
curl http://localhost:9000/health

# Check OpenChoreo cluster
kubectl get pods --all-namespaces --context k3d-openchoreo-local-setup | head -20
```

### Access Points

| Service | URL |
|---------|-----|
| Console | http://localhost:3000 |
| Agent Manager API | http://localhost:9000 |
| Traces Observer API | http://localhost:9098 |
| Database | `postgresql://agentmanager:agentmanager@localhost:5432/agentmanager` |

---

## Option B: Core Services Only (Docker Compose)

If you only need the console, API, and database (no OpenChoreo, no observability):

### Step 1: Generate JWT Keys

```bash
make gen-keys
```

### Step 2: Install Console Dependencies

```bash
make setup-console-local
```

### Step 3: Start Services

```bash
make dev-up
```

This starts three Docker Compose services:
- **PostgreSQL 16** on port 5432
- **Agent Manager Service** on port 9000 (with hot-reload via Air)
- **Console** on port 3000 (with hot-reload via Vite HMR)

### Step 4: Run Database Migrations

```bash
make dev-migrate
```

### Step 5: Verify

```bash
curl http://localhost:9000/health
# Open http://localhost:3000 in your browser
```


---

## Daily Development Workflow

Once the initial setup is complete, your daily workflow is:

### Starting Your Day

```bash
# Start OpenChoreo cluster (if using full setup)
make openchoreo-up

# Start platform services (Console + API + DB)
make dev-up

# Forward Kubernetes services (in a separate terminal, if using full setup)
make port-forward
```

### During Development

- **Code changes to Agent Manager Service** — Auto-reloaded by Air (inside Docker container)
- **Code changes to Console** — Auto-reloaded by Vite HMR (inside Docker container)
- **Database migration needed** — `make dev-migrate`
- **View logs** — `make dev-logs` (all services) or `make service-logs` / `make console-logs`
- **Shell into service container** — `make service-shell`
- **Connect to database** — `make db-connect`

### Ending Your Day

```bash
# Stop platform services
make dev-down

# Stop OpenChoreo cluster (saves resources)
make openchoreo-down
```

## Running Services Outside Docker

If you prefer to run services directly on your host (e.g., for debugging with an IDE):

### Agent Manager Service

See the full guide in [agent-manager-service/README.md](../agent-manager-service/README.md).

```bash
cd agent-manager-service
cp .env.example .env
# Edit .env as needed
make gen-keys
make dev-migrate
make run
```

### Console

See the full guide in [console/README.md](../console/README.md).

```bash
cd console
nvm use
make install
make build-webapp
cp apps/webapp/public/config.template.js apps/webapp/public/config.js
# Edit config.js (set apiBaseUrl to http://localhost:9000, etc.)
make dev
```

### Traces Observer Service

See the full guide in [traces-observer-service/README.MD](../traces-observer-service/README.MD).

```bash
cd traces-observer-service
cp .env.example .env
# Edit .env (ensure OpenSearch is reachable)
make run
```

## Rebuilding After Major Changes

If you pull changes that affect Docker images, dependencies, or configuration:

```bash
# Rebuild Docker images and restart
make dev-rebuild

# If console dependencies changed (rush.json or pnpm-lock.yaml)
make setup-console-local-force
```

## Cleanup

### Stop Everything

```bash
make dev-down
make openchoreo-down
```

### Full Teardown (removes k3d cluster and all data)

```bash
make teardown
```

## Troubleshooting

### Docker Compose services won't start

- Ensure Colima is running: `colima status`
- Ensure Docker is accessible: `docker info`
- Check if ports are already in use: `lsof -i :3000` / `lsof -i :9000` / `lsof -i :5432`

### Agent Manager Service keeps restarting

- Check logs: `make service-logs`
- Ensure database is healthy: `make db-logs`
- Ensure JWT keys exist: `make gen-keys`

### Console shows blank page

- Check console logs: `make console-logs`
- Verify `config.js` exists and has correct API URLs
- Check that the Agent Manager Service is reachable at the configured `apiBaseUrl`

### k3d cluster not accessible

- Verify cluster is running: `k3d cluster list`
- Refresh kubeconfig: `k3d kubeconfig merge openchoreo-local-setup --kubeconfig-merge-default --kubeconfig-switch-context`
- Check node status: `kubectl get nodes --context k3d-openchoreo-local-setup`

### Port forwarding fails

- Ensure the k3d cluster is running: `make openchoreo-up`
- Check pod status: `kubectl get pods --all-namespaces --context k3d-openchoreo-local-setup`
- Wait for pods to be ready if cluster just started

### Database connection issues

- Check if PostgreSQL container is running: `docker ps | grep agent-manager-db`
- Verify with: `docker exec agent-manager-db pg_isready -U agentmanager`
- Check connection from host: `psql -h localhost -U agentmanager -d agentmanager`

### "k3d-openchoreo-local-setup" network not found

The Docker Compose file references an external Docker network created by k3d. If you're running Option B (core services only) without k3d, create the network manually:

```bash
docker network create k3d-openchoreo-local-setup
```

## Make Command Reference

Run `make help` from the project root to see all available commands. Key commands:

| Command | Description |
|---------|-------------|
| `make setup` | Complete first-time setup |
| `make dev-up` | Start platform services |
| `make dev-down` | Stop platform services |
| `make dev-restart` | Restart platform services |
| `make dev-rebuild` | Rebuild images and restart |
| `make dev-logs` | Tail all platform logs |
| `make dev-migrate` | Run database migrations |
| `make port-forward` | Forward k3d services to localhost |
| `make openchoreo-up` | Start k3d cluster |
| `make openchoreo-down` | Stop k3d cluster |
| `make openchoreo-status` | Check k3d cluster status |
| `make db-connect` | Open PostgreSQL shell |
| `make service-logs` | View Agent Manager Service logs |
| `make console-logs` | View Console logs |
| `make teardown` | Remove everything |
