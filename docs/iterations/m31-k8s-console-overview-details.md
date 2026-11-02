# M31 Kubernetes 控制台概况聚合与资源详情

通用规范见 [conventions.md](conventions.md)。本文在 M14 的集群接入、M16 的工作负载运维与 M25 的操作级授权之上，把集群详情页的「概况」和「详情抽屉」从浏览器侧聚合改成服务端能力：

1. 概况改由一次聚合接口返回，替代前端 5–7 次列表请求，并去掉 Pod 采样上限。
2. 集群绑定指标数据源，概况补齐节点 CPU/内存与 Pod 用量；未绑定或查询失败返回明确状态。
3. 节点、Pod、工作负载详情补齐 conditions、taints、容量/可分配、卷、归属、最近事件与 ReplicaSet 历史。

本步只读，不新增 Kubernetes 写操作，不改 kubeconfig、Pod 日志/终端和 Job/CronJob 的既有语义。

## 目标与依赖

操作者可以：

1. 打开集群概况时只发一次请求，看到节点总数/Ready、Pod 按 phase 分布、三种工作负载的就绪/期望、Job 活跃/失败、Service/Ingress 数量、API Server 版本和可分配 CPU/内存。
2. 给集群绑定一个指标类数据源后，在概况看到节点 CPU/内存使用率与 Pod 用量汇总；未绑定显示「未绑定数据源」，查询失败显示失败原因，不弹服务端错误。
3. 在节点详情看到全部 conditions、taints、容量与可分配量、Pod 数和最近事件。
4. 在 Pod 详情看到容器及状态、conditions、volumes、ownerReferences 和最近事件。
5. 在 Deployment / StatefulSet / DaemonSet 详情看到 conditions、副本状态，以及 Deployment 的 ReplicaSet 历史（只读，供回滚参考）。

依赖：M14 / M16 / M25（集群接入、资源列表与写操作、操作级权限）、M2（数据源与指标查询）、M19 / M21 / M22（能力、套餐快照与菜单权限）。

口径：节点与容量始终是全集群；Pod、工作负载、Job、Service、Ingress 受请求里的 `namespace` 影响，空表示全部命名空间。权限仍按资源校验，不新增能力码。

## 交付

| 层 | 文件 | 变更 |
| --- | --- | --- |
| proto | `api/alioth/v1/cluster.proto` | `ClusterItem` 增加 `metrics_datasource_uid` / `metrics_datasource_name`；创建/更新接受绑定；新增 `GetClusterMetricsSummary` |
| proto | `api/alioth/v1/kubernetes.proto` | 新增概况、节点详情、Pod 详情 RPC；扩展工作负载详情与 ReplicaSet 消息 |
| 业务对象 | `app/alioth/pkg/bo/cluster.go` | 绑定字段、概况/详情/指标汇总对象 |
| 仓储接口 | `app/alioth/internal/data/repository/` | 集群绑定字段、节点/Pod 详情、事件查询 |
| 用例 | `app/alioth/internal/biz/cluster/inspector.go`（新增） | 只读巡检：概况聚合、节点/Pod 详情、指标汇总 |
| 用例 | `app/alioth/internal/biz/cluster/registry.go` | 绑定校验（同租户、指标类、未删除） |
| 存储 | `app/alioth/internal/data/ent/schema/cluster.go`、`cluster_repo.go` | `metrics_datasource_uid` 列与读写 |
| 存储 | `app/alioth/internal/data/kubernetes_detail.go`、`kubernetes_events.go`（新增） | 详情与事件查询 |
| Web | `web/src/pages/cluster/detail/`、`web/src/api/cluster/` | 概况单请求、详情抽屉区块、指标卡与绑定对话框 |

### 概况聚合接口

`GET /v1/cluster/{cluster_uid}/overview?namespace=`，要求 `cluster:read`。服务端并行调用节点、工作负载、Pod、Job、Service、Ingress 六个列表，在 `alioth` 内汇总后一次返回，不落库、不加缓存。

