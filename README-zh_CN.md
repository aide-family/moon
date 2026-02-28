# Moon

<div align="right">

[English](README.md) | [中文](README-zh_CN.md)

</div>

[![许可证](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go 版本](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Kratos](https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go)](https://github.com/go-kratos/kratos)

## 项目介绍

**Moon** 是 [aide-family](https://github.com/aide-family) 下的 Go 后端平台，采用单仓库（monorepo）结构，包含通用工具库与基于 [Kratos](https://github.com/go-kratos/kratos) 和 [Cobra](https://github.com/spf13/cobra) 构建的微服务。

## 项目结构

| 路径 | 名称 | 说明 |
|------|------|------|
| [`magicbox/`](magicbox/) | Magic Box | 通用开发工具库，被各应用引用 |
| [`app/goddess/`](app/goddess/) | Goddess（嫦娥） | 通用认证与授权服务 |
| [`app/rabbit/`](app/rabbit/) | Rabbit（玉兔） | 业务服务（如时间引擎等相关能力） |

- **magicbox** 为独立 Go 模块，**goddess** 与 **rabbit** 通过各自 `go.mod` 中的 `replace` 引用本地 magicbox。
- 各应用拥有独立的 `go.mod`、Makefile 及详细 README（中英双语）。

## 环境要求

- [Go](https://golang.org/) 1.25+
- [Make](https://www.gnu.org/software/make/)

## 快速开始

在仓库根目录执行：

```bash
# 以开发模式运行 Goddess（认证服务）
make goddess

# 以开发模式运行 Rabbit
make rabbit

# 查看所有 make 目标
make help
```

若需在子目录中单独构建与运行，可进入对应应用目录操作：

```bash
# Goddess
cd app/goddess && make init && make build && ./bin/goddess run all -h

# Rabbit
cd app/rabbit && make init && make build && ./bin/rabbit run all -h
```

## 开发说明

1. **克隆并进入仓库**

   ```bash
   git clone https://github.com/aide-family/moon.git
   cd moon
   ```

2. **按应用进行配置**

   各应用的完整构建与运行方式见其目录下的 README：

   - [Goddess README](app/goddess/README.md)（英文）/ [中文](app/goddess/README-zh_CN.md)
   - [Rabbit README](app/rabbit/README.md)（英文）/ [中文](app/rabbit/README-zh_CN.md)
   - [Magic Box README](magicbox/README.md)

3. **根目录 Makefile**

   根目录 [Makefile](Makefile) 提供各应用的开发模式快捷命令，执行 `make help` 可查看所有目标。

## 许可证

本项目采用 MIT 许可证。各子项目目录下可能有独立的 LICENSE 文件，以具体文件为准。

## 致谢

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
