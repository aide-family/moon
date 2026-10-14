# 开发规范

本规范适用于 M0 到 M7。各里程碑只写本步新增的约束，通用部分以本文为准。设计见 [../design.md](../design.md)，里程碑顺序见 [../roadmap.md](../roadmap.md)。

## 仓库与分层

单仓。四个后端应用各自一个 `go.mod`，用 `replace` 指向仓库根的 `pkg/`。

```
Alioth/
├── api/
│   ├── alioth/v1/     # 控制面对外 HTTP/gRPC，以及 Agent 流
│   ├── megrez/v1/     # 评测领取与回写，alioth 实现，megrez 调用
│   └── merak/v1/      # 通知领取与回写，alioth 实现，merak 调用
├── pkg/               # 鉴权、错误码、Redis 锁、加密字段
├── app/
│   ├── alioth/
│   ├── megrez/
│   ├── merak/
│   └── alkaid/
└── web/
```

每个 Kratos 应用分层：

| 层 | 目录 | 职责 |
| --- | --- | --- |
| 入口 | `cmd` | Cobra 子命令。`alioth` 提供 HTTP 与 gRPC。`megrez`、`merak` 提供运行循环和 `/health` |
| 传输 | `internal/server` | HTTP、gRPC、中间件、健康检查 |
| 接口 | `internal/service` | 把请求转成 biz 入参，不写业务判断 |
| 用例 | `internal/biz` | 权限、状态机、锁的持有时机 |
| 存储 | `internal/data` | ent、外部数据源客户端、SMTP、Webhook、SSH |

依赖方向是 `server` → `service` → `biz` → `data`。`biz` 定义仓储接口，`data` 实现。用 Wire 组装。

业务表只出现在 `app/alioth` 的 ent schema。迁移只由 `alioth` 执行。`megrez`、`merak`、`alkaid` 不连接业务库，也不执行迁移。

Go 1.25+，Kratos v2，Cobra，Wire，protobuf，`buf.validate`，`google.api.http`。ORM 使用 ent，数据库使用 PostgreSQL。

## 端口与入口

| 进程 | 监听 | 访问范围 |
| --- | --- | --- |
| `web` | HTTP 8080 | 集群外唯一入口 |
| `alioth` | HTTP 8000，gRPC 9000 | 集群内。浏览器经 nginx 访问 8000 |
| `megrez` | HTTP 8001，仅 `/health` | 集群内探针 |
| `merak` | HTTP 8002，仅 `/health` | 集群内探针 |
| `alkaid` | HTTP 8022，gRPC 9022 | 所在机器或集群内。业务会话出站连接 `alioth:9000` |

`web` 的 nginx 只托管静态文件，并把 `/v1` 反代到 `alioth:8000`。开发时 Vite 做同样的代理。不要按路径把请求分到 `megrez`、`merak` 或 `alkaid`。

8022 与 9022 不对集群外发布。`alkaid` 不在公网提供管理入口。

配置使用 Bootstrap YAML。端口、数据库、Redis、密钥用环境变量占位。

## 接口

对外 HTTP 与对外 gRPC 只由 `alioth` 提供。前缀 `/v1`，健康检查 `/health`。字段与路径在 `api/alioth/v1` 定稿。本文引用的 Moon 路径是能力来源，不是 Alioth 必须逐字复制的 URL。

内部 gRPC 不经过 Web nginx：

- `api/megrez/v1`：领取到期规则快照，回写评测结果
- `api/merak/v1`：领取待发送任务，回写投递结果
- `api/alioth/v1` 中的 Agent 双向流：注册、心跳、SSH、SFTP、Kubernetes 代理

错误分为四类：客户端错误、未认证、无权限、服务端错误。列表用 `page` 与 `pageSize`。下拉选择用 `limit` 与 `lastUid`。写接口用 `buf.validate`。

同一 Namespace 内，带名称的资源在未删除记录上满足 `(namespace_uid, name)` 唯一。全局资源是 `User`、`UserOAuth`、`Permission`。其余业务表带 `namespace_uid`。

## 身份

三种凭证互不替代：

