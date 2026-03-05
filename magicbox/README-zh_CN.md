# Magic Box（月光宝盒）

<p align="center">
  <strong>Moon 平台共享开发工具库</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
</p>

---

## 目录

- [项目简介](#项目简介)
- [功能特性](#功能特性)
- [模块概览](#模块概览)
- [安装](#安装)
- [常用用法](#常用用法)
- [Proto 定义](#proto-定义)
- [许可证](#许可证)
- [致谢](#致谢)

---

## 项目简介

**Magic Box**（月光宝盒）是 Moon 各应用（Goddess、Rabbit、Marksman）共用的工具库，提供认证、OAuth、校验、安全容器、日志、服务端中间件、定时任务、验证码、编码等通用能力。`proto/magicbox/` 下的 Proto 定义被平台内多处引用（枚举、健康检查、oauth2、配置等）。

- **模块路径**：`github.com/aide-family/magicbox`
- **Go 版本**：1.25+

---

## 功能特性

- **认证与 OAuth**：JWT、Basic Auth、OAuth2（飞书、GitHub、Gitee 等）
- **安全容器**：Safe map/slice，并发安全访问
- **服务端**：HTTP 中间件（如校验）、Cron
- **校验**：请求/字段校验，供 Kratos 服务使用
- **验证码**：图形验证码生成与校验
- **编码**：JSON、YAML、压缩
- **日志**：Stdio、Sugared、GORM 日志
- **配置**：配置加载与工具
- **工具**：字符串（加密、常量）、定时（类 cron）、context、目录、httpx
- **插件**：数据源、缓存抽象
- **API/Proto**：健康检查、枚举、策略枚举、OAuth2、配置等 Proto

---

## 模块概览

| 包路径 | 说明 |
|--------|------|
| `auth` | 认证辅助（如 Basic Auth） |
| `auth/basic` | Basic 认证实现 |
| `captcha` | 验证码生成与校验 |
| `compress` | 压缩工具 |
| `config` | 配置加载与类型 |
| `contextx` | Context 工具 |
| `dir` | 目录工具 |
| `encoding/json` | JSON 编码工具 |
| `encoding/yaml` | YAML 编码工具 |
| `enum` | 通用枚举（状态、角色等） |
| `httpx` | HTTP 客户端工具 |
| `jwt` | JWT 生成/解析 |
| `log` | 日志（stdio、sugared、gormlog） |
| `oauth` | OAuth2；提供方：`feishu`、`github`、`gitee` |
| `password` | 密码哈希/校验 |
| `plugin/cache` | 缓存插件抽象 |
| `plugin/datasource` | 数据源插件抽象 |
| `safety` | Safe map/slice（并发安全、写时复制等） |
| `server/middler` | HTTP 中间件（如 validate） |
| `server/cron` | 定时任务 |
| `strutil` | 字符串工具（加密、常量） |
| `timer` | 定时（hour、day、week、month） |
| `api/v1` | API 类型（如 health） |

Proto 定义（供 Moon 内代码生成使用）：

- `proto/magicbox/api/v1/health.proto` — 健康检查
- `proto/magicbox/enum/enum.proto` — 全局枚举（状态、角色、用户/成员状态等）
- `proto/magicbox/enum/strategy.proto` — 策略相关枚举（采样模式、条件）
- `proto/magicbox/oauth/oauth2.proto` — OAuth2 登录请求/响应
- `proto/magicbox/config/config.proto` — 配置
- `proto/magicbox/merr/err.proto` — 错误定义

---

## 安装

**在 Moon 单仓内**（开发推荐）：Goddess、Rabbit、Marksman 通过 `go.mod` 的 `replace` 指向本地 magicbox：

```go
replace github.com/aide-family/magicbox => ../../magicbox
```

无需单独安装，在仓库根目录或 `app/<应用名>` 下构建/运行即可。

**作为独立依赖**（若模块已发布）：

```bash
go get github.com/aide-family/magicbox@latest
```

---

## 常用用法

### 导入与使用

```go
import (
    "github.com/aide-family/magicbox/safety"
    "github.com/aide-family/magicbox/jwt"
    "github.com/aide-family/magicbox/server/middler"
    "github.com/aide-family/magicbox/strutil"
)

// 安全 map（如写时复制或并发安全）
m := safety.NewMap[string, int]()

// JWT
token, err := jwt.NewToken(claims, secret, duration)

// 中间件（如 validate）一般在 Kratos HTTP 服务中注册
// 参见 app/goddess、app/rabbit、app/marksman 的服务端配置
```

### 使用枚举 / Proto 类型

各应用通过自己的 protoc 从 `proto/magicbox/` 生成 Go 代码。在应用内使用生成好的枚举与 message 类型（来自应用 pkg 或 magicbox 生成代码）。

---

## Proto 定义

| 路径 | 说明 |
|------|------|
| `proto/magicbox/api/v1/health.proto` | 健康检查 RPC/消息 |
| `proto/magicbox/enum/enum.proto` | GlobalStatus、UserStatus、MemberStatus、MemberRole、HTTPMethod、WebhookAPP、MessageType、MessageStatus、DatasourceType、DatasourceDriver 等 |
| `proto/magicbox/enum/strategy.proto` | SampleMode、ConditionMetric |
| `proto/magicbox/oauth/oauth2.proto` | OAuth2 登录请求/响应 |
| `proto/magicbox/config/config.proto` | 配置消息 |
| `proto/magicbox/merr/err.proto` | 错误定义 |

---

## 许可证

[MIT](LICENSE)

---

## 致谢

- [Kratos](https://github.com/go-kratos/kratos)
- [GORM](https://gorm.io/)
- [go-redis](https://github.com/redis/go-redis)
- [robfig/cron](https://github.com/robfig/cron/v3)
