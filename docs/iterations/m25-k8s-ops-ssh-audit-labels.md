# M25 Kubernetes 操作级能力、Pod 终端、SSH 统一审计与机器标签

通用规范见 [conventions.md](conventions.md)。本文在 M16 的 Kubernetes 控制台与 M5 的 SSH 命令之上补四块能力：

1. 套餐里的 Kubernetes 能力从「资源级」细到「具体操作级」，滚动重启、设置镜像、回滚、扩缩容、封锁节点、删 Pod、删 Job、CronJob 暂停 / 触发 / 删除各自成为独立的套餐能力与接口权限。
2. 新增 Pod Exec：从 Pod 列表进入当前页面的子路由，在浏览器里打开容器 Web Shell。
3. SSH 命令页面增加「执行审计」统一视图，集中查看所有执行日志并查看单条详情。
4. 机器支持自定义 labels，执行 SSH 命令时可以按 labels 解析主机列表；解析结果回写到二次确认对话框，允许在确认前移除不需要执行的主机。

控制面仍然是 `alioth`，浏览器仍然只经过 Web。`alkaid` 不新增端口，也不接业务库。

## 目标与依赖

依赖 M14 / M16（集群接入与 Kubernetes 控制台）、M19 / M21 / M22（能力目录、套餐快照与菜单化授权）、M5（SSH 命令、Agent 双向流与 Web Shell）。只改控制面与 Web，不改评测、通知和投递语义。

做：

- 能力目录：`cluster` 模块下按「资源 + 操作」拆出 23 个能力（`cluster_access` 保留），每个能力只含一个接口权限。
- 权限目录：Kubernetes 资源权限的动作从 `read/write` 细到 `read/cordon/scale/restart/image/rollback/delete/logs/exec/suspend/trigger`；旧的 `k8s:*:write` 在升级时展开成对应操作权限，旧能力码按「旧码 = 该资源全部既有操作」展开。
- Pod Exec：一次性票据 + WebSocket，服务端用集群 kubeconfig 走 `pods/exec` 子资源，浏览器用 xterm 终端，URL 是集群详情下的子路由。
- SSH 执行审计：`ssh-command-audits` 列表补齐 kind/status/操作者/关键字/时间范围过滤，返回机器名与命令名；SSH 命令页增加 URL 可回放的「执行审计」标签页，列表 + 详情抽屉。
- 机器 labels：`agent.labels` 存 JSON 对象；新增按 labels 匹配机器接口；SSH 执行对话框先解析主机列表，再由用户二次确认（可逐台移除）后执行。

不做：

- 不做 Deployment / StatefulSet / DaemonSet 的创建、任意 YAML 编辑、Pod 端口转发；Pod Exec 只提供交互 Shell，不提供文件上传下载。
- 不做按操作细分的审计留痕（Kubernetes 操作仍只由集群审计负责），不落库记录谁点了哪个按钮。
- 不做基于标签的批量定时任务、命令模板与标签绑定、标签表达式（AND 精确匹配，不做 `in` / `notin`）。
- 不改 Agent 令牌、SSH 凭据模型与 Web Shell 会话语义。
- 不把机器标签同步到 Kubernetes 节点或 Prometheus，标签只属于租户内的机器台账。

## 能力与权限

### Kubernetes 能力拆分

`cluster` 模块的能力（顺序即 `perm.Capabilities()` 与套餐页顺序）：

| 能力 code | 名称 | 包含权限 |
| --- | --- | --- |
| `cluster_access` | 集群接入 | `cluster:read` / `cluster:write` / `cluster:delete` |
| `k8s_node_read` | 查看节点 | `k8s:node:read` |
| `k8s_node_cordon` | 节点封锁 | `k8s:node:cordon` |
| `k8s_workload_read` | 查看工作负载 | `k8s:workload:read` |
| `k8s_workload_scale` | 工作负载扩缩容 | `k8s:workload:scale` |
| `k8s_workload_restart` | 工作负载滚动重启 | `k8s:workload:restart` |
| `k8s_workload_image` | 设置镜像 | `k8s:workload:image` |
| `k8s_workload_rollback` | 版本回滚 | `k8s:workload:rollback` |
| `k8s_pod_read` | 查看 Pod | `k8s:pod:read` |
| `k8s_pod_logs` | Pod 日志 | `k8s:pod:logs` |
| `k8s_pod_delete` | 删除 Pod | `k8s:pod:delete` |
| `k8s_pod_exec` | Pod 终端 | `k8s:pod:exec` |
| `k8s_ingress_read` | 查看 Ingress | `k8s:ingress:read` |
| `k8s_service_read` | 查看 Service | `k8s:service:read` |
| `k8s_configmap_read` | 查看 ConfigMap | `k8s:configmap:read` |
| `k8s_secret_read` | 查看 Secret | `k8s:secret:read` |
| `k8s_pv_read` | 查看 PV | `k8s:pv:read` |
| `k8s_pvc_read` | 查看 PVC | `k8s:pvc:read` |
| `k8s_job_read` | 查看 Job | `k8s:job:read` |
| `k8s_job_delete` | 删除 Job | `k8s:job:delete` |
| `k8s_cronjob_read` | 查看 CronJob | `k8s:cronjob:read` |
| `k8s_cronjob_suspend` | 暂停 / 恢复 CronJob | `k8s:cronjob:suspend` |
| `k8s_cronjob_trigger` | 立即运行 CronJob | `k8s:cronjob:trigger` |
| `k8s_cronjob_delete` | 删除 CronJob | `k8s:cronjob:delete` |