| 字段 | 类型 | 语义 |
| --- | --- | --- |
| `api_server_version` | string | 实时 discovery 的 `gitVersion`；失败回落 `clusters.version`；都取不到为空 |
| `node_total` / `node_ready` / `node_not_ready` | int32 | 全集群节点数；Ready 取 `Ready` condition |
| `allocatable_cpu_millis` / `allocatable_memory_bytes` | int64 | 只累加 Ready 且 `schedulable=true` 节点的 allocatable；解析失败按 0 计 |
| `pod_total` / `pod_phases[]` | int32 / 列表 | phase 固定按 Running、Pending、Succeeded、Failed、Unknown 五行返回，含 0 值；不再采样 |
| `workloads[]` | 列表 | 固定按 Deployment、StatefulSet、DaemonSet 三行返回 `kind`、`total`、`desired`、`ready`；DaemonSet 的 desired 取 `desiredNumberScheduled` |
| `job_total` / `job_active` / `job_failed` | int32 | 汇总 `status.active` 与 `status.failed` |
| `service_total` / `ingress_total` | int32 | 匹配当前 `namespace` 的数量 |
| `denied_sections[]` / `failed_sections[]` | 列表 | 取值 `nodes`、`workloads`、`pods`、`jobs`、`services`、`ingresses` |
| `generated_at` | int64 | 服务端完成时间（Unix 秒） |

- 权限：概况本身要 `cluster:read`；每个区块只在该资源读权限通过时统计，否则该区块字段置 0 并列入 `denied_sections`，接口仍返回 200。查询失败同理列入 `failed_sections`；单区块失败不拖垮整页。
- 若六个区块全部因同一个连接错误失败，返回 502/503（沿用 `kubeError`），不伪装成全空集群。
- 概况统计全量对象，不使用 M16 的「最多 300 个 Pod」采样；大集群的耗时与 API Server 压力见「风险与边界」。
- Web 的「整体失败」提示改为 `denied_sections` / `failed_sections` 驱动，卡片互不阻塞。

### 指标数据源绑定与指标汇总

- `clusters.metrics_datasource_uid` 默认空，可空；只接受同一租户、未删除、`type=METRICS` 的数据源，驱动必须是 `METRICS_PROMETHEUS` 或 `METRICS_VICTORIA_METRICS`。
- 绑定时要 `cluster:write` + `datasource:read`；读取集群详情只需 `cluster:read`。`ClusterItem` 同时返回 `metrics_datasource_name`，数据源已删除时名称为空。
- 创建/更新请求里的 `metrics_datasource_uid` 用 proto3 `optional string`：字段缺失表示保持原值，显式空串表示解绑。旧客户端更新集群不会误清绑定。
- 数据源状态不在绑定处拦截；查询时停用或删除按失败状态返回。

`GET /v1/cluster/{cluster_uid}/metrics/summary`，要求 `cluster:read`，固定执行四条 PromQL（不做用户输入表达式，5 分钟区间、整体 5 秒超时）：

| key | PromQL | 单位 |
| --- | --- | --- |
| `node_cpu_usage_percent` | `100 - avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100` | % |
| `node_memory_usage_percent` | `100 * (1 - sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes))` | % |
| `pod_cpu_usage_cores` | `sum(rate(container_cpu_usage_seconds_total{image!=""}[5m]))` | cores |
| `pod_memory_usage_bytes` | `sum(container_memory_working_set_bytes{image!=""})` | bytes |

响应统一 200，`status` 说明结果，避免前端把「没绑数据源」当成接口错误：

| status | 含义 | `metrics` |
| --- | --- | --- |
| `UNBOUND` | 集群未绑定数据源 | 不返回 |
| `OK` | 四条查询都取到值 | 四个字段都有值 |
| `PARTIAL` | 部分查询为空、失败或超时 | 只返回取到的字段，`unavailable_metrics[]` 列出缺失 key |
| `FAILED` | 数据源停用/已删除、地址不可达、鉴权失败或全部查询失败 | 不返回，`message` 说明原因 |

`message` 在非 `OK` 时非空；成功字段使用 proto3 `optional`，未取到时不设值，不能用 0 冒充。绑定与汇总都不新增权限码或能力码。

### 节点详情

`GET /v1/cluster/{cluster_uid}/nodes/{name}`，要求 `k8s:node:read`。

| 区块 | 字段 |
| --- | --- |
| 基础信息 | 复用 `KubernetesNodeItem`：状态、可调度、角色、内网 IP、kubelet 版本、OS 镜像、Pod 数、创建时间 |
| 条件 | `conditions[]`：type、status、reason、message、lastTransitionAt |
| 污点 | `taints[]`：key、value、effect |
| 容量/可分配 | `capacity` / `allocatable`：字符串 map，键为 `cpu`、`memory`、`pods`、`ephemeral-storage` 及扩展资源（如 `nvidia.com/gpu`） |
| 最近事件 | `events[]`：type、reason、message、count、source、firstSeenAt、lastSeenAt，按最后发生时间倒序，最多 20 条 |
| 降级 | `warnings[]`：事件读取缺少 RBAC 等非致命问题；接口仍成功 |

