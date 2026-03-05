# Moon

<p align="center">
  <strong>Go backend platform for aide-family — monorepo of shared tooling and microservices</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="app/goddess/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="https://github.com/go-kratos/kratos"><img src="https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go" alt="Kratos"></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/Cobra-v1.10-00ADD8?style=flat&logo=go" alt="Cobra"></a>
</p>

---

## Table of Contents

- [About](#about)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## About

**Moon** is a [Go](https://go.dev/) backend platform maintained by [aide-family](https://github.com/aide-family). This repository is a **monorepo** that includes:

- A shared toolkit and library used by all applications
- Microservices built with [Kratos](https://github.com/go-kratos/kratos) and [Cobra](https://github.com/spf13/cobra)

Each application has its own `go.mod`, build pipeline, and documentation (English & 中文).

---

## Features

- **Monorepo** — One repository for shared code and multiple services
- **Kratos** — Consistent service framework (HTTP/gRPC, config, logging)
- **Cobra** — CLI and subcommands for each app
- **Magic Box** — Shared utilities (safety, OAuth, validation, etc.) used across apps
- **Bilingual docs** — README and docs in English and 中文 for root and sub-projects

---

## Prerequisites

- [Go](https://go.dev/) **1.25+**
- [Make](https://www.gnu.org/software/make/)

---

## Installation

```bash
git clone https://github.com/aide-family/moon.git
cd moon
```

No global install is required at the repo root. Each app is built and run from its own directory (see [Quick Start](#quick-start) and [Documentation](#documentation)).

---

## Quick Start

From the repository root:

**Build all apps:**

```bash
make all
```

**Run a single app in development mode:**

```bash
make <app>
```

Where `<app>` is one of: `goddess`, `rabbit`, `marksman`. Example:

```bash
make rabbit
```

**List all Make targets:**

```bash
make help
```

**Generate code (e.g. from protos) for all apps:**

```bash
make gen
```

For app-specific setup (config, database, etc.), see the README of each app under [Documentation](#documentation).

---

## Project Structure

| Path | Name | Description |
|------|------|-------------|
| [`magicbox/`](magicbox/) | Magic Box (月光宝盒) | Shared toolkit and library (safety, OAuth, validation, etc.) |
| [`app/goddess/`](app/goddess/) | Goddess (嫦娥) | Authentication and authorization service |
| [`app/rabbit/`](app/rabbit/) | Rabbit (玉兔) | Business service (e.g. time engine and related features) |
| [`app/marksman/`](app/marksman/) | Marksman (后羿) | Event service |
| [`proto/`](proto/) | Protos | API definitions (e.g. goddess, rabbit) |

- **magicbox** is a Go module; **goddess**, **rabbit**, and **marksman** depend on it via `replace` in their `go.mod`.
- Each app has its own `go.mod`, Makefile, and README (EN / 中文).

---

## Documentation

| Project | English | 中文 |
|---------|---------|------|
| Goddess | [README](app/goddess/README.md) | [README](app/goddess/README-zh_CN.md) |
| Rabbit | [README](app/rabbit/README.md) | [README](app/rabbit/README-zh_CN.md) |
| Marksman | [README](app/marksman/README.md) | [README](app/marksman/README-zh_CN.md) |
| Magic Box | [README](magicbox/README.md) | [README](magicbox/README-zh_CN.md) |

---

## Contributing

Contributions are welcome. Please:

1. Open an [Issue](https://github.com/aide-family/moon/issues) for bugs or feature requests.
2. Fork the repo, create a branch, make your changes, and open a [Pull Request](https://github.com/aide-family/moon/pulls).

Ensure code and style match the existing project conventions.

---

## License

This project is under the **MIT License**. Each sub-project may have its own LICENSE file; see the following for details:

- [app/goddess/LICENSE](app/goddess/LICENSE)
- [app/rabbit/LICENSE](app/rabbit/LICENSE)
- [app/marksman/LICENSE](app/marksman/LICENSE)
- [magicbox/LICENSE](magicbox/LICENSE)

---

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos) — Go microservice framework
- [Cobra](https://github.com/spf13/cobra) — CLI library for Go