| 凭证 | 持有者 | 可以做 | 不可以做 |
| --- | --- | --- | --- |
| 用户 JWT | 浏览器 | 调用管理 HTTP | 领取评测或通知任务，打开 Agent 流 |
| service key，`Authorization: Bearer sk-...` | `megrez`、`merak` | 调用对应的内部 gRPC | 当作普通用户访问管理页面 |
| Agent 令牌 | `alkaid` | 注册、心跳、接收执行帧 | 调用 Namespace、用户、规则等管理接口 |

业务请求带 `X-Namespace`。`alioth` 把它写入上下文，查询和写入都加上这个租户条件。缺少租户头的业务接口返回客户端错误。

预置角色：

| 角色 | 边界 |
| --- | --- |
| `owner` | 该 Namespace 的全部权限，含删除 Namespace、移交 owner |
| `admin` | 管理资源。不含删除 Namespace、移交 owner |
| `member` | 查看配置；处理告警；使用已授权的 SSH、SFTP、Kubernetes 代理 |
| `viewer` | 只读 |

权限名是 `resource:action`。资源为 `user`、`member`、`role`、`namespace`、`datasource`、`alert_rule`、`alert`、`notification`、`agent`、`ssh`、`sftp`、`k8s`。动作至少有 `read` 与 `write`。告警另有 `intervene`。SFTP 分开 `sftp:read` 与 `sftp:write`。执行动作为 `ssh:exec` 与 `k8s:proxy`。具体枚举在 proto 中定稿后，服务端与 Web 菜单使用同一份。

鉴权放在 `alioth` 的中间件和 biz。页面隐藏按钮不能代替服务端拒绝。

## Redis 锁

Redis 只保存锁。业务状态、告警去重、发送任务都在 PostgreSQL。`alkaid` 和浏览器不连接 Redis。

| 键 | 持有者 | TTL | 续期 |
| --- | --- | --- | --- |
| `alioth:lock:schedule:{name}` | `alioth` | 15 秒 | 不续期 |
| `alioth:lock:eval:{rule_uid}` | `megrez` | 30 秒 | 每 10 秒 |
| `alioth:lock:notify:{task_uid}` | `merak` | 60 秒 | 每 20 秒 |

协议：

1. 加锁使用 `SET key token NX PX ttl`。`token` 为 `{instance_id}:{random}`。
2. 续期与释放用 Lua。value 与 token 一致才 `PEXPIRE` 或 `DEL`。
3. 续期失败时放弃本次执行，不把任务报成成功。
4. Redis 不可用时拒绝评测和投递。
5. 锁客户端只编译进 `alioth`、`megrez`、`merak`。

PostgreSQL 用唯一约束挡住锁失效窗口内的重复写入，例如评测批次 `(rule_uid, eval_slot)`、当前 FIRING 的 `(namespace_uid, fingerprint)`、同一事件的一次投递。这是数据约束，不是互斥手段。

## 敏感数据

用户密码只存哈希。数据源凭据、SMTP 密码、Webhook 签名、SSH 凭据加密后入库。接口响应不回明文密钥，更新时用空值表示保持原值。

SSH、SFTP、Kubernetes 代理写入审计。`member` 和 `viewer` 不能删除审计。

## Web

一个 Vite 应用，React、TypeScript、Tailwind CSS、shadcn/ui。不拆微前端。

- 页面放在 `web/src/pages`，请求函数放在 `web/src/api`。组件不直接使用 axios。
- 请求带 `Authorization: Bearer <token>` 和 `X-Namespace`。401 时清除令牌并回到登录页。
- 菜单按当前用户在当前 Namespace 的权限生成。
- 壳提供 Namespace 切换、主题（light / dark / system）和中英文。
- 列表页保留筛选、分页，并把查询条件放进 URL。新建和编辑用对话框。主从页面用左侧列表加右侧明细。
- 页面验收在浏览器里走 `http://<web>:8080`，包含登录、提交和跳转。

## 本期不做

探测任务、微前端、每个服务一块业务库、把 Redis 当缓存或消息队列、短信的发送实现、日志与链路的规则评测。这些不进入任何里程碑的交付。