事件按 `involvedObject`（优先 UID）过滤；节点事件在 `default` 命名空间读取，失败只写 `warnings`，不影响节点详情主体。

### Pod 详情

`GET /v1/cluster/{cluster_uid}/pods/{namespace}/{name}`，要求 `k8s:pod:read`。

| 区块 | 字段 |
| --- | --- |
| 基础信息 | 阶段、就绪、重启、节点、Pod IP、QoS、ServiceAccount、开始时间、状态原因/消息 |
| 容器 | `containers[]` 与 `initContainers[]`：名称、镜像、就绪、重启次数、状态（RUNNING / WAITING / TERMINATED）、状态原因/消息、启动时间、端口 |
| 条件 | `conditions[]`：type、status、reason、message、lastTransitionAt |
| 存储卷 | `volumes[]`：名称、类型、只读、来源摘要；PVC 卷返回 claimName，Secret/ConfigMap 卷只返回名称，不读取任何值 |
| 归属 | `ownerReferences[]`：kind、name、uid、controller |
| 最近事件 | 同节点详情，命名空间中按 `involvedObject` 过滤，最多 20 条 |
| 降级 | `warnings[]` |

Pod 日志与终端继续用既有接口和权限，本步不改。

### 工作负载详情

复用 `GET /v1/cluster/{cluster_uid}/workload-detail?kind=&namespace=&name=`，要求 `k8s:workload:read`，只扩展响应：

- `conditions[]`：type、status、reason、message、lastTransitionAt，取工作负载 `status.conditions`。
- `replica_status`：desired、current、ready、updated、available、unavailable；DaemonSet 取 desired/current/ready/updated/available。
- `replicasets[]`（仅 Deployment）在现有字段上增加 `revision` 与 `current`；按 revision 倒序，最多 10 条。`revision` 取 ReplicaSet 的 `deployment.kubernetes.io/revision` 注解，`current` 等于 Deployment 当前 revision。
- 回滚仍走 M16 的 `GET /workload/revisions` 与 `POST /workload/rollback`，详情只展示历史，不新增写接口。

### Web

- 概况：`use-cluster-overview.ts` 改为一次 `getClusterOverview`，删除 `POD_SAMPLE_*` 常量与四路 `settle`；`overview-cards.tsx` 按新字段映射，采样提示改为全量口径，`denied_sections` / `failed_sections` 分别显示「无权限」和「重试」。
- 指标卡：`UNBOUND` 显示「未绑定数据源」空态与绑定按钮（`cluster:write` + `datasource:read` 可见）；绑定对话框用 `listDatasources({type:"METRICS"})` 过滤两种指标驱动。`OK` 显示 CPU/内存百分比与 Pod 用量，`PARTIAL`/`FAILED` 显示 `message` 与重试，不整页报错。
- 详情抽屉：节点、Pod、工作负载的「详细信息」增加条件/污点/可分配/容器状态/卷/归属/事件/副本集历史分组；关联数据打开时懒加载，失败可重试；YAML Tab 和字段复制行为不变。抽屉保持 `DetailSheet` 头部固定、单一滚动区。
- 工作负载详情的 ReplicaSet 历史提供「回滚」入口，仅 `k8s:workload:rollback` 可见，复用现有对话框。
- 权限、菜单、能力码都不新增；缺某个 `k8s:*:read` 时对应入口按现状隐藏，直接访问回落到概况。
- 中英文补齐 `kube.*` 新文案（conditions、taints、allocatable、volumes、ownerReferences、events、metrics 状态等）。

## 验收

