<div style="display: flex; align-items: center;">
  <img 
    src="../images/logo.svg" 
    alt="Logo" 
    style="height: 4em; width: auto; vertical-align: middle; margin-right: 10px;" 
  />
  <h1 style="margin: 0; font-size: 24px; line-height: 1.5;">Moon - Multi-Domain Monitoring and Alerting Platform</h1>
</div>

| [English](DEV.md) | [简体中文](DEV.zh-CN.md) |

# Preface

Before starting development, please ensure you have read [README.md](../../README.md) and [GOPHER.md](GOPHER.md). The
former helps you understand `Moon`, while the latter provides programming experience and team guidelines for this
project.

# Environment Setup

## Windows Make Environment

1. **Download w64devkit**  
   Visit releases page: [https://github.com/skeeto/w64devkit/releases](https://github.com/skeeto/w64devkit/releases)  
   Download latest `w64devkit-*.zip` (e.g. `w64devkit-1.20.0.zip`).

2. **Extract to Local Directory**  
   Extract ZIP file to custom directory (e.g. `C:\w64devkit`).

3. **Configure Environment Variables**  
   Add w64devkit's `bin` directory to system PATH:

- Right-click **This PC** → **Properties** → **Advanced system settings** → **Environment Variables**
- Under **System variables**, edit `Path` → Add path (e.g. `C:\w64devkit\bin`)

4. **Verify Installation**  
   Open Command Prompt and run:

```bash
bash make --version # Should output "GNU Make 4.x"
```

> If you encounter encoding issues, follow these steps:
 1. Modify the console encoding:
    ```bash
    reg add HKEY_CURRENT_USER\Console /v CodePage /t REG_DWORD /d 65001 /f
    ```
 2. Enable system-level UTF-8 support:
    - Settings → Time & language → Language → Manage language settings → Check "Beta: Unicode UTF-8" → Restart


## 1. protoc Installation

### MacOS

```bash
# Install
brew install protobuf
# Verify
protoc --version
```

### Windows

1. **Download protoc Compiler**  
   Visit the official releases
   page: [https://github.com/protocolbuffers/protobuf/releases](https://github.com/protocolbuffers/protobuf/releases)  
   Find and download the latest `protoc-x.x.x-win64.zip` (e.g. `protoc-25.1-win64.zip`).

2. **Extract and Configure Path**  
   Extract the ZIP file to a custom directory (e.g. `C:\protoc`).  
   Add the `protoc.exe` path to system environment variables:

- Right-click **This PC** → **Properties** → **Advanced system settings** → **Environment Variables**
- Under **System variables**, select `Path` → **Edit** → Add the extraction directory path (e.g. `C:\protoc\bin`)

3. **Verify Installation**  
   Open Command Prompt and run:

```bash
   protoc --version # Should output similar to "libprotoc 25.1"
```

Key improvements:

1. Added clear section headers with **bold** formatting
2. Used proper technical terms ("extract" instead of "unzip")
3. Standardized path formatting (`C:\protoc\bin`)
4. Streamlined environment variable configuration steps
5. Maintained consistent command/comment formatting

### Linux

**Ubuntu/Debian:**

```bash
# Install
sudo apt-get update
sudo apt-get install protobuf-compiler
# Verify
protoc --version
```

**Fedora/CentOS/RHEL:**

```bash
# Install
sudo yum install protobuf-compiler
# Verify
protoc --version
```

## 2. Project Environment Dependencies

### Go Version

* Go version >= 1.24.1

### Plugin Installation

```bash
make init
```

### Project Initialization

```bash
make all app=<palace|houyi|rabbit>
```

# Development

This project adopts a mini DDD design philosophy, divided into the following modules:

* API: Interface layer, which includes HTTP, gRPC, and other interface definitions
    * proto
    * pb
* service: Service entry point, which connects the API (Server) and biz layers, typically performing tasks like
  parameter validation, conversion, etc.
* biz: Business logic layer, which contains the implementation of business logic
    * bo
    * do
    * repository
    * vobj
* data: Data access layer, responsible for handling resources like databases, external data, and external services
    * cache
    * db
    * impl

## Dependency Inversion

The correct calling relationship is:

`service` -> `biz` -> `repository` -> `impl` -> `data`

It is important to note that each layer involves data structure transformation, meaning each layer only cares about the
data structures it receives. For example:

- The `service` layer only cares about the data structures in the `pb` package, which are the `req` and `resp` generated
  by `proto`. When the `service` layer calls the `biz` layer, it will convert the `pb` to `bo`.

## Transaction Management

> Transaction management is handled at the biz layer.

* Repository Definition

```go
type Transaction interface {
MainExec(ctx context.Context, fn func (ctx context.Context) error) error
BizExec(ctx context.Context, fn func (ctx context.Context) error) error
EventExec(ctx context.Context, fn func (ctx context.Context) error) error
}
```
