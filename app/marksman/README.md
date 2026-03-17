# Marksman (后羿)

<p align="center">
  <strong>Event and alerting service for the Moon platform</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="https://github.com/go-kratos/kratos"><img src="https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go" alt="Kratos"></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/Cobra-v1.10-00ADD8?style=flat&logo=go" alt="Cobra"></a>
</p>

---

## Table of Contents

- [About](#about)
- [Features](#features)
- [API Overview](#api-overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Common Usage](#common-usage)
- [Development](#development)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## About

**Marksman** (后羿) is the Moon platform’s event and alerting service. It manages datasources (Prometheus, VictoriaMetrics, Elasticsearch, Jaeger), strategy groups and strategies, alert levels, and strategy metrics (expressions, levels, receivers binding).

- **Proto definitions**: `proto/marksman/api/v1/`
- **HTTP + gRPC** via Kratos; CLI via Cobra.

---

## Features

- **Self** (goddess): Current user info, namespaces list, change email/avatar/remark, refresh token
- **User** (goddess): User CRUD, list, select, ban/permit
- **Member** (goddess): List/get/select members in namespace, invite, dismiss, update status
- **Namespace** (goddess): Namespace management (shared with goddess; requires `namespaceConfig`)
- **Captcha** (goddess): Get graphic captcha (id, base64 image) for login and other unauthenticated flows
- **Datasource**: CRUD, list, select, status time series (per uid, from main TSDB), metric metadata (label names and label values) for metrics datasources (Prometheus, VictoriaMetrics, Elasticsearch, Jaeger)
- **Strategy group**: CRUD, list, select, status; bind receivers (recipient groups)
- **Strategy**: CRUD, list, status; link to strategy group; type (METRICS/LOGS/TRACE) and driver
- **Level**: Alert level CRUD, list, select, status (for grouping alert severity)
- **Strategy metric**: Save/get metric config (expr, labels, datasourceUIDs, levels); save/update/delete/get metric levels (mode, condition, values, duration); bind receivers per strategy (optional levelUID)
- **Alert (real-time)**: Alert page CRUD (name, color, sort order, filter by strategy group/level/strategy); list real-time alert events by alert page; operate events: intervene (on-call takeover), suppress (until time), recover (manual)

---

## API Overview

| Service | Method / HTTP | Description |
|---------|----------------|-------------|
| **Self** (goddess) | `GET /v1/self/info` | Current user info |
| | `GET /v1/self/namespaces` | Current user namespaces list |
| | `PUT /v1/self/change-email`, `change-avatar`, `change-remark` | Change email / avatar / remark |
| | `GET /v1/self/refresh-token` | Refresh token |
| **User** (goddess) | `GET /v1/user/{uid}` | Get user |
| | `GET /v1/users`, `GET /v1/users/select` | List users, select for dropdown |
| | `PUT /v1/user/ban/{uid}`, `PUT /v1/user/permit/{uid}` | Ban / permit user |
| **Member** (goddess) | `GET /v1/members`, `GET /v1/member/{uid}`, `GET /v1/members/select` | List members, get member, select |
| | `POST /v1/member/invite` | Invite member (email, role) |
| | `DELETE /v1/member/{uid}`, `PUT /v1/member/{uid}/status` | Dismiss member, update status |
| **Captcha** (goddess) | `GET /v1/captcha` | Get graphic captcha (returns captchaId, captchaB64s) |
| **Datasource** | `POST /v1/datasource` | Create datasource (name, type, driver, url, metadata, remark) |
| | `PUT /v1/datasource/{uid}` | Update datasource |
| | `DELETE /v1/datasource/{uid}` | Delete datasource |
| | `GET /v1/datasource/{uid}` | Get datasource |
| | `GET /v1/datasources` | List (keyword, page, pageSize, type, driver, status) |
| | `GET /v1/datasources/select` | Select for dropdown |
| | `GET /v1/datasource/{uid}/status` | Status time series for one datasource (from main TSDB; query: startTime, endTime, stepSeconds; default last 1h, step 60s) |
| | `GET /v1/datasource/{uid}/metrics` | List metrics (name, description, unit, type only; optional match[], limit; default limit 100) |
| | `GET /v1/datasource/{uid}/metric-label-detail` | One metric's label detail: labels + each label's values (query: metric=name) |
| **Strategy** (group) | `POST /v1/strategy-group` | Create strategy group |
| | `PUT /v1/strategy-group/{uid}` | Update strategy group |
| | `PUT /v1/strategy-group/{uid}/status` | Update status (ENABLED/DISABLED) |
| | `DELETE /v1/strategy-group/{uid}` | Delete strategy group |
| | `GET /v1/strategy-group/{uid}` | Get strategy group |
| | `GET /v1/strategy-groups` | List strategy groups |
| | `GET /v1/strategy-groups/select` | Select for dropdown |
| | `POST /v1/strategy-group/{uid}/receivers` | Bind receivers (receiverUIDs) to group |
| **Strategy** (item) | `POST /v1/strategy` | Create strategy (name, type, driver, strategyGroupUID, status, etc.) |
| | `PUT /v1/strategy/{uid}` | Update strategy |
| | `PUT /v1/strategy/{uid}/status` | Update status |
| | `DELETE /v1/strategy/{uid}` | Delete strategy |
| | `GET /v1/strategy/{uid}` | Get strategy |
| | `GET /v1/strategies` | List strategies (keyword, page, pageSize, status, strategyGroupUID, type, driver) |
| | `GET /v1/strategies/select` | Select strategies for dropdown (keyword, limit, lastUid, status, strategyGroupUids list to filter by groups) |
| **Level** | `POST /v1/level` | Create level (name, remark, metadata) |
| | `PUT /v1/level/{uid}` | Update level |
| | `PUT /v1/level/{uid}/status` | Update status |
| | `DELETE /v1/level/{uid}` | Delete level |
| | `GET /v1/level/{uid}` | Get level |
| | `GET /v1/levels` | List levels (page, pageSize, keyword, status) |
| | `GET /v1/levels/select` | Select for dropdown |
| **StrategyMetric** | `POST /v1/metric/strategy/{strategyUID}` | Save strategy metric (expr, labels, datasourceUIDs, summary, description, status) |
| | `GET /v1/metric/strategy/{strategyUID}` | Get strategy metric (with levels) |
| | `POST /v1/metric/strategy/{strategyUID}/level` | Save metric level (levelUID, mode, condition, values, duration, status) |
| | `PUT /v1/metric/strategy/{strategyUID}/level/{uid}/status` | Update metric level status |
| | `DELETE /v1/metric/strategy/{strategyUID}/level/{uid}` | Delete metric level |
| | `GET /v1/metric/strategy/{strategyUID}/level/{uid}` | Get metric level |
| | `POST /v1/metric/strategy/{strategyUID}/receivers` | Bind receivers (receiverUIDs; optional levelUID) |
| **Alert** (alert page) | `POST /v1/alert-pages` | Create alert page (name, color, sortOrder, filter) |
| | `PUT /v1/alert-pages/{uid}` | Update alert page |
| | `DELETE /v1/alert-pages/{uid}` | Delete alert page |
| | `GET /v1/alert-pages/{uid}` | Get alert page |
| | `GET /v1/alert-pages` | List alert pages (page, pageSize, keyword) |
| **Alert** (realtime) | `GET /v1/alert-pages/{alertPageUid}/realtime-alerts` | List real-time alert events for page (page, pageSize, status) |
| | `POST /v1/realtime-alerts/{uid}/intervene` | Mark event as intervened (on-call takeover) |
| | `POST /v1/realtime-alerts/{uid}/suppress` | Suppress event until time (body: suppressUntil RFC3339) |
| | `POST /v1/realtime-alerts/{uid}/recover` | Mark event as manually recovered |

**Types**: `DatasourceType`: METRICS, LOGS, TRACE. **Drivers**: METRICS_PROMETHEUS, METRICS_VICTORIA_METRICS, LOGS_ELASTICSEARCH, TRACE_JAEGER.

API definitions: Marksman-owned APIs are in `proto/marksman/api/v1/` (e.g. `datasource.proto`, `strategy.proto`, `level.proto`, `strategy_metric.proto`, `alert.proto`). Self, User, Member, Namespace, Captcha, etc. come from goddess at `proto/goddess/api/v1/`. OpenAPI can be generated via `make api`.

---

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## Installation

From the Moon repo root or from this directory:

```bash
cd app/marksman   # if at repo root
make init         # install protoc plugins, wire, etc.
make build        # generate API/conf/wire and build binary → bin/marksman
```

---

## Quick Start

```bash
# Build
make init && make build

# Run all (HTTP + gRPC) in development
./bin/marksman run all --log-level=DEBUG
# or
make dev
```

---

## Common Usage

### CLI

```bash
# Help
./bin/marksman -h
./bin/marksman version
./bin/marksman run all -h
./bin/marksman run grpc -h
./bin/marksman run http -h
```

### Run modes

| Command | Description |
|--------|-------------|
| `./bin/marksman run all` | Run both HTTP and gRPC servers |
| `./bin/marksman run http` | HTTP only |
| `./bin/marksman run grpc` | gRPC only |

### Make targets

| Target | Description |
|--------|-------------|
| `make init` | Install protoc plugins, wire, kratos, etc. |
| `make conf` | Generate config from proto |
| `make api` | Generate Go/HTTP/gRPC/OpenAPI from `proto/marksman` |
| `make wire` | Generate Wire DI |
| `make all` | api + conf + wire |
| `make build` | all + build binary to `bin/marksman` |
| `make dev` | `go run . run all --log-level=DEBUG` |
| `make gen` | Generate DO/data layer (e.g. test with generate tag) |
| `make clean` | Remove `bin/` |
| `make help` | List all targets |

---

## Development

1. **Regenerate after proto changes**

   ```bash
   make api
   make wire   # if service graph changed
   ```

2. **Run without building**

   ```bash
   go run . run all
   go run . run http
   go run . run grpc
   ```

3. **Config**: See `internal/conf/`; set config path via flag or env. For datasource status time series (`GET /v1/datasource/{uid}/status`), configure **mainTsdb** in `config/server.yaml`: `driver` (`prometheus` or `victoria_metrics`) and `url`; env overrides: `MARKSMAN_MAIN_TSDB_DRIVER`, `MARKSMAN_MAIN_TSDB_URL`.

---

## License

[MIT](LICENSE)

---

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
