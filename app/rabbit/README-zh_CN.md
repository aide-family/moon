# Rabbit（玉兔）

<p align="center">
  <strong>Moon 平台消息与通知服务</strong>
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

## 目录

- [项目简介](#项目简介)
- [功能特性](#功能特性)
- [接口概览](#接口概览)
- [环境要求](#环境要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [常用用法](#常用用法)
- [开发说明](#开发说明)
- [许可证](#许可证)
- [致谢](#致谢)

---

## 项目简介

**Rabbit**（玉兔）是 Moon 平台的消息与通知服务，负责邮件配置、Webhook 配置、消息模板、收件人组管理，以及邮件/Webhook 的发送（正文或模板），并提供消息发送记录与重试/取消。

- **接口定义**：`proto/rabbit/api/v1/`
- 基于 Kratos 提供 **HTTP + gRPC**，CLI 基于 Cobra。

---

## 功能特性

- **领域（Goddess）**：命名空间增删改查与选择；当前用户（Self）信息、命名空间列表、修改邮箱/头像/备注、刷新 Token；用户（User）查询/列表/选择/封禁/解封；成员（Member）列表、邀请、删除、状态；验证码（Captcha）获取；认证（邮箱登录、发送验证码）
- **邮件**：SMTP 配置增删改查、列表、下拉选择；发送原始邮件或按模板发送
- **Webhook**：Webhook 配置增删改查、列表、下拉选择（钉钉、企微、飞书、自定义等）；原始或模板发送
- **模板**：消息模板增删改查、列表、下拉选择（EMAIL、SMS_ALICLOUD、WEBHOOK_*）
- **收件人组**：创建/更新/删除/列表/选择；绑定模板、邮件配置、Webhook、成员；状态启用/禁用
- **发送**：按收件人组发送消息；按邮件配置发送；按 Webhook 配置发送（支持模板）
- **消息日志**：按时间范围、类型、状态列表；单条查询；重试、取消

---

## 接口概览

| 服务 | 方法 / HTTP | 说明 |
|------|-------------|------|
| **Namespace** | `POST /v1/namespace` | 创建命名空间 |
| | `PUT /v1/namespace/{uid}` | 更新命名空间 |
| | `PUT /v1/namespace/{uid}/status` | 更新命名空间状态 |
| | `DELETE /v1/namespace/{uid}` | 删除命名空间 |
| | `GET /v1/namespace/{uid}` | 获取命名空间 |
| | `GET /v1/namespaces` | 列表 |
| | `GET /v1/namespaces/select` | 下拉选择 |
| | `GET /v1/namespaces/simple` | 获取命名空间简要（可未登录） |
| **Self** | `GET /v1/self/info` | 获取当前用户信息 |
| | `GET /v1/self/namespaces` | 获取当前用户命名空间列表 |
| | `PUT /v1/self/change-email` | 修改邮箱 |
| | `PUT /v1/self/change-avatar` | 修改头像 |
| | `PUT /v1/self/change-remark` | 修改备注 |
| | `GET /v1/self/refresh-token` | 刷新 Token |
| **User** | `GET /v1/user/{uid}` | 获取用户 |
| | `GET /v1/users` | 用户列表 |
| | `GET /v1/users/select` | 用户下拉选择 |
| | `PUT /v1/user/ban/{uid}` | 封禁用户 |
| | `PUT /v1/user/permit/{uid}` | 解封用户 |
| **Member** | `GET /v1/members` | 成员列表 |
| | `GET /v1/member/{uid}` | 获取成员 |
| | `GET /v1/members/select` | 成员下拉选择 |
| | `POST /v1/member/invite` | 邀请成员 |
| | `DELETE /v1/member/{uid}` | 删除成员 |
| | `PUT /v1/member/{uid}/status` | 更新成员状态 |
| **Captcha** | `GET /v1/captcha` | 获取验证码图片（可未登录） |
| **Auth** | `POST /v1/auth/email/login/code` | 发送邮箱登录验证码 |
| | `POST /v1/auth/email/login` | 邮箱登录 |
| **Email** | `POST /v1/email/config` | 创建邮件配置（SMTP） |
| | `PUT /v1/email/config/{uid}` | 更新邮件配置 |
| | `PUT /v1/email/config/{uid}/status` | 更新状态（ENABLED/DISABLED） |
| | `DELETE /v1/email/config/{uid}` | 删除邮件配置 |
| | `GET /v1/email/config/{uid}` | 获取邮件配置 |
| | `GET /v1/email/configs` | 列表（分页、keyword、status，支持可选 `uids` 批量过滤） |
| | `GET /v1/email/configs/select` | 下拉选择 |
| **Webhook** | `POST /v1/webhook/config` | 创建 Webhook（app: OTHER/DINGTALK/WECHAT/FEISHU） |
| | `PUT /v1/webhook/config/{uid}` | 更新 Webhook |
| | `PUT /v1/webhook/config/{uid}/status` | 更新状态 |
| | `DELETE /v1/webhook/config/{uid}` | 删除 Webhook |
| | `GET /v1/webhook/config/{uid}` | 获取 Webhook |
| | `GET /v1/webhook/configs` | 列表（支持可选 `uids` 批量过滤） |
| | `GET /v1/webhook/configs/select` | 下拉选择 |
| **Template** | `POST /v1/template` | 创建模板（messageType、jsonData） |
| | `PUT /v1/template/{uid}` | 更新模板 |
| | `PUT /v1/template/{uid}/status` | 更新状态 |
| | `DELETE /v1/template/{uid}` | 删除模板 |
| | `GET /v1/template/{uid}` | 获取模板 |
| | `GET /v1/templates` | 列表（分页、keyword、status、messageType，支持可选 `uids` 批量过滤） |
| | `GET /v1/templates/select` | 下拉选择 |
| **RecipientGroupService** | `POST /v1/recipient-group` | 创建收件人组（name、templates、emailConfigs、webhookConfigs、members） |
| | `GET /v1/recipient-group/{uid}` | 获取组（含模板、邮件配置、Webhook、成员） |
| | `PUT /v1/recipient-group/{uid}` | 更新组 |
| | `PUT /v1/recipient-group/{uid}/status` | 更新状态 |
| | `DELETE /v1/recipient-group/{uid}` | 删除组 |
| | `GET /v1/recipient-groups` | 列表 |
| | `GET /v1/recipient-groups/select` | 下拉选择 |
| **Sender** | `POST /v1/sender/message` | 按收件人组 uid 发送消息 |
| | `POST /v1/sender/email/{uid}` | 发送原始邮件（uid 为邮件配置；body、to、cc 等） |
| | `POST /v1/sender/email/{uid}/template` | 按模板发送邮件（templateUID、jsonData、to、cc） |
| | `POST /v1/sender/webhook/{uid}` | 发送原始 Webhook（uid 为 Webhook 配置；data） |
| | `POST /v1/sender/webhook/{uid}/template` | 按模板发送 Webhook（templateUID、jsonData） |
| **MessageLog** | `GET /v1/message-log/{uid}` | 获取单条消息日志 |
| | `GET /v1/message-logs` | 列表（page、pageSize、status、messageType、startAtUnix、endAtUnix；时间范围最多 31 天） |
| | `PUT /v1/message-log/{uid}/retry` | 重试发送 |
| | `PUT /v1/message-log/{uid}/cancel` | 取消发送 |

接口定义位于 `proto/rabbit/api/v1/`（如 `email.proto`、`webhook.proto`、`template.proto`、`recipient_group.proto`、`sender.proto`、`message_log.proto`、`job.proto`）。可通过 `make api` 生成 OpenAPI。

---

## 环境要求

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## 安装

在 Moon 仓库根目录或本应用目录下执行：

```bash
cd app/rabbit   # 若在仓库根目录
make init       # 安装 protoc 插件、wire 等
make build      # 生成 API/conf/wire 并编译 → bin/rabbit
```

---

## 快速开始

```bash
# 编译
make init && make build

# 开发模式运行（HTTP + gRPC）
./bin/rabbit run all --log-level=DEBUG
# 或
make dev
```

---

## 常用用法

### 命令行

```bash
./bin/rabbit -h
./bin/rabbit version
./bin/rabbit run all -h
./bin/rabbit run grpc -h
./bin/rabbit run http -h
```

### 运行模式

| 命令 | 说明 |
|------|------|
| `./bin/rabbit run all` | 同时启动 HTTP 与 gRPC |
| `./bin/rabbit run http` | 仅 HTTP |
| `./bin/rabbit run grpc` | 仅 gRPC |

### Make 目标

| 目标 | 说明 |
|------|------|
| `make init` | 安装 protoc 插件、wire、kratos 等 |
| `make conf` | 从 proto 生成配置 |
| `make api` | 从 `proto/rabbit` 生成 Go/HTTP/gRPC/OpenAPI |
| `make wire` | 生成 Wire 依赖注入 |
| `make all` | api + conf + wire |
| `make build` | all + 编译到 `bin/rabbit` |
| `make dev` | `go run . run all --log-level=DEBUG` |
| `make gen` | 生成 DO/数据层（如带 generate tag 的测试） |
| `make clean` | 删除 `bin/` |
| `make help` | 列出所有目标 |

---

## 开发说明

1. **修改 proto 后重新生成**

   ```bash
   make api
   make wire   # 若依赖有变更
   ```

2. **不编译直接运行**

   ```bash
   go run . run all
   go run . run http
   go run . run grpc
   ```

3. **配置**：见 `internal/conf/`；配置文件路径通过启动参数或环境变量指定。

---

## 许可证

[MIT](LICENSE)

---

## 致谢

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
