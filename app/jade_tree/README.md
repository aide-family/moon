# Jade Tree

Jade Tree is the Moon agent runtime service.

## Purpose

- Deploy as an agent on target servers.
- Provide a stable collection and communication endpoint.
- Support RPM + `systemctl` based production operations.

## Architecture

Jade Tree follows the same layered structure used by existing Moon services:

- `cmd` -> command entry and runtime modes
- `internal/server` -> HTTP/gRPC server bootstrap and middleware
- `internal/service` -> service handlers
- `internal/biz` -> business layer
- `internal/data` -> data/repository layer

## API Overview

| Service | Method / HTTP | Description |
|---------|---------------|-------------|
| `magicbox.api.v1.Health` | `GET /health` | Health check endpoint for liveness/readiness |

## Deployment

Jade Tree is designed for RPM + systemd deployment.

- systemd unit template: `deploy/systemd/jade-tree.service`
- RPM notes: `deploy/rpm/README.md`

## Run

```bash
make all
make dev
```
