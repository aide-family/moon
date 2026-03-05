# Moon

<p align="center">
  <strong>aide-family 的 Go 后端平台 — 共享工具与微服务单仓库</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="app/goddess/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go 版本"></a>
  <a href="https://github.com/go-kratos/kratos"><img src="https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go" alt="Kratos"></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/Cobra-v1.10-00ADD8?style=flat&logo=go" alt="Cobra"></a>
</p>

---

## 目录

- [项目简介](#项目简介)
- [特性](#特性)
- [环境要求](#环境要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [文档](#文档)
- [参与贡献](#参与贡献)
- [许可证](#许可证)
- [致谢](#致谢)

---

## 项目简介

**Moon** 是由 [aide-family](https://github.com/aide-family) 维护的 [Go](https://go.dev/) 后端平台。本仓库采用 **单仓库（monorepo）** 结构，包含：

- 供各应用共用的共享工具库
- 基于 [Kratos](https://github.com/go-kratos/kratos) 与 [Cobra](https://github.com/spf13/cobra) 构建的微服务

每个应用拥有独立的 `go.mod`、构建流程及文档（英文与中文）。

---

## 特性

- **单仓库** — 共享代码与多服务统一管理
- **Kratos** — 统一的微服务框架（HTTP/gRPC、配置、日志等）
- **Cobra** — 各应用的 CLI 与子命令
- **Magic Box** — 跨应用共享的工具集（安全、OAuth、校验等）
- **中英双语文档** — 根目录与子项目均提供英文与中文 README

---

## 环境要求

- [Go](https://go.dev/) **1.25+**
- [Make](https://www.gnu.org/software/make/)

---

## 安装

```bash
git clone https://github.com/aide-family/moon.git
cd moon
```

根目录无需全局安装。各应用在其自身目录内完成构建与运行（见 [快速开始](#快速开始) 与 [文档](#文档)）。

---

## 快速开始

在仓库根目录执行：

**构建所有应用：**

```bash
make all
```

**以开发模式运行单个应用：**

```bash
make <应用名>
```

其中 `<应用名>` 可为：`goddess`、`rabbit`、`marksman`。示例：

```bash
make rabbit
```

**查看所有 Make 目标：**

```bash
make help
```

**为所有应用生成代码（如基于 proto）：**

```bash
make gen
```

各应用的配置、数据库等具体说明见 [文档](#文档) 中对应 README。

---

## 项目结构

| 路径 | 名称 | 说明 |
|------|------|------|
| [`magicbox/`](magicbox/) | Magic Box（月光宝盒） | 共享工具库（安全、OAuth、校验等） |
| [`app/goddess/`](app/goddess/) | Goddess（嫦娥） | 认证与授权服务 |
| [`app/rabbit/`](app/rabbit/) | Rabbit（玉兔） | 业务服务（如时间引擎及相关能力） |
| [`app/marksman/`](app/marksman/) | Marksman（后羿） | 事件服务 |
| [`proto/`](proto/) | Protos | API 定义（如 goddess、rabbit） |

- **magicbox** 为独立 Go 模块；**goddess**、**rabbit**、**marksman** 通过各自 `go.mod` 中的 `replace` 引用本地 magicbox。
- 各应用拥有独立的 `go.mod`、Makefile 及 README（中/英）。

---

## 文档

| 项目 | 英文 | 中文 |
|------|------|------|
| Goddess | [README](app/goddess/README.md) | [README](app/goddess/README-zh_CN.md) |
| Rabbit | [README](app/rabbit/README.md) | [README](app/rabbit/README-zh_CN.md) |
| Marksman | [README](app/marksman/README.md) | [README](app/marksman/README-zh_CN.md) |
| Magic Box | [README](magicbox/README.md) | [README](magicbox/README-zh_CN.md) |

---

## 参与贡献

欢迎贡献代码与文档。请：

1. 通过 [Issue](https://github.com/aide-family/moon/issues) 提交 Bug 或功能建议。
2. Fork 本仓库，创建分支，完成修改后提交 [Pull Request](https://github.com/aide-family/moon/pulls)。

请保持与现有项目风格与约定一致。

---

## 许可证

本项目采用 **MIT 许可证**。各子项目可能拥有独立的 LICENSE 文件，详见：

- [app/goddess/LICENSE](app/goddess/LICENSE)
- [app/rabbit/LICENSE](app/rabbit/LICENSE)
- [app/marksman/LICENSE](app/marksman/LICENSE)
- [magicbox/LICENSE](magicbox/LICENSE)

---

## 致谢

- [Kratos](https://github.com/go-kratos/kratos) — Go 微服务框架
- [Cobra](https://github.com/spf13/cobra) — Go CLI 库
