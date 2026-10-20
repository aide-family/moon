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
├── pkg/               # 全仓库共用的错误码、Redis 锁、配置装载
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
| 传输 | `internal/server` | 注入多个 service。HTTP、gRPC、中间件、健康检查 |
| 接口 | `internal/service` | 把请求转成 biz 入参，不写业务判断。一个 service 可以调用多个 biz |
| 用例 | `internal/biz/<领域>` | 权限、状态机、锁的持有时机。一个 biz 可以调用多个 repository。关键判断写明为什么拒绝、状态怎么变、和别的领域怎么配合 |
| 业务对象 | `pkg/bo` | 本应用 biz 使用的结构体 |
| 权限检查 | `internal/access` | 当前用户在当前租户是否持有权限。不和领域包放在同一目录 |
| 仓储接口 | `internal/data/repository` | 只放接口，按聚合拆文件。不引用 ent，不写查询 |
| 存储 | `internal/data` | 实现 `repository`。ent、迁移、种子、外部数据源客户端、Webhook、SSH |
| 事务邮件 | `internal/mailer` | SMTP 纯文本邮件发送。登录验证码和租户邀请共用，不直接写业务判断 |

同一父目录下的包不能互相 import。父包可以汇总子包的 Wire `ProviderSet`。领域包需要的结构体和接口放到父目录之外。

调用方向是 `server` → `service` → `biz` → `data/repository`。`data` 实现仓储接口，依赖 `repository` 和本应用的 `pkg/bo`，不依赖具体用例包。用 Wire 组装。应用之间不 import 对方的 `internal`。

每个手写包提供 `doc.go`，用包注释说明这个包做什么。`ent` 生成的子包不加。

手写 Go 文件的 import 分成四组，组之间空一行：空白导入、标准库、第三方、本仓库（`github.com/aide-family/Alioth`）。`ent`、`*.pb.go`、`wire_gen.go` 由生成器维护，不手改。

proto 的 service、rpc、message 和字段都写中文注释。

跨应用共用的函数、方法和常量放仓库根 `pkg`。只在一个应用里多处使用的，放 `app/<应用>/pkg`。

项目自定义错误放在仓库根 `pkg/merr`。只属于一个应用的错误放在该应用的 `pkg`。手写代码返回的错误必须是 Kratos error，不用 `errors.New` 或 `fmt.Errorf`。`ent` 生成代码除外。

业务表只出现在 `app/alioth` 的 ent schema。迁移只由 `alioth` 执行。`megrez`、`merak`、`alkaid` 不连接业务库，也不执行迁移。

Go 1.25+，Kratos v3，Cobra，Wire，protobuf，`buf.validate`，`google.api.http`。ORM 使用 ent，数据库使用 PostgreSQL。

日志使用标准库 `log/slog`（经 Kratos v3 `log.NewLogger` / `log.NewHandler` 构造），禁止再使用 Kratos v2 的 `log.Logger` / `log.Helper`。手写代码与生成代码的 import 与生成器均使用 `github.com/go-kratos/kratos/v3`（含 `protoc-gen-go-http/v3`）。HTTP JSON 在迁移期 blank-import `github.com/go-kratos/kratos/contrib/encoding/json/v3`，以保持与 Web 现有契约；不要同时注册 core `json` 与该兼容 codec。SSE、`config.Get[T]`、自定义 ValidatorFunc 等 v3 可选能力按需采用，不是默认交付。

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

## 进程退出

每个 Kratos 应用的 `kratos.New` **必须**配置 `kratos.StopTimeout`（建议 10 秒）。未配置时 stop 上下文永不超时，gRPC `GracefulStop` 与 HTTP `Shutdown` 会一直等待长连接（Agent 流、WebSocket shell 等），开发时 Ctrl+C 无法完整退出。

带长连接的服务须在 `kratos.BeforeStop`（或等价钩子）里主动关闭会话，不要只依赖强制停。实现 `transport.Server` 的 `Stop` 必须尊重传入的 stop `ctx`：等待后台 goroutine 结束时用 `select` 兼顾 `done` 与 `ctx.Done()`，禁止无限阻塞。

## 接口

