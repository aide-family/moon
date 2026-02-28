# Moon

<div align="right">

[English](README.md) | [中文](README-zh_CN.md)

</div>

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Kratos](https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go)](https://github.com/go-kratos/kratos)

## Introduction

**Moon** is a Go-based backend platform by [aide-family](https://github.com/aide-family). It is a monorepo containing shared tooling and microservices built with [Kratos](https://github.com/go-kratos/kratos) and [Cobra](https://github.com/spf13/cobra).

## Project Structure

| Path | Name | Description |
|------|------|-------------|
| [`magicbox/`](magicbox/) | Magic Box | Common development toolkit and shared library used by apps |
| [`app/goddess/`](app/goddess/) | Goddess (嫦娥) | Universal authentication and authorization service |
| [`app/rabbit/`](app/rabbit/) | Rabbit (玉兔) | Business service (e.g. time engine and related features) |

- **magicbox** is a Go module consumed by both **goddess** and **rabbit** via `replace` in their `go.mod`.
- Each app has its own `go.mod`, Makefile, and detailed README (with EN/ZH).

## Prerequisites

- [Go](https://golang.org/) 1.25+
- [Make](https://www.gnu.org/software/make/)

## Quick Start

From the repository root:

```bash
# Run Goddess (auth service) in development mode
make goddess

# Run Rabbit in development mode
make rabbit

# Show available make targets
make help
```

To build and run a specific app, see its directory:

```bash
# Goddess
cd app/goddess && make init && make build && ./bin/goddess run all -h

# Rabbit
cd app/rabbit && make init && make build && ./bin/rabbit run all -h
```

## Development

1. **Clone and enter the repo**

   ```bash
   git clone https://github.com/aide-family/moon.git
   cd moon
   ```

2. **Per-app setup**

   Each application manages its own toolchain and config. For full build and run instructions, see:

   - [Goddess README](app/goddess/README.md) (EN) / [中文](app/goddess/README-zh_CN.md)
   - [Rabbit README](app/rabbit/README.md) (EN) / [中文](app/rabbit/README-zh_CN.md)
   - [Magic Box README](magicbox/README.md)

3. **Root Makefile**

   The root [Makefile](Makefile) provides shortcuts to run apps in dev mode; run `make help` for the list of targets.

## License

This project is under the MIT License. See the [LICENSE](app/goddess/LICENSE) file in the respective sub-projects for details.

## Acknowledgments

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
