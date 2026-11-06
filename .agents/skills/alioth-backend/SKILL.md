---
name: alioth-backend
description: >-
  Implements Alioth Go services with Kratos v3, ent, and Redis locks. Use when
  editing app/alioth, app/megrez, app/merak, app/alkaid, api protobuf, pkg, or
  when the user mentions ent, PostgreSQL schema, service key, JWT, or
  distributed locks.
---

# Alioth 后端

先读 [alioth-platform](../alioth-platform/SKILL.md) 和 [docs/iterations/conventions.md](../../../docs/iterations/conventions.md)。

## 框架

- 使用 **Kratos v3**（`github.com/go-kratos/kratos/v3`）。禁止引入 `/v2`。
- 日志为标准库 `*slog.Logger`，用 v3 `log.NewLogger` / `log.NewHandler` 构造；不要用 v2 的 `log.Logger` / `log.Helper`。
- HTTP 生成器用 `protoc-gen-go-http/v3`（`make tools` / `make api`）。不要手改 `*_http.pb.go`。
- HTTP JSON 迁移期 blank-import `github.com/go-kratos/kratos/contrib/encoding/json/v3`，保持与 Web 契约；不要同时注册 core `json` 与该兼容 codec。
- 鉴权继续用自研中间件 + `golang-jwt`，不要默认引入 contrib JWT。OTel、熔断、SSE、`config.Get[T]`、自定义 ValidatorFunc 等按需采用，不是默认交付。
- **进程退出（硬约束）**：每个应用的 `kratos.New` 必须带 `kratos.StopTimeout`（建议 10s）。有 Agent 流 / WebSocket 等长连接时用 `BeforeStop` 主动关会话。自定义 `Server.Stop` 必须尊重 stop ctx，禁止 `<-done` 无限等待。漏配会导致 Ctrl+C 挂死。

## 分层

每个应用使用 `cmd`、`internal/server`、`internal/service`、`internal/biz/<领域>`、`internal/access`、`internal/data/repository`、`internal/data`、`internal/conf`、`pkg/bo`。`server` 注入多个 service，service 调用多个 biz，biz 调用多个 repository。`data/repository` 只放仓储接口，`data` 按同一聚合实现，并只依赖 `repository` 和本应用的 `pkg/bo`。biz 用到的结构体放在 `app/<应用>/pkg/bo`。用 Wire 组装。应用之间不 import 对方的 `internal`。

同一父目录下的包不能互相 import。父包可以汇总子包的 ProviderSet。

## 命名

完整规则见 [docs/iterations/conventions.md](../../../docs/iterations/conventions.md#命名)，`make lint` 会跑 `scripts/check-naming.sh`。要点：

- 类型名说明业务对象或业务行为，不用 `Usecase`、`Biz`、`Manager`、`Helper`、`Util`、`Impl`、`Common`、`Base` 这类层次名。`Data` 只有在包名已经限定业务时保留（`alerttmpl.Data`），`data.Data` 不行。
- biz 类型用业务角色：`Authenticator`（认证与会话）、`Coordinator`（告警事件调度与处置）、`Registry`（Agent 注册与在线连接）、`Membership`（成员与邀请）、`Inbox`（系统消息）、`Gateway`（数据源配置与查询出口）。
- 构造函数写 `New<类型>`；参数用注入对象的业务名；接收者用类型名小写（`func (registry *Registry)`），不用 `this`、`self`、`uc`、`d`。
- service 类型与 proto service 同名，字段用 biz 业务名（`AlertEventService{coordinator *alertevent.Coordinator}`）；仓储接口 `<聚合>Repo`；data 层实现不导出，共享容器叫 `Dependencies`。
- 同一概念在各层用同一个词干，缩写只用 `HTTP`、`gRPC`、`SSH`、`ID`、`UID`、`URL`、`DSN`、`JWT`、`TLS`、`SQL`、`API`。

biz 的关键判断要写清为什么拒绝、状态怎么变、以及和别的领域怎么配合。

每个手写包有 `doc.go`。手写 Go 的 import 按空白导入、标准库、第三方、本仓库四组排列，组间空一行。生成的 `ent`、`*.pb.go`、`wire_gen.go` 不手改。proto 的 service、rpc、message 和字段都写中文注释。

跨应用的函数、方法和常量放仓库根 `pkg`。只在一个应用内复用的放 `app/<应用>/pkg`。自定义错误放在 `pkg/merr`，只属于一个应用的放在该应用的 `pkg`。手写代码返回 Kratos error，不用 `errors.New` 或 `fmt.Errorf`。

业务表和迁移只放在 `app/alioth` 的 ent schema。`megrez`、`merak`、`alkaid` 不连接业务库。

## 接口

对外 HTTP 与 gRPC 只由 `alioth` 提供，前缀 `/v1`。路径在 `api/alioth/v1` 定稿，不新造设计文档没有的资源。

内部调用不进 Web nginx：

- `api/megrez/v1`：`alioth` 实现，`megrez` 调用
- `api/merak/v1`：`alioth` 实现，`merak` 调用
- Agent 双向流在 `api/alioth/v1`，调用者是 `alkaid`

## 凭证

用户 JWT、`Bearer sk-...` service key、Agent 令牌三者不能互相代替。业务请求带 `X-Namespace`，查询必须带 `namespace_uid`。

## Redis 锁

只使用 `pkg` 里的锁客户端。

| 键 | 持有者 | TTL |
| --- | --- | --- |
| `alioth:lock:schedule:{name}` | `alioth` | 15 秒，不续期 |
| `alioth:lock:eval:{rule_uid}:{datasource_uid}` | `megrez` | 30 秒，每 10 秒续期 |
| `alioth:lock:notify:{task_uid}` | `merak` | 60 秒，每 20 秒续期 |

加锁用 `SET key token NX PX`。续期和释放用 Lua，value 必须等于 token。续期失败或 Redis 不可用时，不把本次执行报成成功。任务状态留在 PostgreSQL。
