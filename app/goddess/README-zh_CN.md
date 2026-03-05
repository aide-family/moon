# Goddess（嫦娥）

<p align="center">
  <strong>Moon 平台通用认证与授权服务</strong>
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

**Goddess**（嫦娥）是 Moon 平台的认证与授权服务，提供登录（OAuth2、邮箱+验证码）、图形验证码、用户/空间/成员管理以及当前用户信息与 token 刷新等能力。

- **接口定义**：`proto/goddess/api/v1/`
- 基于 Kratos 提供 **HTTP + gRPC**，CLI 基于 Cobra。

---

## 功能特性

- **认证**：OAuth2 登录、邮箱验证码登录、Token 刷新
- **验证码**：获取图形验证码（captchaId + base64 图片），用于登录/注册
- **空间（Namespace）**：创建/更新/删除/列表/下拉选择；状态启用/禁用；支持按 uid+secret 免鉴权查询
- **用户（User）**：查询/列表/下拉选择；封禁/解封
- **成员（Member）**：列表/查询/下拉选择；邀请、移除、更新状态（按空间）
- **当前用户（Self）**：当前用户信息、所属空间、修改邮箱/头像/备注、刷新 Token

---

## 接口概览

| 服务 | 方法 / HTTP | 说明 |
|------|-------------|------|
| **AuthService** | `OAuth2Login` | OAuth2 登录（如飞书、GitHub） |
| | `POST /v1/auth/email/login/code` | 发送邮箱登录验证码（需先验证图形验证码） |
| | `POST /v1/auth/email/login` | 邮箱+验证码登录 |
| **Captcha** | `GET /v1/captcha` | 获取图形验证码（captchaId、captchaB64s） |
| **Namespace** | `POST /v1/namespace` | 创建空间 |
| | `PUT /v1/namespace/{uid}` | 更新空间 |
| | `PUT /v1/namespace/{uid}/status` | 更新空间状态（ENABLED/DISABLED） |
| | `DELETE /v1/namespace/{uid}` | 删除空间 |
| | `GET /v1/namespace/{uid}` | 获取空间（需鉴权） |
| | `GET /v1/namespaces` | 分页列表（支持 keyword、status） |
| | `GET /v1/namespaces/select` | 下拉选择 |
| | `GET /v1/namespaces/simple` | 按 uid+secret 查询（免鉴权） |
| **User** | `GET /v1/user/{uid}` | 获取用户 |
| | `GET /v1/users` | 用户列表（分页、email、keyword、status） |
| | `GET /v1/users/select` | 用户下拉选择 |
| | `PUT /v1/user/ban/{uid}` | 封禁用户 |
| | `PUT /v1/user/permit/{uid}` | 解封用户 |
| **Member** | `GET /v1/members` | 成员列表（分页、keyword、status 等） |
| | `GET /v1/member/{uid}` | 获取成员 |
| | `GET /v1/members/select` | 成员下拉选择 |
| | `POST /v1/member/invite` | 邀请成员（邮箱、角色） |
| | `DELETE /v1/member/{uid}` | 移除成员 |
| | `PUT /v1/member/{uid}/status` | 更新成员状态 |
| **Self** | `GET /v1/self/info` | 当前用户信息 |
| | `GET /v1/self/namespaces` | 当前用户所属空间 |
| | `PUT /v1/self/change-email` | 修改邮箱 |
| | `PUT /v1/self/change-avatar` | 修改头像 |
| | `PUT /v1/self/change-remark` | 修改备注 |
| | `GET /v1/self/refresh-token` | 刷新 Token |

接口定义位于 `proto/goddess/api/v1/`（如 `auth.proto`、`namespace.proto`、`user.proto`、`member.proto`、`self.proto`、`captcha.proto`）。可通过 `make api` 生成 OpenAPI/Swagger（见开发说明）。

---

## 环境要求

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## 安装

在 Moon 仓库根目录或本应用目录下执行：

```bash
cd app/goddess   # 若在仓库根目录
make init        # 安装 protoc 插件、wire 等
make build       # 生成 API/conf/wire 并编译 → bin/goddess
```

---

## 快速开始

```bash
# 编译（生成 API、conf、wire 并构建）
make init && make build

# 开发模式运行（HTTP + gRPC）
./bin/goddess run all --log-level=DEBUG
# 或
make dev
```

---

## 常用用法

### 命令行

```bash
# 帮助
./bin/goddess -h

# 版本
./bin/goddess version

# 子命令帮助
./bin/goddess run all -h
./bin/goddess run grpc -h
./bin/goddess run http -h
```

### 运行模式

| 命令 | 说明 |
|------|------|
| `./bin/goddess run all` | 同时启动 HTTP 与 gRPC |
| `./bin/goddess run http` | 仅 HTTP |
| `./bin/goddess run grpc` | 仅 gRPC |

### Make 目标

| 目标 | 说明 |
|------|------|
| `make init` | 安装 protoc 插件、wire、kratos 等 |
| `make conf` | 从 proto 生成配置 |
| `make api` | 从 `proto/goddess` 生成 Go/HTTP/gRPC/OpenAPI |
| `make wire` | 生成 Wire 依赖注入 |
| `make all` | api + conf + wire |
| `make build` | all + 编译到 `bin/goddess` |
| `make dev` | `go run . run all --log-level=DEBUG` |
| `make gen` | 生成 DO/数据层（如带 generate tag 的测试） |
| `make test` | 运行测试 |
| `make clean` | 删除 `bin/` |
| `make help` | 列出所有目标 |

---

## 开发说明

1. **修改 proto 后重新生成代码**

   ```bash
   make api   # 重新生成 API 与 OpenAPI
   make wire  # 若服务依赖有变更
   ```

2. **不编译直接运行**

   ```bash
   go run . run all --log-level=DEBUG
   go run . run http --log-level=DEBUG
   go run . run grpc --log-level=DEBUG
   ```

3. **配置**：见 `internal/conf/`；配置文件路径一般通过启动参数或环境变量指定。

---

## 许可证

[MIT](LICENSE)

---

## 致谢

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
