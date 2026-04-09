# Jade Tree

Jade Tree is the Moon agent runtime service.

## Purpose

- Deploy as an agent on target servers.
- Provide a stable collection and communication endpoint.
- Support RPM + `systemctl` based production operations.
- Manage predefined SSH command templates with an approval workflow, then execute them against remote hosts.
- Support one-click command dispatch with target preview/count and include/exclude filters (machine, system type, and agent version).
- Collect deployment machine profile details (CPU, memory, disk/mount usage, network, hostname, and system basics).
- Actively report local machine profile (including self-reported `agent` endpoint/version) to configured HTTP endpoints on schedule.
- Provide CLI machine operations: `machine get`, `machine list`, `machine push`, and `machine pull` with `table/json/yaml` output.
- Expose probe metrics in Prometheus format (`probe_tcp_*`, `probe_http_*`, `probe_port_*`, `probe_tls_cert_*`) on `/metrics`.

## Architecture

Jade Tree follows the same layered structure used by existing Moon services:

- `cmd` -> command entry and runtime modes
- `internal/server` -> HTTP/gRPC server bootstrap and middleware
- `internal/service` -> service handlers
- `internal/biz` -> business layer
- `internal/data` -> data/repository layer

API definitions live under `proto/jade_tree/api/v1/`; generated Go code is in `pkg/api/v1/`. Run `make api` (included in `make all`) after changing protos. Audit **kind** and **status** use `magicbox.enum.SSHCommandAuditKind` / `SSHCommandAuditStatus` from `proto/magicbox/enum/enum.proto` (regenerate magicbox with `make proto` in `magicbox/` when those enums change).

## Configuration

- **Database** (`bootstrap.database` in `config/server.yaml`) is **required**. SQLite is the default for local development; use MySQL or PostgreSQL in production by switching `dialector` and `options` (same shape as other Moon apps).
- Tables `ssh_commands` and `ssh_command_audits` are created via GORM `AutoMigrate` on startup.
- Active machine info reporting is configured by `bootstrap.machineInfoReport` (`enabled`, `collectSelfData`, `interval`, `timeout`, `endpoints`, and optional request `headers`).

## API Overview

| Service | Method / HTTP | Description |
|---------|---------------|-------------|
| `magicbox.api.v1.Health` | `GET /health` | Health check endpoint for liveness/readiness |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/submissions` | Submit a new command definition for review (creates audit row) |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/{command_uid}/submissions` | Submit changes to an existing command for review |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-commands` | List approved commands (paginated) |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-commands/{uid}` | Get one approved command |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-command-audits` | List audit records (optional `statusFilter` query) |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-command-audits/{uid}` | Get one audit record |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-audits/{uid}/approve` | Approve audit; applies payload to `ssh_commands` and sets audit status |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-audits/{uid}/reject` | Reject pending audit with a reason |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/{command_uid}/execute` | Run stored command on a remote host (host, credentials, optional timeout in body) |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/batch-execute` | Batch receive and execute multiple command requests in one call |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/dispatch/count` | Count dispatch target machines before one-click dispatch (default excludes self) |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/dispatch` | One-click dispatch command to all matched reported agents with include/exclude filters |
| `jade_tree.api.v1.MachineInfo` | `GET /v1/machine-info` | Get deployment machine details (CPU, memory, disk+mount usage, network, hostname, machine UUID, arch/os/version/kernel, and `agent` endpoint/version) |
| `jade_tree.api.v1.MachineInfo` | `POST /v1/machine-info/report` | Report machines (deduped by machine UUID + hostname + local IP) and persist; empty response |
| `jade_tree.api.v1.MachineInfo` | `GET /v1/machine-infos` | List machines known by the receiver/leader (same composite identity, paginated by `page`/`pageSize`) |
| `jade_tree.api.v1.ProbeTask` | `POST /v1/probe-tasks` | Create a probe task persisted in database |
| `jade_tree.api.v1.ProbeTask` | `PUT /v1/probe-tasks/{uid}` | Update a probe task fields (`type/host/port/url/name/timeoutSeconds`) |
| `jade_tree.api.v1.ProbeTask` | `PATCH /v1/probe-tasks/{uid}/status` | Manage probe task status (`ENABLED` / `DISABLED`) |
| `jade_tree.api.v1.ProbeTask` | `DELETE /v1/probe-tasks/{uid}` | Delete a probe task and remove it dynamically |
| `jade_tree.api.v1.ProbeTask` | `GET /v1/probe-tasks/{uid}` | Get one probe task |
| `jade_tree.api.v1.ProbeTask` | `GET /v1/probe-tasks` | List probe tasks (paginated) |

Probe metrics are configured via `bootstrap.probe` in `config/server.yaml` and exported via the existing `GET /metrics` endpoint.

All SSH command APIs require a logged-in JWT user (`contextx` user UID). OpenAPI for `SSHCommand` is generated to `internal/server/swagger/openapi.yaml`.

## Deployment

Jade Tree is designed for RPM + systemd deployment.

- systemd unit template: `deploy/systemd/jade-tree.service`
- RPM notes: `deploy/rpm/README.md`
- RPM packaging files: `packaging/rpm/`

## RPM Packaging

Local Linux build (requires `rpmbuild` and `rsync`):

```bash
make rpm
```

Local source RPM build:

```bash
make srpm
```

macOS / no local rpmbuild (recommended):

```bash
make rpm-docker
```

Artifacts:

- binary RPM: `rpmbuild/RPMS/<arch>/jade-tree-<version>-<release>.<arch>.rpm`
- source RPM: `rpmbuild/SRPMS/jade-tree-<version>-<release>.src.rpm`

## Run

```bash
make all
make dev
```

## CLI Machine Commands

- `jade_tree machine get [endpoint...]` calls `GET /v1/machine-info` on each target: one summary row per endpoint’s local machine (positional args, config `endpoints`, else `--endpoint` or config `endpoint`). Subcommands `get cpu|memory|network|disk|sys [endpoint...]` print the same detail sections for each endpoint (`sys`: architecture, OS, version, kernel).
- `jade_tree machine list [endpoint...]` calls `GET /v1/machine-infos` (paginated with `--page-size`) on each target and prints every machine that endpoint knows in storage. Subcommands `list cpu|memory|network|disk|sys [endpoint...]` print details for all those machines.
- `jade_tree machine push [endpoint...]` pushes local and stored machine data from `--from` (or config `endpoint`) to the given target endpoints.
- `jade_tree machine pull [endpoint...]` pulls machine data from `--from` (source endpoint) and stores it into the given target endpoints.
- All commands support `--output table|json|yaml` (default `table`).
- All machine commands support JWT auth via `--jwt` or config file.

Default client config file: `~/.jade_tree/client.yaml`

```yaml
endpoint: http://127.0.0.1:8000
endpoints:
  - http://10.0.0.11:8000
  - http://10.0.0.12:8000
jwt: eyJhbGciOi...
```

Flag values override config file values.

To regenerate GORM query code after changing models in `internal/data/impl/do/`:

```bash
make gen
```
