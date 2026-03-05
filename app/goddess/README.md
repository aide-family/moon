# Goddess (嫦娥)

<p align="center">
  <strong>Universal authentication and authorization service for the Moon platform</strong>
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

**Goddess** (嫦娥) is the Moon platform’s authentication and authorization service. It provides login (OAuth2, email+code), captcha, user/namespace/member management, and self-service (profile, token refresh).

- **Proto definitions**: `proto/goddess/api/v1/`
- **HTTP + gRPC** via Kratos; CLI via Cobra.

---

## Features

- **Auth**: OAuth2 login, email login with verification code, token refresh
- **Captcha**: Get captcha (ID + base64 image) for login/register
- **Namespace**: Create/update/delete/list/select namespaces; status control; simple get by uid+secret (no auth)
- **User**: Get/list/select users; ban/permit
- **Member**: List/get/select members; invite, dismiss, update status (per namespace)
- **Self**: Current user info, namespaces, change email/avatar/remark, refresh token

---

## API Overview

| Service    | Method / HTTP | Description |
|-----------|----------------|-------------|
| **AuthService** | `OAuth2Login` | OAuth2 login (e.g. Feishu, GitHub) |
| | `POST /v1/auth/email/login/code` | Send email login code (requires captcha) |
| | `POST /v1/auth/email/login` | Email login with code |
| **Captcha** | `GET /v1/captcha` | Get captcha (captchaId, captchaB64s) |
| **Namespace** | `POST /v1/namespace` | Create namespace |
| | `PUT /v1/namespace/{uid}` | Update namespace |
| | `PUT /v1/namespace/{uid}/status` | Update namespace status (ENABLED/DISABLED) |
| | `DELETE /v1/namespace/{uid}` | Delete namespace |
| | `GET /v1/namespace/{uid}` | Get namespace (auth) |
| | `GET /v1/namespaces` | List namespaces (pagination, keyword, status) |
| | `GET /v1/namespaces/select` | Select namespaces (for dropdown) |
| | `GET /v1/namespaces/simple` | Get namespace by uid+secret (no auth) |
| **User** | `GET /v1/user/{uid}` | Get user |
| | `GET /v1/users` | List users (pagination, email, keyword, status) |
| | `GET /v1/users/select` | Select users (for dropdown) |
| | `PUT /v1/user/ban/{uid}` | Ban user |
| | `PUT /v1/user/permit/{uid}` | Permit user |
| **Member** | `GET /v1/members` | List members (pagination, keyword, status, etc.) |
| | `GET /v1/member/{uid}` | Get member |
| | `GET /v1/members/select` | Select members |
| | `POST /v1/member/invite` | Invite member (email, role) |
| | `DELETE /v1/member/{uid}` | Dismiss member |
| | `PUT /v1/member/{uid}/status` | Update member status |
| **Self** | `GET /v1/self/info` | Current user info |
| | `GET /v1/self/namespaces` | Current user’s namespaces |
| | `PUT /v1/self/change-email` | Change email |
| | `PUT /v1/self/change-avatar` | Change avatar |
| | `PUT /v1/self/change-remark` | Change remark |
| | `GET /v1/self/refresh-token` | Refresh token |

API is defined in `proto/goddess/api/v1/` (e.g. `auth.proto`, `namespace.proto`, `user.proto`, `member.proto`, `self.proto`, `captcha.proto`). OpenAPI/Swagger can be generated via `make api` (see Development).

---

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## Installation

From the Moon repo root or from this directory:

```bash
cd app/goddess   # if at repo root
make init        # install protoc plugins, wire, etc.
make build       # generate API/conf/wire and build binary → bin/goddess
```

---

## Quick Start

```bash
# Build (generates API, conf, wire and builds)
make init && make build

# Run all (HTTP + gRPC) in development
./bin/goddess run all --log-level=DEBUG
# or
make dev
```

---

## Common Usage

### CLI

```bash
# Help
./bin/goddess -h

# Version
./bin/goddess version

# Run subcommands help
./bin/goddess run all -h
./bin/goddess run grpc -h
./bin/goddess run http -h
```

### Run modes

| Command | Description |
|--------|-------------|
| `./bin/goddess run all` | Run both HTTP and gRPC servers |
| `./bin/goddess run http` | Run HTTP only |
| `./bin/goddess run grpc` | Run gRPC only |

### Make targets

| Target | Description |
|--------|-------------|
| `make init` | Install protoc plugins, wire, kratos, etc. |
| `make conf` | Generate config from proto |
| `make api` | Generate Go/HTTP/gRPC/OpenAPI from `proto/goddess` |
| `make wire` | Generate Wire DI |
| `make all` | api + conf + wire |
| `make build` | all + build binary to `bin/goddess` |
| `make dev` | `go run . run all --log-level=DEBUG` |
| `make gen` | Generate DO/data layer (e.g. test with generate tag) |
| `make test` | Run tests |
| `make clean` | Remove `bin/` |
| `make help` | List all targets |

---

## Development

1. **Generate code after proto changes**

   ```bash
   make api    # regenerate API and OpenAPI
   make wire   # if service graph changed
   ```

2. **Run without building**

   ```bash
   go run . run all --log-level=DEBUG
   go run . run http --log-level=DEBUG
   go run . run grpc --log-level=DEBUG
   ```

3. **Config**: See `internal/conf/`; config file path is typically set via flag or env.

---

## License

[MIT](LICENSE)

---

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