菜单保持按资源（`k8s_node`、`k8s_workload`、`k8s_pod`、`k8s_job`、`k8s_cronjob`、只读资源各一个），一个菜单由多个能力开通，菜单权限不变（`menu:k8s_node` 等）。菜单里的默认读权限仍是该资源的读权限。

### 权限名

- 读：`k8s:node:read`、`k8s:workload:read`、`k8s:pod:read`、`k8s:pod:logs`、`k8s:ingress:read`、`k8s:service:read`、`k8s:configmap:read`、`k8s:secret:read`、`k8s:pv:read`、`k8s:pvc:read`、`k8s:job:read`、`k8s:cronjob:read`。
- 操作：`k8s:node:cordon`；`k8s:workload:{scale,restart,image,rollback}`；`k8s:pod:{delete,exec}`；`k8s:job:delete`；`k8s:cronjob:{suspend,trigger,delete}`。
- 只读判定仍按动作：`read` / `logs` 是只读，其余动作（含 `exec`）不是只读，平台管理员跨租户只读视角不能执行。
- `cluster:*` 只负责集群接入、连通性、凭据与删除，不再覆盖任何资源操作。

### 升级兼容

- 旧能力码按「该资源当前已实现的操作」展开成新能力：`k8s_node` → 读 + 封锁；`k8s_workload` → 读 + 扩缩容 + 重启 + 设置镜像 + 回滚；`k8s_pod` → 读 + 日志 + 删除；`k8s_job` → 读 + 删除；`k8s_cronjob` → 读 + 暂停 + 触发 + 删除；`k8s_ingress` / `k8s_service` / `k8s_configmap` / `k8s_secret` / `k8s_pv` / `k8s_pvc` → 各自的读能力。
- **Pod Exec 是新增权限，不随旧 `k8s_pod` 或旧模块级 `cluster` 编码自动授予**，避免存量租户在升级后直接拿到容器终端。需要超级管理员在套餐里显式开通 `k8s_pod_exec`。
- 旧角色权限 `k8s:node:write` / `k8s:workload:write` / `k8s:pod:write` / `k8s:job:write` / `k8s:cronjob:write` 分别展开成上表对应的操作权限；迁移完成后删除旧 `k8s:*:write` 权限行与角色关联，`k8s:pod:logs` 保留原名不动。
- 迁移只加不减、可重复执行；读取路径继续归一化旧能力码，保证降级或旧客户端提交也不会丢能力。

## 数据与业务对象

不需要新表；只改两处结构：

| 位置 | 变更 |
| --- | --- |
| `agents` | 新增 `labels` 文本列，默认 `{}`，保存 JSON 对象（字符串键值对）。跟 `probe_target_groups.labels` 一样用 JSON 文本，不建关联表。 |
| `bo.Agent` | 新增 `Labels string`；`bo.SSHCommandAudit` 新增 `AgentName`、`CommandName`（列表展示用，不入库）。 |

- 标签键值都做长度与字符校验：键 1–63、值 0–63，只允许字母、数字、`.`、`_`、`-`；最多 32 对；`{}` 表示无标签。
- 标签匹配是 AND 精确匹配：请求里的每对键值都必须同时出现在机器标签中；请求为空返回客户端错误。
- 匹配只看当前租户、`status=ACTIVE` 的机器，返回全部命中项（含 `online`），在线与否由页面提示，执行时仍逐台校验在线。
- SSH 审计列表的 `agent_name` / `command_name` 由数据层按 `agent_uid`、`command_uid` 批量 join 补全；机器或命令被删除后留空，不因为 join 失败丢审计。