对外 HTTP 与对外 gRPC 只由 `alioth` 提供。前缀 `/v1`，健康检查 `/health`。字段与路径在 `api/alioth/v1` 定稿。本文引用的 Moon 路径是能力来源，不是 Alioth 必须逐字复制的 URL。

内部 gRPC 不经过 Web nginx：

- `api/megrez/v1`：领取到期规则快照，回写评测结果
- `api/merak/v1`：领取待发送任务，回写投递结果
- `api/alioth/v1` 中的 Agent 双向流：注册、心跳、SSH、SFTP、Kubernetes 代理

错误分为四类：客户端错误、未认证、无权限、服务端错误。列表用 `page` 与 `pageSize`。下拉选择用 `limit` 与 `lastUid`。写接口用 `buf.validate`。

同一 Namespace 内，带名称的资源在未删除记录上满足 `(namespace_uid, name)` 唯一。平台级资源是 `User`、`UserOAuth`；`Permission` 是全局维护的租户权限定义，系统权限不建表。其余业务表带 `namespace_uid`。

## 身份

三种凭证互不替代：

| 凭证 | 持有者 | 可以做 | 不可以做 |
| --- | --- | --- | --- |
| 用户 JWT | 浏览器 | 调用管理 HTTP | 领取评测或通知任务，打开 Agent 流 |
| service key，`Authorization: Bearer sk-...` | `megrez`、`merak` | 调用对应的内部 gRPC | 当作普通用户访问管理页面 |
| Agent 令牌 | `alkaid` | 注册、心跳、接收执行帧 | 调用 Namespace、用户、规则等管理接口 |

业务请求带 `X-Namespace`。`alioth` 把它写入上下文，查询和写入都加上这个租户条件。缺少租户头的业务接口返回客户端错误。租户被停用后，除查看、启停和删除租户自身外，业务接口一律拒绝。平台用户管理和邀请消息不需要 `X-Namespace`。

`User` 是平台实体，不属于任何租户；`Member` 才是用户在租户中的身份。用户列表、详情、封禁、解封和平台角色授权只按系统权限校验，不按当前租户过滤。租户管理员不能枚举或封禁平台用户。

系统权限和租户权限分开：

| 类型 | 目录 | 授权载体 | 校验 |
| --- | --- | --- | --- |
| 系统权限 | `system:user:read`、`system:user:write`、`system:role:read`、`system:role:write`、`system:invite:read` | `User.system_role` | `access.RequireSystem`，不读 `X-Namespace` |
| 租户权限 | `member`、`role`、`namespace`、`datasource`、`alert_rule`、`alert`、`notification`、`agent`、`ssh`、`sftp`、`k8s` 的 `resource:action` | `Role` + `MemberRole` | `access.Allow` / `Require`，必须带 `X-Namespace` |

平台角色固定三档：第一个注册用户是 `SUPER_ADMIN`；`SYSTEM_ADMIN` 由超级管理员授权，可查看和管理普通用户；`USER` 没有系统权限。超级管理员唯一性由数据库部分唯一索引保证。已有部署升级时，最早注册的用户补为超级管理员。`system:user:read` 及以上才能查看用户列表。

租户预置角色：

| 角色 | 边界 |
| --- | --- |
| `owner` | 该 Namespace 的全部权限，含删除 Namespace、移交 owner |
| `admin` | 管理资源。不含删除 Namespace、移交 owner |
| `member` | 查看配置；处理告警；使用已授权的 SSH、SFTP、Kubernetes 代理 |
| `viewer` | 只读 |

权限名是 `resource:action`。租户资源为 `member`、`role`、`namespace`、`datasource`、`alert_rule`、`alert`、`notification`、`agent`、`ssh`、`sftp`、`k8s`；`user` 不属于租户权限。动作至少有 `read` 与 `write`。告警另有 `intervene`。SFTP 分开 `sftp:read` 与 `sftp:write`。执行动作为 `ssh:exec` 与 `k8s:proxy`。系统权限不写入 `permissions` 表，租户角色授权页也不返回它们。

租户邀请不直接创建成员：

