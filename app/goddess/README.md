# Goddess (嫦娥)

<div align="right">

[English](README.md) | [中文](README-zh_CN.md)

</div>

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Kratos](https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go)](https://github.com/go-kratos/kratos)
[![Cobra](https://img.shields.io/badge/Cobra-v1.10.2-00ADD8?style=flat&logo=go)](https://github.com/spf13/cobra)

## 📖 Introduction

Goddess (嫦娥) is a universal service template project for the Moon platform.

## Quick Start

```bash
make init
make build
```

### run the binary

- help

```bash
./bin/goddess -h
```

- version

```bash
./bin/goddess version
```

- run all

```bash
./bin/goddess run all -h
```

- run grpc

```bash
./bin/goddess run grpc -h
```

- run http

```bash
./bin/goddess run http -h
```

## Development

```bash
make init
make all
```

### run the application

- run all

```bash
go run . run all --log-level=DEBUG
```

- run grpc

```bash
go run . run grpc --log-level=DEBUG
```

- run http

```bash
go run . run http --log-level=DEBUG
```

## Acknowledgments

- [kratos](https://github.com/go-kratos/kratos)
- [cobra](https://github.com/spf13/cobra)