---
name: kratos-development
description: >-
  Develops Go microservices with Kratos v2 following official design philosophy,
  DDD/Clean Architecture layout, Protobuf API, error/config/middleware patterns,
  and observability. Use when building or modifying Kratos services, wiring servers,
  writing proto APIs, middleware, errors, config, or when the user mentions Kratos,
  go-kratos, kratos-layout, or microservice framework conventions.
---

# Kratos 微服务开发

基于 [Kratos Design](https://go-kratos.dev/docs/intro/design/) 与 [Layout](https://go-kratos.dev/docs/intro/layout/) 官方文档，指导在 Moon monorepo 中用 Kratos v2 开发微服务。

## 何时使用

- 新建或改造 Kratos 微服务（HTTP/gRPC/Job）
- 编写 proto API、错误码、中间件、配置加载
- 接入注册发现、日志、监控、链路追踪
- 理解分层职责与依赖方向

**Moon 具体模块实现**（biz/data/service 等）请配合 [proto-backend-module](../proto-backend-module/SKILL.md)。

## 设计哲学（核心）

Kratos 是 Go 微服务**工具箱**，不是绑定特定基础设施的全家桶：

| 原则 | 含义 |
|------|------|
| 插件化 | 各功能模块定义标准接口，第三方库通过实现接口接入 |
| 可定制 | 不强制项目结构；layout 是推荐实践，可改 |
| 轻量 | 框架本身轻，按需选组件 |
| DDD + Clean Architecture | V2 重构目标：可读性、可测试性、可维护性 |

生态：`kratos`（核心）、`contrib`（配置/日志/注册等插件）、`aegis`（限流/熔断，可独立于 Kratos）、`layout`（项目模板）。

## 标准分层（Layout）

```
cmd/run/{http,grpc,job,all}/   # 入口 + wire 注入
internal/
  conf/      # conf.proto → 配置结构
  server/    # HTTP/gRPC 实例、中间件、Register*Service
  service/   # 实现 proto Server 接口，DTO↔BO，薄层
  biz/       # 业务逻辑；repository 接口定义于此（依赖倒置）
  data/      # repo 实现；DB/缓存/远程调用封装
api/ 或 proto/                 # .proto + 生成代码
config/                        # 运行时 YAML
```

**依赖方向**：`server → service → biz ← data`（biz 定义接口，data 实现）。

| 层 | 职责 | 禁止 |
|----|------|------|
| service | 对接 API、参数转换、调用 biz | 复杂业务逻辑 |
| biz | 组合业务、定义 repository 接口 | 直接操作 DB |
| data | repo 实现、PO↔DTO 转换 | 跨表组合业务判断 |

Moon 差异：`proto/` 在仓库根目录；各 app 独立 `go.mod`；共享能力在 `magicbox/`。

## API 定义（Protobuf）

- 默认 `protoc` 生成 gRPC；HTTP 需在 proto 中加 `google.api.http`：

```protobuf
import "google/api/annotations.proto";

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply) {
    option (google.api.http) = { get: "/helloworld/{name}" };
  }
}
```

- HTTP 默认 JSON 序列化；其他格式见官方 Serialization 文档。
- **不适合 proto 的场景**（文件上传、无法映射的 JSON）：用普通 `http.Handler` 挂到路由，或在 service 层用 struct 定义字段。
- Moon：`make api` / `make gen` 生成；`go_package` 指向 `github.com/aide-family/<app>/pkg/api/v1`。

## Metadata

服务间调用需传递、但不进 payload 的元信息，用 Kratos `metadata` 包读写。Moon 中常见：`contextx` 封装 namespace、user 等（见 `magicbox/contextx`）。

## 错误处理

Kratos 错误四要素：

| 字段 | 用途 |
|------|------|
| `code` | 类 HTTP Status，网关可据此做重试/限流/熔断 |
| `reason` | 服务内唯一可读错误码（如 `USER_NOT_FOUND`） |
| `message` | 用户可读提示 |
| `metadata` | 扩展信息 |

```go
// 手动创建
errors.New(500, "USER_NAME_EMPTY", "user name is empty")

// proto 定义 + make errors 生成
// enum ErrorReason { USER_NOT_FOUND = 0 [(errors.code) = 404]; }
api.ErrorUserNotFound("user %s not found", name)

// 断言
if errors.Is(err, errors.BadRequest("USER_NAME_EMPTY", "")) { ... }
if api.IsUserNotFound(err) { ... }
e := errors.FromError(err) // e.Reason, e.Code
```

HTTP 错误 JSON：`{"code":500,"reason":"USER_NOT_FOUND","message":"...","metadata":{...}}`

Moon：业务错误优先用 `magicbox/merr` 与 app 内 proto 生成的 errors，保持与现有模块一致。

## 配置

统一接口：实现 `Source` + `Watcher` 即可接入任意配置源。

内置/插件：file、apollo、etcd、kubernetes、nacos。

Moon 模式：
- `internal/conf/*.proto` 定义 `Bootstrap`
- 运行时 `config/server.yaml` **字段必须与 proto 同步**
- 修改 conf.proto 后同步 YAML 并重新生成 pb

## 注册与发现

实现 `Registrar` + `Discovery` 接入注册中心。插件：consul、etcd、nacos、kubernetes、zookeeper 等。

Moon：`data` 层提供 `Registry()`；`kratos.New` 时 `kratos.Registrar(registry)` 注册服务。

## 日志

- `Logger`：底层接口，仅 `Log` 方法，用于适配第三方库
- `Helper`：带级别与格式化，**业务层推荐用 Helper**

插件：std（内置）、zap、fluent。Moon 中 biz/service 注入 `*klog.Helper`，错误用 `helper.Errorw(...)`。

## 可观测性

| 能力 | 方式 |
|------|------|
| Metrics | 实现 Metrics 接口；插件 prometheus、datadog |
| Tracing | OpenTelemetry；server/client 初始化时加 tracing 中间件 |
| 负载均衡 | 客户端：Weighted RR（默认）、P2C、Random |

Moon HTTP 示例中间件链：`recovery → logging → tracing → metadata → auth → validate`（见 `internal/server/http.go`）。

## 中间件

通过实现 `middleware.Middleware` 统一横切逻辑。常用内置/官方：recovery、logging、tracing、metadata、ratelimit、circuitbreaker。

```go
// 服务端注册
http.NewServer(http.Middleware(recovery.Recovery(), logging.Server(logger), ...))

// 选择性应用（Moon 模式）
authMiddleware := selector.Server(jwtMiddlewares...).
    Match(middler.AllowListMatcher(skipList...)).Build()
```

限流/熔断算法在 [aegis](https://github.com/go-kratos/aegis)，可独立于 Kratos 使用。

## 应用生命周期

```go
app := kratos.New(
    kratos.Name("service.http"),
    kratos.Server(httpSrv, grpcSrv),
    kratos.Logger(logger),
    kratos.Registrar(registry), // 可选
)
app.Run()
```

Moon：`cmd/run/server.go` 的 `NewApp` 为每个 transport 创建独立 `kratos.App`；wire 在 `cmd/run/{http,grpc,all}/wire.go`。

## Wire 依赖注入

```go
//go:build wireinject
func WireApp(...) (*kratos.App, func(), error) {
    panic(wire.Build(
        server.ProviderSetServerHTTP,
        service.ProviderSetService,
        biz.ProviderSetBiz,
        impl.ProviderSetImpl,
        data.ProviderSetData,
        run.NewApp,
    ))
}
```

各层用 `wire.NewSet(...)` 导出 ProviderSet；新增组件时注册到对应 ProviderSet。

## 新增 HTTP/gRPC 服务 checklist

```
- [ ] proto 定义 service + google.api.http（HTTP）
- [ ] make api 生成代码
- [ ] internal/service 实现生成的 Server 接口
- [ ] internal/biz 业务逻辑 + repository 接口
- [ ] internal/data/impl repository 实现
- [ ] ProviderSet 注册（service/biz/impl）
- [ ] internal/server Register*Service 注册路由
- [ ] 中间件白名单（若需跳过 auth/namespace）
- [ ] config/server.yaml 同步（若涉及配置）
```

## 数据库 / 缓存 / 消息队列

Kratos **不绑定** ORM/驱动，自选集成。常见：gorm、ent、go-redis、sarama/kafka-go。

Moon：GORM + gorm.io/gen；新增 model 后 gen query，禁止手写 SQL 字符串条件（见 proto-backend-module）。

## 反模式

- ❌ 在 service 层写复杂业务或直连 DB
- ❌ 在 data 层组合多表业务判断
- ❌ 跳过 wire ProviderSet 手动 new 依赖
- ❌ proto 改完不跑 make api / conf.proto 改完不同步 YAML
- ❌ 中间件链顺序随意（recovery/logging 应靠前）

## 延伸阅读

- 官方设计细节：[reference.md](reference.md)
- Moon 模块实现：[proto-backend-module](../proto-backend-module/SKILL.md)
- 官方文档：https://go-kratos.dev/
