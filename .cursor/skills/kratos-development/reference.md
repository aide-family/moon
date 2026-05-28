# Kratos 官方设计文档参考

来源：[Design](https://go-kratos.dev/docs/intro/design/) · [Layout](https://go-kratos.dev/docs/intro/layout/)

## 为何 V2 完全重构

V1 随业务迭代出现：变更成本高、缺乏合理抽象、模块难测、第三方库迁移困难。V2 参考 DDD 与 Clean Architecture 重设计项目结构，并将框架设计为「插座」——轻量、插件化、可定制，几乎每个微服务功能模块都有标准接口与第三方插件。

## 项目生态

| 项目 | 说明 |
|------|------|
| kratos | 核心：CLI、HTTP/gRPC 生成、生命周期；链路/配置/日志/注册/监控接口定义 |
| contrib | 配置、日志、注册、监控等可直接使用的组件 |
| aegis | 限流、熔断等可用性算法，少依赖、不绑定 Kratos |
| layout | 默认项目模板（DDD + Clean Architecture），非强制 |
| gateway | API 网关（建设中） |

## Layout 目录详解

```
.
├── api/              # .proto 及生成代码
├── cmd/server/       # main.go, wire.go, wire_gen.go
├── configs/          # 本地配置
└── internal/
    ├── conf/         # conf.proto → conf.pb.go
    ├── data/         # 数据源封装；实现 biz 层 repo 接口；PO→DTO
    ├── biz/          # 业务逻辑；repo 接口定义（依赖倒置）
    ├── service/      # 实现 API；DTO→DO；组合 biz（应用层）
    └── server/       # http/grpc 实例创建
```

各层 README（layout 仓库）进一步说明职责边界。

## CLI

`kratos new <project-name>` 从 layout 创建项目；CLI 还用于维护依赖版本等。

## API 完整示例

```protobuf
syntax = "proto3";

package helloworld.v1;

import "google/api/annotations.proto";

option go_package = "github.com/go-kratos/kratos-layout/api/helloworld/v1;v1";

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (google.api.http) = {
      get: "/helloworld/{name}"
    };
  }
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}
```

生成后：注册 server 端代码于项目内，或引用 client 端做远程调用。

## 错误 proto 定义示例

```protobuf
syntax = "proto3";

package api.blog.v1;

import "errors/errors.proto";

option go_package = "github.com/go-kratos/kratos/examples/blog/api/v1;v1";

enum ErrorReason {
  option (errors.default_code) = 500;

  USER_NOT_FOUND = 0 [(errors.code) = 404];
  CONTENT_MISSING = 1 [(errors.code) = 400];
}
```

layout 中常用 `make errors` 生成辅助方法。

## 错误创建与断言

```go
errors.New(500, "USER_NAME_EMPTY", "user name is empty")

api.ErrorUserNotFound("user %s not found", "kratos")

err := errors.New(500, "USER_NAME_EMPTY", "user name is empty")
err = err.WithMetadata(map[string]string{"foo": "bar"})

// 断言
if errors.Is(err, errors.BadRequest("USER_NAME_EMPTY", "")) { }
e := errors.FromError(err)
if e.Reason == "USER_NAME_EMPTY" && e.Code == 500 { }
if api.IsUserNotFound(err) { }
```

## 配置插件

| 插件 | 说明 |
|------|------|
| file | 内置 |
| apollo | 远程 |
| etcd | 远程 |
| kubernetes | 远程 |
| nacos | 远程 |

接口：`Source`（加载）、`Watcher`（订阅变更）。

## 注册发现插件

consul、discovery、etcd、kubernetes、nacos、zookeeper。

接口：`Registrar`（注册）、`Discovery`（发现）。

## 日志插件

| 层级 | 说明 |
|------|------|
| Logger | 底层，一个 `Log` 方法，适配各日志库 |
| Helper | 高级，带级别与格式化，业务推荐 |

插件：std、fluent、zap。

## 监控插件

datadog、prometheus（实现 Metrics 接口）。

## 链路追踪

OpenTelemetry 标准；client/server 初始化时配置 tracing middleware，对接 Jaeger 等。

## 负载均衡

Weighted round robin（默认）、P2C、Random；客户端初始化时配置。

## 限流与熔断

Ratelimit、Circuitbreaker 中间件；算法见 aegis 仓库，可独立使用。

## 第三方库推荐（官方列举）

**Database**: database/sql, gorm, ent

**Cache**: go-redis, redigo, gomemcache

**Message Queue**: sarama, kafka-go

更多见 awesome-go。

## 社区

- GitHub: https://github.com/go-kratos
- 文档: https://go-kratos.dev/
- Examples: https://github.com/go-kratos/kratos/tree/main/examples