- [ ] 概况接口一次返回六个区块；用真实集群对比节点/工作负载/Pod/Job/Service/Ingress 数量与现有列表 `total` 一致。
- [ ] 概况接口在 `namespace=default` 与空值下的 Pod/工作负载/Service/Ingress 数量不同，节点数不变。
- [ ] 用户只有 `cluster:read` 时概况返回 200，其余六个区块在 `denied_sections` 中，字段为 0；补上对应读权限后区块恢复。
- [ ] 单个资源列表失败只进入 `failed_sections`，其余区块正常；断开集群 API Server 时返回 502/503。
- [ ] 未绑定指标数据源时 `metrics/summary` 返回 `UNBOUND`，Web 显示「未绑定数据源」。
- [ ] 绑定非指标数据源、其他租户数据源或不存在的数据源被拒绝；显式空串解绑成功；旧客户端不传字段时绑定保留。
- [ ] 绑定可用 Prometheus/VictoriaMetrics 后返回 `OK` 与四个数值；停用数据源后返回 `FAILED`；只缺一条指标时返回 `PARTIAL` 并列出 `unavailable_metrics`。
- [ ] 节点详情返回全部 conditions、taints、capacity/allocatable（含扩展资源）和最多 20 条事件；无 events 权限时 `warnings` 非空且接口成功。
- [ ] Pod 详情返回普通/初始化容器状态、conditions、volumes、ownerReferences 和事件；Secret 卷只返回名称。
- [ ] Deployment 详情返回 conditions、副本状态与最多 10 条 ReplicaSet 历史，`current`/`revision` 正确；StatefulSet/DaemonSet 不返回 ReplicaSet。
- [ ] 节点/Pod/工作负载详情缺少对应 `k8s:*:read` 时返回 403；集群不属于当前租户返回 404。
- [ ] `make vet`、`make lint`、`make test` 通过（fake clientset 覆盖概况聚合、详情转换、事件过滤与权限门禁）。
- [ ] Web `pnpm typecheck`、`pnpm lint`、`pnpm build` 通过；概况网络面板只有一次 `/overview` 请求。
- [ ] 浏览器回归：概况六区块与指标卡状态、绑定/解绑、三类详情抽屉、窄屏 390×844 与 1366×768 滚动、中英文切换。

## 不做什么

- 不做 HPA、StorageClass、NetworkPolicy 等二级资源；不做控制面组件健康探测（scheduler/controller-manager/etcd 仍按现状标注未接入）。后续单独立项：HPA、StorageClass、NetworkPolicy、控制面真实健康探测。
- 不做指标图表、趋势、Top 消耗榜单；本期只返回四个汇总数值。
- 不做节点 Pod 列表过滤、事件分页/检索、事件保留策略。
- 不做概况服务端缓存、后台预聚合或跨集群汇总。
- 不改 kubeconfig 接入、Pod 日志/终端、Job/CronJob 写操作、YAML 预览语义。
- 不新增菜单、能力码或 `k8s:*` 权限；指标绑定沿用 `cluster:write` + `datasource:read`。

## 风险与边界

- **大集群压力**：概况一次触发六个全量 List，集群很大时 API Server 与 `alioth` 内存压力高于单页列表；先按现有 15 秒调用超时与并行执行落地，后续再考虑 fieldSelector、分页或短 TTL 缓存。
- **部分失败语义**：`denied_sections` / `failed_sections` 让接口始终 200，前端必须按区块降级；漏判会把「无数据」显示成 0，验收需同时覆盖权限缺失和查询失败。
- **指标命名依赖**：四条 PromQL 依赖 node-exporter 与 cAdvisor 的标准指标；未部署或指标名不同会得到 `PARTIAL`/`FAILED`，不自动发现替代指标。
- **可分配口径**：概况只累加 Ready 且可调度节点，和现有「容量」卡片（全部节点 capacity）不同，页面需分别标注口径。
- **事件权限**：`events` 是独立资源，常见只读角色可能没有 list 权限；只能降级为 `warnings`，不能因此拒绝整个详情。
- **绑定生命周期**：数据源软删除后集群仍保存 UID；汇总返回 `FAILED`，需要用户重新绑定。绑定不校验数据源启用状态，避免临时停用导致无法保存集群配置。
- **版本来源**：API Server 版本优先实时 discovery，失败回落最近一次连通性探测的缓存值，可能与当前版本短暂不一致。

## 文档同步

- 实现后更新 [kubernetes-console-design.md](../kubernetes-console-design.md) 第 5、7 节与接口清单，删除「后续可增强」中已交付项。
- 在 [../design.md](../design.md) 的 Kubernetes 与 Web 章节补充概况聚合、指标绑定与详情字段。
- 在 [../roadmap.md](../roadmap.md) 增加 M31 条目；本步不改变能力目录数量，`conventions.md` 不需要改权限口径。
