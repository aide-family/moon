# Magic Box (月光宝盒)

<p align="center">
  <strong>Shared development toolkit and library for the Moon platform</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
</p>

---

## Table of Contents

- [About](#about)
- [Features](#features)
- [Module Overview](#module-overview)
- [Installation](#installation)
- [Common Usage](#common-usage)
- [Proto Definitions](#proto-definitions)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## About

**Magic Box** (月光宝盒) is the shared toolkit and library used by Moon applications (Goddess, Rabbit, Marksman). It provides common utilities for auth, OAuth, validation, safety helpers, logging, server middleware, cron, captcha, encoding, and more. Proto definitions under `proto/magicbox/` are used across the platform (enums, health, oauth2, config, etc.).

- **Module**: `github.com/aide-family/magicbox`
- **Go**: 1.25+

---

## Features

- **Auth & OAuth**: JWT, basic auth, OAuth2 (e.g. Feishu, GitHub, Gitee)
- **Safety**: Safe map/slice helpers, concurrent-safe access patterns
- **Server**: HTTP middleware (e.g. validate), cron
- **Validation**: Request/field validation used by Kratos services
- **Captcha**: Captcha generation and verification
- **Encoding**: JSON, YAML, compression
- **Logging**: Stdio, sugared, GORM logger
- **Config**: Configuration loading and helpers
- **Utils**: String (encrypt, constants), timer (cron-like), context, dir, httpx
- **Plugin**: Datasource, cache abstractions
- **API/Proto**: Health check, enums, strategy enums, OAuth2, config protos

---

## Module Overview

| Package | Description |
|---------|-------------|
| `auth` | Auth helpers (e.g. basic auth) |
| `auth/basic` | Basic auth implementation |
| `captcha` | Captcha generation/verification |
| `compress` | Compression utilities |
| `config` | Config loading and types |
| `contextx` | Context utilities |
| `dir` | Directory helpers |
| `encoding/json` | JSON encoding helpers |
| `encoding/yaml` | YAML encoding helpers |
| `enum` | Common enums (e.g. status, role) |
| `httpx` | HTTP client helpers |
| `jwt` | JWT create/parse |
| `log` | Logging (stdio, sugared, gormlog) |
| `oauth` | OAuth2; providers: `feishu`, `github`, `gitee` |
| `password` | Password hash/verify |
| `plugin/cache` | Cache plugin abstraction |
| `plugin/datasource` | Datasource plugin abstraction |
| `safety` | Safe map/slice (concurrent-safe, copy-on-write style) |
| `server/middler` | HTTP middleware (e.g. validate) |
| `server/cron` | Cron / scheduled jobs |
| `strutil` | String utilities (encrypt, constants) |
| `timer` | Timer helpers (hour, day, week, month) |
| `api/v1` | API types (e.g. health) |

Proto definitions (used by codegen across Moon):

- `proto/magicbox/api/v1/health.proto` — Health check
- `proto/magicbox/enum/enum.proto` — Global enums (status, role, user/member status, etc.)
- `proto/magicbox/enum/strategy.proto` — Strategy-related enums (sample mode, condition)
- `proto/magicbox/oauth/oauth2.proto` — OAuth2 login request/reply
- `proto/magicbox/config/config.proto` — Config
- `proto/magicbox/merr/err.proto` — Error types

---

## Installation

**Inside the Moon monorepo** (recommended for development): Goddess, Rabbit, and Marksman use a `replace` in their `go.mod` to point to the local magicbox:

```go
replace github.com/aide-family/magicbox => ../../magicbox
```

No separate install step; just build/run the app from the repo root or from `app/<name>`.

**As a standalone dependency** (if the module is published):

```bash
go get github.com/aide-family/magicbox@latest
```

---

## Common Usage

### Import and use packages

```go
import (
    "github.com/aide-family/magicbox/safety"
    "github.com/aide-family/magicbox/jwt"
    "github.com/aide-family/magicbox/server/middler"
    "github.com/aide-family/magicbox/strutil"
)

// Safe map (e.g. copy-on-write or concurrent-safe)
m := safety.NewMap[string, int]()

// JWT
token, err := jwt.NewToken(claims, secret, duration)

// Middleware (e.g. validate) is typically registered with Kratos HTTP server
// See app/goddess, app/rabbit, app/marksman for server setup
```

### Using enums / proto types

Apps generate Go code from `proto/magicbox/` (via their own protoc invocations). Use the generated enum and message types from the app’s `pkg` or from magicbox if the app imports magicbox’s generated code.

---

## Proto Definitions

| Path | Description |
|------|-------------|
| `proto/magicbox/api/v1/health.proto` | Health check RPC/messages |
| `proto/magicbox/enum/enum.proto` | GlobalStatus, UserStatus, MemberStatus, MemberRole, HTTPMethod, WebhookAPP, MessageType, MessageStatus, DatasourceType, DatasourceDriver, SSHCommandAuditKind, SSHCommandAuditStatus, etc. |
| `proto/magicbox/enum/strategy.proto` | SampleMode, ConditionMetric |
| `proto/magicbox/oauth/oauth2.proto` | OAuth2 login request/reply |
| `proto/magicbox/config/config.proto` | Config messages |
| `proto/magicbox/merr/err.proto` | Error definitions |

---

## License

[MIT](LICENSE)

---

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos)
- [GORM](https://gorm.io/)
- [go-redis](https://github.com/redis/go-redis)
- [robfig/cron](https://github.com/robfig/cron/v3)