## 接口

所有路径前缀 `/v1`。集群相关路径中的 `{cluster_uid}` 仍按当前租户校验，跨租户按不存在处理。

### Kubernetes Pod 终端

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/pod/{name}/exec-ticket` | `k8s:pod:exec` | 请求带 `container`、`cols`、`rows`；返回 60 秒内一次性票据 |
| `GET` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/pod/{name}/exec?ticket=…` | 票据 | WebSocket；二进制帧为终端字节，文本帧 `{"type":"resize","cols":80,"rows":24}`；服务端关闭前发 `{"type":"exit","reason":"…"}` |

- 票据绑定集群、命名空间、Pod、容器与窗口大小；同一票据只能消费一次，过期或目标不匹配返回 401。
- 服务端用已保存的 kubeconfig 建 `rest.Config`，通过 `pods/exec` 子资源建立 SPDY 流，TTY 打开 `/bin/sh`；容器不存在、容器无 shell、凭据无 exec 权限分别返回可读错误说明。
- 会话断开即结束远端进程；服务端进程退出时关闭全部活动终端（沿用 `kratos.StopTimeout` 与 `BeforeStop` 约束）。

### 机器标签与匹配

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/v1/agents/match` | `ssh:exec` | 请求体 `labels` 为字符串键值对象；返回命中的 ACTIVE 机器（含 `online`） |
| `PUT` | `/v1/agent/{uid}` | `agent:write` | 请求新增 `labels`；创建接口同样接受 |
| `GET` | `/v1/agent/{uid}`、`/v1/agents` | `agent:read` | 返回项新增 `labels`（JSON 文本） |

- `labels` 传 `{}` 表示清空；非法 JSON、非法键值、超过 32 对返回客户端错误。
- 匹配只返回当前租户的机器；不存在的标签组合返回空列表（`items=[]`），不是错误。
- 执行命令仍用 `POST /v1/ssh-command/{uid}/execute` 显式传 `agentUids`：标签只用于解析候选主机，最终执行以确认后的主机列表为准。

### SSH 执行审计

`GET /v1/ssh-command-audits` 扩展查询参数：

| 参数 | 说明 |
| --- | --- |
| `kind` | 留空为全部；`CREATE` / `UPDATE` / `APPROVE` / `REJECT` / `EXECUTE`。统一审计页传 `EXECUTE` |
| `status` | 留空为全部；`PENDING` / `APPROVED` / `REJECTED` / `DONE` / `FAILED` |
| `commandUid` / `agentUid` | 按命令、机器过滤（保留既有语义） |
| `operatorUid` | 按操作者过滤 |
| `keyword` | 命令正文子串（大小写不敏感） |
| `startTime` / `endTime` | Unix 秒；`startTime` 含、`endTime` 不含；0 表示不限 |

- 响应项新增 `agentName`、`commandName`；排序仍按创建时间倒序，分页语义不变。
- 统一审计页只读，不提供删除审计的接口。

## Web

### Kubernetes 控制台

- 能力与权限拆分后，集群详情仍然按资源视图过滤导航（读能力 + 菜单权限），写入口按操作权限分别隐藏：
  - 节点：`k8s:node:cordon` 才显示封锁 / 解除封锁。
  - 工作负载：扩缩容、滚动重启、设置镜像、回滚分别看 `k8s:workload:scale|restart|image|rollback`；详情与历史版本只要读权限。
  - Pod：删除看 `k8s:pod:delete`，日志看 `k8s:pod:logs`，终端看 `k8s:pod:exec`。
  - Job：删除看 `k8s:job:delete`；CronJob：暂停 / 恢复、立即运行、删除分别看 `k8s:cronjob:suspend|trigger|delete`。
- 终端 WebSocket 地址必须是绝对的 `ws://` / `wss://`（`{proto}//{host}/v1/cluster/.../exec?ticket=...`），不接受相对路径；握手 10 秒未完成时显示超时提示与「重新连接」，不允许停在无限「正在连接」。
- Pod 行操作新增「终端」（`k8s:pod:exec`）。点击后在当前页面内进入子路由
  `/clusters/{cluster_uid}/pods/{namespace}/{name}/exec?container={容器}`：页头显示集群 / 命名空间 / Pod / 容器，正文是占满可用高度的 xterm 终端，返回按钮回到 Pod 视图。
