# Rabbit (玉兔)

<p align="center">
  <strong>Message and notification service for the Moon platform</strong>
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

**Rabbit** (玉兔) is the Moon platform’s message and notification service. It manages email configs, webhook configs, message templates, recipient groups, sends messages (email/webhook, raw or template-based), and provides message logs with retry/cancel.

- **Proto definitions**: `proto/rabbit/api/v1/`
- **HTTP + gRPC** via Kratos; CLI via Cobra.

---

## Features

- **Domain (Goddess)**: Namespace CRUD and select; Self (current user info, namespaces, change email/avatar/remark, refresh token); User (get/list/select/ban/permit); Member (list, invite, delete, status); Captcha (get captcha image); Auth (email login, send code)
- **Email**: SMTP config CRUD, list, select; send raw email or send with template
- **Webhook**: Webhook config CRUD, list, select (e.g. DingTalk, WeChat, Feishu, custom); send raw or template
- **Template**: Message template CRUD, list, select (EMAIL, SMS_ALICLOUD, WEBHOOK_*)
- **Recipient group**: Create/update/delete/list/select recipient groups; bind templates, email configs, webhooks, members; status control
- **Sender**: Send message by recipient group; send email by config; send webhook by config (with optional template)
- **Message log**: List logs (by time range, type, status); get single log; retry; cancel

---

## API Overview

| Service | Method / HTTP | Description |
|---------|----------------|-------------|
| **Namespace** | `POST /v1/namespace` | Create namespace |
| | `PUT /v1/namespace/{uid}` | Update namespace |
| | `PUT /v1/namespace/{uid}/status` | Update namespace status |
| | `DELETE /v1/namespace/{uid}` | Delete namespace |
| | `GET /v1/namespace/{uid}` | Get namespace |
| | `GET /v1/namespaces` | List namespaces |
| | `GET /v1/namespaces/select` | Select namespaces for dropdown |
| | `GET /v1/namespaces/simple` | Get namespace simple (allow unauthenticated) |
| **Self** | `GET /v1/self/info` | Get current user info |
| | `GET /v1/self/namespaces` | Get current user's namespaces |
| | `PUT /v1/self/change-email` | Change email |
| | `PUT /v1/self/change-avatar` | Change avatar |
| | `PUT /v1/self/change-remark` | Change remark |
| | `GET /v1/self/refresh-token` | Refresh token |
| **User** | `GET /v1/user/{uid}` | Get user |
| | `GET /v1/users` | List users |
| | `GET /v1/users/select` | Select users for dropdown |
| | `PUT /v1/user/ban/{uid}` | Ban user |
| | `PUT /v1/user/permit/{uid}` | Permit user |
| **Member** | `GET /v1/members` | List members |
| | `GET /v1/member/{uid}` | Get member |
| | `GET /v1/members/select` | Select members for dropdown |
| | `POST /v1/member/invite` | Invite member |
| | `DELETE /v1/member/{uid}` | Delete member |
| | `PUT /v1/member/{uid}/status` | Update member status |
| **Captcha** | `GET /v1/captcha` | Get captcha image (allow unauthenticated) |
| **Auth** | `POST /v1/auth/email/login/code` | Send email login code |
| | `POST /v1/auth/email/login` | Email login |
| **Email** | `POST /v1/email/config` | Create email config (SMTP) |
| | `PUT /v1/email/config/{uid}` | Update email config |
| | `PUT /v1/email/config/{uid}/status` | Update status (ENABLED/DISABLED) |
| | `DELETE /v1/email/config/{uid}` | Delete email config |
| | `GET /v1/email/config/{uid}` | Get email config |
| | `GET /v1/email/configs` | List email configs (pagination, keyword, status) |
| | `GET /v1/email/configs/select` | Select for dropdown |
| **Webhook** | `POST /v1/webhook/config` | Create webhook (app: OTHER/DINGTALK/WECHAT/FEISHU) |
| | `PUT /v1/webhook/config/{uid}` | Update webhook |
| | `PUT /v1/webhook/config/{uid}/status` | Update status |
| | `DELETE /v1/webhook/config/{uid}` | Delete webhook |
| | `GET /v1/webhook/config/{uid}` | Get webhook |
| | `GET /v1/webhook/configs` | List webhooks |
| | `GET /v1/webhook/configs/select` | Select for dropdown |
| **Template** | `POST /v1/template` | Create template (messageType, jsonData) |
| | `PUT /v1/template/{uid}` | Update template |
| | `PUT /v1/template/{uid}/status` | Update status |
| | `DELETE /v1/template/{uid}` | Delete template |
| | `GET /v1/template/{uid}` | Get template |
| | `GET /v1/templates` | List templates (pagination, keyword, status, messageType) |
| | `GET /v1/templates/select` | Select for dropdown |
| **RecipientGroupService** | `POST /v1/recipient-group` | Create group (name, templates, emailConfigs, webhookConfigs, members) |
| | `GET /v1/recipient-group/{uid}` | Get group (with templates, emailConfigs, webhookConfigs, members) |
| | `PUT /v1/recipient-group/{uid}` | Update group |
| | `PUT /v1/recipient-group/{uid}/status` | Update status |
| | `DELETE /v1/recipient-group/{uid}` | Delete group |
| | `GET /v1/recipient-groups` | List groups |
| | `GET /v1/recipient-groups/select` | Select for dropdown |
| **Sender** | `POST /v1/sender/message` | Send message by recipient group uid |
| | `POST /v1/sender/email/{uid}` | Send raw email (uid = email config; body, to, cc, etc.) |
| | `POST /v1/sender/email/{uid}/template` | Send email with template (templateUID, jsonData, to, cc) |
| | `POST /v1/sender/webhook/{uid}` | Send raw webhook (uid = webhook config; data) |
| | `POST /v1/sender/webhook/{uid}/template` | Send webhook with template (templateUID, jsonData) |
| **MessageLog** | `GET /v1/message-log/{uid}` | Get message log |
| | `GET /v1/message-logs` | List logs (page, pageSize, status, messageType, startAtUnix, endAtUnix; max 31 days range) |
| | `PUT /v1/message-log/{uid}/retry` | Retry message |
| | `PUT /v1/message-log/{uid}/cancel` | Cancel message |

