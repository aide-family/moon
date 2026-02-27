# rabbit (玉兔)

<div align="right">

[English](README.md) | [中文](README-zh_CN.md)

</div>

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Kratos](https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go)](https://github.com/go-kratos/kratos)
[![Cobra](https://img.shields.io/badge/Cobra-v1.10.2-00ADD8?style=flat&logo=go)](https://github.com/spf13/cobra)

## 📖 Introduction

rabbit (玉兔) 

## Quick Start

```bash
make init
make build
```

### run the binary

- help

```bash
./bin/rabbit -h
```

- version

```bash
./bin/rabbit version
```

- run all

```bash
./bin/rabbit run all -h
```

- run grpc

```bash
./bin/rabbit run grpc -h
```

- run http

```bash
./bin/rabbit run http -h
```

## Development

```bash
make init
make all
```

### run the application

- run all

```bash
go run . run all
```

- run grpc

```bash
go run . run grpc
```

- run http

```bash
go run . run http
```

## Acknowledgments

- [kratos](https://github.com/go-kratos/kratos)
- [cobra](https://github.com/spf13/cobra)