1. `POST /v1/member/invite` 生成 `MemberInvite`，同一租户同一邮箱只保留一条待确认邀请。数据库只保存令牌摘要。
2. 被邀请人在平台邀请列表确认；未注册邮箱可先注册。确认时校验邮箱或平台用户、有效期和角色，之后才创建 `Member` 并写入 `MemberRole`。
3. 邀请邮件链接为 `{web_url}/invite/{token}`。`web_url` 来自 `auth.web_url` / `ALIOTH_WEB_URL`。邮件失败不回滚邀请。
4. 邀请默认 7 天过期；拒绝为 `DECLINED`，撤销为 `REVOKED`，过期由服务端惰性标记。管理员不能直接把成员状态改成 `JOINED`。
5. 邀请创建、同意和拒绝都写 `SystemMessage`。被邀请人收到 `INVITE_RECEIVED`，邀请人收到 `INVITE_ACCEPTED` / `INVITE_DECLINED`；同意或拒绝后把对应的收到消息标为已读。成员被移出租户时给被移除用户写 `MEMBER_REMOVED`。消息按 `user_uid` 或未注册邮箱匹配，接口在 `/v1/self/messages`。

鉴权放在 `alioth` 的中间件和 biz。页面隐藏按钮不能代替服务端拒绝。

## Redis 锁

Redis 只保存锁。业务状态、告警去重、发送任务都在 PostgreSQL。`alkaid` 和浏览器不连接 Redis。

| 键 | 持有者 | TTL | 续期 |
| --- | --- | --- | --- |
| `alioth:lock:schedule:{name}` | `alioth` | 15 秒 | 不续期 |
| `alioth:lock:eval:{rule_uid}:{datasource_uid}` | `megrez` | 30 秒 | 每 10 秒 |
| `alioth:lock:notify:{task_uid}` | `merak` | 60 秒 | 每 20 秒 |

协议：

1. 加锁使用 `SET key token NX PX ttl`。`token` 为 `{instance_id}:{random}`。
2. 续期与释放用 Lua。value 与 token 一致才 `PEXPIRE` 或 `DEL`。
3. 续期失败时放弃本次执行，不把任务报成成功。
4. Redis 不可用时拒绝评测和投递。
5. 锁客户端只编译进 `alioth`、`megrez`、`merak`。

PostgreSQL 用唯一约束挡住锁失效窗口内的重复写入，例如评测批次 `(rule_uid, datasource_uid, eval_slot)`、当前 FIRING 的 `(namespace_uid, fingerprint)`、同一事件的一次投递。这是数据约束，不是互斥手段。

## 敏感数据

用户密码只存哈希。数据源凭据、SMTP 密码、Webhook 签名、SSH 凭据加密后入库。接口响应不回明文密钥，更新时用空值表示保持原值。

SSH、SFTP、Kubernetes 代理写入审计。`member` 和 `viewer` 不能删除审计。

## Web

一个 Vite 应用，React、TypeScript、Tailwind CSS、shadcn/ui。不拆微前端。

- 页面放在 `web/src/pages`，请求函数放在 `web/src/api`。组件不直接使用 axios。
- 请求带 `Authorization: Bearer <token>` 和 `X-Namespace`。401 时清除令牌并回到登录页。
- 菜单分别按系统权限和当前用户在当前 Namespace 的租户权限生成。用户列表入口使用 `system:user:read`；邀请函和个人资料一样放在账户菜单里，不占侧栏菜单。
- 壳提供 Namespace 切换、主题（light / dark / system）和中英文。主题只通过显式菜单切换，不注册 `d` 等单键全局快捷键。
- 表单统一使用 `FormField`。必填项的 label 必须以红色 `*` 开头；提交时逐项校验，校验失败的输入框显示红色边框，并在对应输入框下显示该字段的错误原因。不要依赖浏览器原生气泡，提交表单使用 `noValidate`。
- 「级别条件」等复合字段必须提供说明 label（可见说明或 `hint`），内部每个输入项也要有 label。
- 列表页保留筛选、分页，并把查询条件放进 URL。新建和编辑用对话框。主从页面用左侧列表加右侧明细。
- 页面验收在浏览器里走 `http://<web>:8080`，包含登录、提交和跳转。

## 本期不做

探测任务、微前端、每个服务一块业务库、把 Redis 当缓存或消息队列、短信的发送实现、日志与链路的规则评测。这些不进入任何里程碑的交付。
