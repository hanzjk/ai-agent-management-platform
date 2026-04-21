# Agent Manager Console

React/TypeScript web application for the Agent Manager platform, built as a Rush monorepo.

## Tech Stack

- **React 19** — UI framework
- **TypeScript** — Type safety
- **Vite** — Build tool and dev server
- **Rush** — Monorepo management
- **pnpm** — Package manager (managed by Rush)

## Prerequisites

- **Node.js**: Version 20.19.0+ or 22.12.0+ (see `.nvmrc` for the pinned version)
- **Rush**: Monorepo management tool
- **pnpm**: Installed automatically by Rush

### Install Node.js

Use [nvm](https://github.com/nvm-sh/nvm) to install the correct version:

```bash
cd console
nvm install    # reads .nvmrc → installs 22.12.0
nvm use
```

Or install Node.js 22.12.0+ manually from [nodejs.org](https://nodejs.org/).

### Install Rush

```bash
npm install -g @microsoft/rush@5.157.0
```

Verify installation:
```bash
rush --version
```

## Getting Started

### 1. Install Dependencies

From the `console/` directory:

```bash
cd console
make install
```

This runs `rush install`, which:
- Installs Rush's local copy of pnpm
- Installs all dependencies for all projects in the monorepo
- Creates symlinks between local packages

### 2. Build Libraries

Build the webapp and all its dependencies:

```bash
make build-webapp
```

Or build all projects:
```bash
make build
```

### 3. Configure the Application

The console uses a runtime configuration file to connect to backend services. Copy the template:

```bash
cp apps/webapp/public/config.template.js apps/webapp/public/config.js
```

Edit `apps/webapp/public/config.js` for local development:

```javascript
window.__RUNTIME_CONFIG__ = {
  authConfig: {
    // ... leave auth defaults for local dev
  },
  disableAuth: 'true' === 'true',        // Set to true for local dev without IDP
  apiBaseUrl: 'http://localhost:9000',     // Agent Manager Service API
  obsApiBaseUrl: 'http://localhost:9098',  // Traces Observer Service API
  gatewayControlPlaneUrl: 'http://localhost:9243',
  gatewayVersion: 'v0.9.0',
  instrumentationUrl: '',
  guardrailsCatalogUrl: '',
  guardrailsDefinitionBaseUrl: '',
};
```

**Key configuration values for local development:**

| Variable | Description | Local Value |
|----------|-------------|-------------|
| `disableAuth` | Bypass authentication (set `'true'` string comparison) | `true` |
| `apiBaseUrl` | Agent Manager Service URL | `http://localhost:9000` |
| `obsApiBaseUrl` | Traces Observer Service URL | `http://localhost:9098` |
| `gatewayControlPlaneUrl` | Gateway control plane WebSocket URL | `http://localhost:9243` |

### 4. Start Development Server

```bash
make dev
```

This starts the Vite dev server at `http://localhost:3000` with hot module replacement (HMR).

Press `Ctrl+C` to stop.

## Available Commands

### Make Commands (Recommended)

```bash
make dev         # Start development mode with hot-reload
make install     # Install dependencies
make build       # Build all projects
make build-webapp # Build webapp and its dependencies
make clean       # Clean build outputs
make purge       # Purge Rush cache
make help        # Show all available commands
```

### Rush Commands

```bash
rush install                                    # Install dependencies
rush build                                      # Build all projects
rush build --to @agent-management-platform/webapp  # Build webapp + dependencies
rush lint                                       # Lint all projects
rush test                                       # Test all projects
rush purge                                      # Clean all build outputs
rush update                                     # Update dependencies
rush create-page                                # Create a new page component
```

### Project-Specific Commands

Navigate to any project directory and use `rushx`:

```bash
cd apps/webapp
rushx dev        # Start development server
rushx build      # Build for production
rushx lint       # Run linting
rushx lint:fix   # Fix linting issues
rushx preview    # Preview production build
```

## Creating New Page Components

Create new page components using the integrated Yeoman generator:

```bash
cd console
rush create-page
```

Answer the prompts:
- **Package name** (e.g., `user-dashboard`) — use kebab-case
- **Display title** (e.g., `User Dashboard`)
- **Description** (e.g., `A dashboard page for managing users`)
- **Route path** (e.g., `/user-dashboard`)

After generating a new page:

1. Add the new page to `rush.json` projects list:
   ```json
   {
     "packageName": "@agent-management-platform/your-page-name",
     "projectFolder": "workspaces/pages/your-page-name"
   }
   ```

2. Update Rush:
   ```bash
   rush update
   ```

3. Build your new page:
   ```bash
   cd workspaces/pages/your-page-name
   rushx build
   ```

## Project Structure

### Apps
- **webapp** — Main React application with Vite build system

### Libraries (`workspaces/libs/`)
- **auth** — Authentication provider and hooks
- **types** — Shared TypeScript type definitions
- **eslint-config** — Shared ESLint configuration
- **views** — Shared UI components and themes
- **api-client** — API client utilities

### Pages (`workspaces/pages/`)
Feature page components, each as a separate Rush project. Use `rush create-page` to scaffold new pages.

## Troubleshooting

### `rush install` fails
- Ensure you're using the correct Node.js version: `nvm use`
- Try purging and reinstalling: `rush purge && rush install`

### Config changes not reflected
- The `config.js` file is loaded at runtime, not build time. Hard-refresh your browser (`Cmd+Shift+R` / `Ctrl+Shift+R`).

### Hot reload not working
- Ensure Vite dev server is running (`make dev`)
- Check that library dependencies are built (`make build-webapp`)

## Running with Docker Compose

For a fully integrated local setup (database + service + console), see the [Local Setup Guide](../deployments/LOCAL-SETUP.md).