- 多容器 Pod 先在终端页顶部选择容器；切换容器重建会话。终端断开显示原因与「重新连接」。
- URL 支持直接打开（权限不足回落 `/forbidden`），跟随 `web/AGENTS.md` 的滚动与高度约束：终端容器自己占满 `flex-1 min-h-0`，不允许文档级滚动。
- 终端页用 `k8s:pod:read` 解析容器列表；只有 `k8s:pod:exec` 的账号拿不到容器列表，此时 URL 带 `?container=` 仍可打开终端（没有容器切换器），服务端仍只校验 `k8s:pod:exec`。

### SSH 命令与审计

- SSH 命令页顶部增加 URL 可回放的标签：`tab=commands`（默认）与 `tab=audits`（`?tab=audits`）。审计标签直接渲染统一审计列表，不重复命令列表的筛选条件。
- 统一审计列表列：时间、命令、机器、操作者、状态、退出码；筛选：关键字、状态、机器、操作者、时间范围（共享 `DateRangePicker`，最多 30 天）。行点击打开详情抽屉：命令正文、输出（等宽、可复制、可换行）、退出码、错误说明、机器、操作者、耗时信息由创建时间与状态表达。
- 执行命令对话框支持两种选机方式：
  - 按机器：保留现有在线机器多选。
  - 按标签：输入标签键值（复用机器的标签编辑器），点「匹配主机」调用 `/v1/agents/match`，把命中的主机列表回写到对话框内的二次确认区；列表默认勾选在线机器，离线机器不可选；用户可以逐台取消，也可以全选 / 全不选。确认执行只提交仍被勾选的机器。
  - 标签模式在每次执行前都可以重新匹配；执行失败保留对话框与已选主机。
- 机器页：创建 / 编辑对话框增加标签编辑器（键值行，支持增删）；列表展示标签 chips（最多 3 个，余下折叠为 `+N`），详情抽屉展示全部标签；`labels` 为空时不显示占位标签。
- 所有新增文案同时补 zh / en 语言包，沿用 `StatusBadge`、`RowActions`、`DetailSheet`、`Dialog` 与 32px 控件约束。

## 验收标准

- [ ] 套餐能力目录里 `cluster` 模块按操作拆成 24 个能力；角色授权页的 `k8s_*` 菜单下能看到对应子权限，勾选菜单仍默认带读权限。
- [ ] 只持有 `k8s:workload:read` 的账号看不到扩缩容、重启、设置镜像、回滚入口，直接调接口 403；分别授予 `k8s:workload:image` 后只能设置镜像，仍不能回滚。
- [ ] 旧的 `k8s:workload:write` 角色升级后自动获得四个工作负载操作权限；旧 `k8s_pod` 能力展开后不含 `k8s_pod_exec`。
- [ ] Pod 终端：`k8s:pod:exec` 账号从 Pod 行操作进入子路由，能看到容器内 shell 并执行 `id` / `ls`；无该权限的账号没有入口且拿不到票据；票据重复使用、过期或目标不匹配被拒绝。
- [ ] 终端窗口缩放后容器内 `stty size` 跟随；关闭页面后服务端远端进程被回收。
- [ ] 机器创建 / 编辑可以保存多组标签，列表与详情正确显示；非法标签返回客户端错误。
- [ ] 按标签匹配命中多台机器后，二次确认列表显示全部命中项；取消勾选的主机不执行，执行结果只包含确认列表中的机器。
- [ ] 统一审计页列出全部 `EXECUTE` 审计，按关键字、状态、机器、操作者、时间范围过滤结果正确；详情抽屉显示完整命令与输出；`ssh:read` 账号可看，无 `ssh:read` 的账号 403。
- [ ] 旧命令结果页与命令审计时间线行为不变。
- [ ] `go vet`、`make lint`、`go test ./...`、Web 的 `typecheck` / `lint` / `build` 全部通过。

## 验证方式

- 后端：`perm` 目录与兼容展开用表驱动单测；机器标签解析、匹配过滤与审计过滤用 fake 客户端 / 仓储单测；Pod 终端票据用单测覆盖一次性、过期与目标绑定；SPDY 执行器由 fake 或真实集群人工验证。
- Web：真实浏览器里走 Pod 终端、审计标签、标签执行二次确认三条路径；至少覆盖 1366×768 与 390×844 两个视口，确认终端与表格滚动都在有界容器内。
- 不具备真实 Kubernetes 集群或 Agent 时，相关浏览器路径如实标注「未验证」，不写「应该没问题」。
