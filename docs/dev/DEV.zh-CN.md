<div style="display: flex; align-items: center;">
  <img 
    src="../images/logo.svg" 
    alt="Logo" 
    style="height: 4em; width: auto; vertical-align: middle; margin-right: 10px;" 
  />
  <h1 style="margin: 0; font-size: 24px; line-height: 1.5;">Moon - 多领域监控告警平台</h1>
</div>

| [English](DEV.md) | [简体中文](DEV.zh-CN.md) |

# 写在前面

在开发前，请确保你已经阅读了[README.md](../README.zh-CN.md), 以及[GOPHER.md](./GOPHER.zh-CN.md), 前者帮助你了解`Moon`,
后者给你一些关于此项目的编程经验和团队规范

# 环境篇

## Window make 环境
1. 下载 w64devkit
   访问官方发布页：https://github.com/skeeto/w64devkit/releases
   下载最新版本的 w64devkit-*.zip（例如 w64devkit-1.20.0.zip）。
2. 解压到本地目录
   将 ZIP 文件解压到自定义目录（例如 C:\w64devkit）。
3. 配置环境变量
   将 w64devkit 的 bin 目录添加到系统环境变量：
   右键 此电脑 → 属性 → 高级系统设置 → 环境变量。
   在 系统变量 中找到 Path → 点击 编辑 → 添加路径：C:\w64devkit\bin（根据实际解压路径调整）。
4. 验证安装
   打开命令提示符，运行：
   ```bash
   make --version # 应输出类似 "GNU Make 4.4"
   ```
> 如果出现编码异常问题， 参考下面两步， 没有则忽略
   1. 修改控制台编码
   ```bash
   reg add HKEY_CURRENT_USER\Console /v CodePage /t REG_DWORD /d 65001 /f
   ```

   2. 启用系统级 UTF-8 支持
   -  设置 → 时间和语言 → 语言 → 管理语言设置 → 勾选 "Beta: Unicode UTF-8" → 重启

**推荐使用 w64devkit 的原因：**
1. **原生编译支持**：w64devkit 提供完整的 Windows 原生 GNU 工具链（含 `make`、`gcc` 等），能直接适配 Makefile 的编译需求。
2. **环境一致性**：集成 MinGW-w64 工具链，避免因环境差异导致的编译问题。
3. **工具链完整性**：内置编译调试工具，无需额外安装其他依赖。

**其他方案的潜在问题：**
- 使用 Chocolatey 等包管理器安装的 `make` 可能存在：
    - **路径冲突**：与 Windows 原生工具链不兼容
    - **依赖缺失**：缺少必要的 C/C++ 编译环境
    - **版本适配问题**：工具链版本与项目要求不匹配


## 1. protoc 安装

### MacOS

```bash
# 安装
brew install protobuf
# 验证
protoc --version
```

### Windows

1. 下载 protoc 编译器
   访问官方发布页：https://github.com/protocolbuffers/protobuf/releases
   找到 protoc-x.x.x-win64.zip（例如 protoc-25.1-win64.zip），下载最新版本。

2. 解压并配置路径
   解压下载的 ZIP 文件到自定义目录（例如 C:\protoc）。
   将 protoc.exe 的路径添加到系统环境变量：
   右键 此电脑 → 属性 → 高级系统设置 → 环境变量。
   在 系统变量 中找到 Path → 点击 编辑 → 添加解压目录的路径（例如 C:\protoc\bin）。

3. 验证安装
   打开命令提示符，运行：

```bash
protoc --version  # 应输出类似 "libprotoc 25.1"
```

### Linux

**Ubuntu/Debian：**

```bash
# 安装
sudo apt-get update
sudo apt-get install protobuf-compiler
# 验证
protoc --version
```

**Fedora/CentOS/RHEL：**

```bash
# 安装
sudo yum install protobuf-compiler
# 验证
protoc --version
```

## 2. 项目环境依赖

### Go版本

* Go版本 >= 1.24.1

### 相关插件安装

```bash
make init
```

### 项目初始化

```bash
make all app=<palace|houyi|rabbit>
```

# 开发篇

本项目采用mini [DDD](https://www.google.com/search?q=DDD) 设计思想，总共分为以下几个模块

* API: 接口层，包含了HTTP、gRPC等接口定义
    * proto
    * pb
* service: 服务入口，承接了API(Server)和biz层，通常做一些参数校验、转换等工作
* biz: 业务逻辑层，包含了业务逻辑的实现
    * bo
    * do
    * repository
    * vobj
* data: 数据访问层，负责包括数据库、外部数据和外部服务等资源的处理
    * cache
    * db
    * impl

## 依赖倒置

正确的调用关系是：

`service` -> `biz` -> `repository` -> `impl` -> `data`

需要注意的是，每一层都会涉及到数据结构的转置，也就是每层只关心自己收到的数据结构，如：

- `service` 层只关心 `pb` 包中的数据结构，也就是 `proto` 生成的 `req` 和 `resp`，当 `service` 调用 `biz` 时，会将 `pb`
  转成 `bo`

## 事务管理

> biz 层使用事务管理

* repository 定义

```go
type Transaction interface {
  MainExec(ctx context.Context, fn func (ctx context.Context) error) error
  BizExec(ctx context.Context, fn func (ctx context.Context) error) error
  EventExec(ctx context.Context, fn func (ctx context.Context) error) error
}
```