API is defined in `proto/rabbit/api/v1/` (e.g. `email.proto`, `webhook.proto`, `template.proto`, `recipient_group.proto`, `sender.proto`, `message_log.proto`, `job.proto`). OpenAPI can be generated via `make api`.

---

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## Installation

From the Moon repo root or from this directory:

```bash
cd app/rabbit   # if at repo root
make init       # install protoc plugins, wire, etc.
make build      # generate API/conf/wire and build binary → bin/rabbit
```

---

## Quick Start

```bash
# Build
make init && make build

# Run all (HTTP + gRPC) in development
./bin/rabbit run all --log-level=DEBUG
# or
make dev
```

---

## Common Usage

### CLI

```bash
# Help
./bin/rabbit -h
./bin/rabbit version
./bin/rabbit run all -h
./bin/rabbit run grpc -h
./bin/rabbit run http -h
```

### Run modes

| Command | Description |
|--------|-------------|
| `./bin/rabbit run all` | Run both HTTP and gRPC servers |
| `./bin/rabbit run http` | HTTP only |
| `./bin/rabbit run grpc` | gRPC only |

### Make targets

| Target | Description |
|--------|-------------|
| `make init` | Install protoc plugins, wire, kratos, etc. |
| `make conf` | Generate config from proto |
| `make api` | Generate Go/HTTP/gRPC/OpenAPI from `proto/rabbit` |
| `make wire` | Generate Wire DI |
| `make all` | api + conf + wire |
| `make build` | all + build binary to `bin/rabbit` |
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

3. **Config**: See `internal/conf/`; set config path via flag or env.

---

## License

[MIT](LICENSE)

---

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
