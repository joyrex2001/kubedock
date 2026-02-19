# Development Documentation

## Architecture Overview

Kubedock is a minimal Docker API implementation that orchestrates containers on a Kubernetes cluster instead of running them locally. It acts as a drop-in replacement for the Docker API, translating container operations into Kubernetes resources (Pods, Services, ConfigMaps).

## Project Structure

```
kubedock/
├── cmd/                  # CLI command implementations
│   ├── root.go           # Root command setup
│   ├── server.go         # Server command and flags
│   ├── version.go        # Version command
│   └── dind.go           # Docker-in-docker support
├── internal/
│   ├── main.go           # Application entry point
│   ├── backend/          # Kubernetes backend implementation
│   ├── server/           # HTTP API server
│   ├── reaper/           # Resource cleanup/garbage collection
│   ├── model/            # Data models and database
│   ├── config/           # Configuration management
│   ├── events/           # Event handling
│   ├── dind/             # Docker-in-docker sidecar
│   └── util/             # Utility packages
├── docs/                 # Documentation
└── main.go               # Binary entry point
```

## Core Components

### 1. Backend (`internal/backend/`)

The backend package handles all Kubernetes API interactions. Key responsibilities:
- Creating, listing, and deleting pods
- Managing services and configmaps
- Handling port forwarding
- Container log retrieval
- Exec operations in containers

Main components:
- `deploy.go` - Pod and service deployment
- `delete.go` - Resource cleanup
- `logs.go` - Log retrieval
- `exec.go` - Container execution
- `copy.go` - File copy to/from containers
- `image.go` - Image handling

### 2. Server (`internal/server/`)

The HTTP API server implements the Docker API and the Podman API. It uses the Gin web framework and exposes two API flavors:
- Docker API (`/v1.24/*`)
- Podman API (`/v1.40/libpod/*`)

The routes are registered in `routes/docker/` and `routes/libpod/` directories.

### Implementing New Endpoints

When adding a new API endpoint, **both Docker and Podman routes must be implemented**. The project supports two API flavors:
- Docker API (`routes/docker/`)
- Podman API (`routes/libpod/`)

Common functionality should be extracted to `routes/common/` to avoid duplication. Each new endpoint needs:
1. Implementation in `routes/docker/` for Docker compatibility
2. Implementation in `routes/libpod/` for Podman compatibility
3. Shared logic in `routes/common/` when applicable

### 4. Reaper (`internal/reaper/`)

The reaper handles automatic cleanup of orphaned resources:
- Removes containers older than configured threshold (default: 60 minutes)
- Deletes resources owned by terminated kubedock instances
- Can be triggered at startup with `--prune-start`

### 5. Model (`internal/model/`)

Contains data models and an in-memory database for tracking:
- Containers
- Networks
- Images
- Exec instances

## Development Setup

### Prerequisites

- Go 1.25+
- Kubernetes cluster (minikube, kind, or remote)
- kubectl configured

### Local Kubernetes Cluster

For local development, **minikube** is recommended. See the [minikube documentation](https://minikube.sigs.k8s.io/docs/) for installation and setup.

Alternatively, you can use [kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker).

### Building

```bash
# Build the binary
make build

# Or run directly
make run
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make cover

# Run linter
make lint
```

### Code Formatting

```bash
# Format code
make fmt
```

## Configuration

Configuration is managed via:
1. Command-line flags
2. Environment variables
3. Configuration files

Key configuration groups:
- `server.*` - API server settings (listen address, TLS)
- `kubernetes.*` - Kubernetes settings (namespace, images, timeouts)
- `reaper.*` - Resource cleanup settings
- `lock.*` - Namespace locking settings

See `cmd/server.go` for all available flags and their corresponding environment variables.

## Running Locally

```bash
# Start kubedock with port forwarding
kubedock server --port-forward -v 2

# Use with testcontainers
export DOCKER_HOST=tcp://127.0.0.1:2475
export TESTCONTAINERS_RYUK_DISABLED=true
mvn test
```

## Key Design Decisions

### Namespace Locking
When multiple kubedock instances share a namespace, namespace locking prevents network alias collisions using Kubernetes leases.

### Volume Handling
Volumes are implemented as one-way copies via init containers. Single files are stored as ConfigMaps.

### Network Architecture
All containers run in the same flattened network namespace. Network aliases create Kubernetes Services.

### Resource Management
- Containers: Kubernetes Pods
- Networking: Kubernetes Services  
- Volumes: Init containers + ConfigMaps
- File transfers: ConfigMaps

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Ensure `make test` and `make lint` pass
5. Submit a pull request
