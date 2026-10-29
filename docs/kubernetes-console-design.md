# Kubernetes 集群详情：控制台化改造设计

本文对应 `web/src/pages/cluster/` 的详情页改造。页面现已覆盖节点、工作负载、Job、CronJob、Pod、Ingress、Service、ConfigMap、Secret、PV、PVC 十一类资源，全部使用 `alioth` 的 `/v1/cluster/...` 接口；Secret 与 ConfigMap 只展示键名，任何接口都不返回值。

## 1. 页面结构

```
集群详情
├─ 页头：集群名 / API Server / 测试连接 / 返回列表
├─ 元数据条：连接状态 · 版本 · cluster/context · 最近检测 · 创建时间 · 备注 · 最近错误
└─ 控制台主体（低屏宽时导航折叠为分组下拉）
   ├─ 左侧资源导航（248px，独立滚动）
   │  ├─ 集群概况
   │  ├─ 节点                        集群级
   │  ├─ 负载
   │  │  ├─ 全部工作负载
   │  │  ├─ Deployment / StatefulSet / DaemonSet
   │  │  ├─ Job / CronJob
   │  │  └─ Pod
   │  ├─ 服务与路由
   │  │  ├─ Ingress
   │  │  └─ Service
   │  ├─ 配置管理
   │  │  ├─ ConfigMap              只读、仅键名
   │  │  └─ Secret                 只读、仅键名，永不返回值
   │  └─ 存储
   │     ├─ PV                      集群级
   │     └─ PVC
   └─ 右侧资源面板
      ├─ 工具栏：资源名 / 作用域徽章 / 命名空间选择（命名空间级资源）/ 筛选开关 / 刷新
      ├─ 已应用筛选 chips（仅名称）
      ├─ 名称筛选面板（默认收起，可展开）
      └─ 内容：概况面板 / 表格列表 / 详情抽屉
```

默认视图仍是「节点」，与改造前的链接、书签和已有浏览器行为一致；「集群概况」是导航第一项。

### URL 契约

| 参数 | 含义 | 说明 |
| --- | --- | --- |
| `tab` | 资源视图 | `overview`、`workloads`、`jobs`、`cronjobs`、`pods`、`ingresses`、`services`、`configmaps`、`secrets`、`pvs`、`pvcs`；缺省为节点（节点不写 `tab`） |
| `kind` | 工作负载类型 | 仅 `tab=workloads` 生效：`Deployment`、`StatefulSet`、`DaemonSet`；空表示全部；Job/CronJob 忽略该参数 |
| `namespace` | 命名空间 | 空表示全部命名空间；节点与 PV 隐藏该控件并忽略该参数 |
| `name` | 名称过滤 | 十一类资源都支持 |
| `page` / `pageSize` | 分页 | 所有列表使用；切换视图、命名空间、名称时回到第 1 页 |

旧链接 `?tab=workloads&kind=Deployment`、`?tab=pods&page=2` 语义不变。`tab + kind` 同时表达导航项。

### 资源作用域

| 作用域 | 资源 | 命名空间控件 |
| --- | --- | --- |
| 集群级 | 节点、PV | 不显示命名空间选择器，也不显示“不受命名空间影响”的提示；作用域徽章保留 |
| 命名空间级 | 概况、工作负载、Pod、Ingress、Service、ConfigMap、Secret、PVC | 显示「全部命名空间 / 具体命名空间」，切换写入 URL 并驱动列表与导航计数 |

## 2. 组件拆分

| 文件 | 导出 | 职责 |
| --- | --- | --- |
| `detail/index.tsx` | `ClusterDetailPage` | 页面壳：集群元数据、命名空间列表、URL 状态、工具栏与面板路由 |
| `detail/use-cluster-detail-query.ts` | `useClusterDetailQuery` | 读写 `tab/kind/namespace/name/page/pageSize`，保持旧链接兼容 |
| `detail/sections.ts` | `CLUSTER_NAV_GROUPS`、`findClusterNavEntry` | 导航分组、作用域与标签；十一类资源全部可用 |
| `detail/cluster-summary.tsx` | `ClusterSummary` | 只读元数据条与连接状态 |
| `detail/resource-nav.tsx` | `ClusterResourceNav`、`ClusterResourcePicker` | 桌面分组导航 / 窄屏分组下拉；显示各资源计数 |
| `detail/use-resource-counts.ts` | `useClusterResourceCounts` | 用 `pageSize=1` 补齐 Service / ConfigMap / Secret / PV / PVC 的导航计数 |
| `detail/use-cluster-overview.ts` | `useClusterOverview` | 汇总节点、工作负载、Pod、Ingress；分区失败标记与采样 |
| `detail/overview-tab.tsx` | `ClusterOverviewTab` | 概况布局与滚动归属 |
| `detail/overview-cards.tsx` | `ClusterOverviewKpis` 等 | KPI、Pod 分布、工作负载健康、容量、基础组件、Top 榜单 |
| `nodes-tab.tsx`、`workloads-tab.tsx`、`pods-tab.tsx`、`ingresses-tab.tsx` | 对应 Tab 组件 | 四类既有资源：列表、筛选、RowActions、详情抽屉与写操作 |
| `jobs-tab.tsx`、`cronjobs-tab.tsx` | `ClusterJobsTab`、`ClusterCronJobsTab` | Job/CronJob 列表、筛选、分页、详情抽屉与写操作（删除、暂停/恢复、立即运行） |
| `workload-image-dialog.tsx`、`workload-rollback-dialog.tsx` | `WorkloadImageDialog`、`WorkloadRollbackDialog` | 按容器查看/设置镜像版本、列出历史版本并回滚 |
| `services-tab.tsx`、`configmaps-tab.tsx`、`secrets-tab.tsx`、`persistent-volumes-tab.tsx`、`persistent-volume-claims-tab.tsx` | 对应 Tab 组件 | 五类新增资源：列表、筛选、分页、详情抽屉 |
| `key-list.tsx` | `KubernetesKeyList` | ConfigMap / Secret 的键名 chip 列表（不渲染值） |
| `resource-state.tsx` | `ClusterResourceState` | 列表加载 / 错误 / 空态统一外观 |
| `format.ts` | `parseCpuMillis`、`parseMemoryBytes` 等 | 容量解析与格式化、就绪比例、Pod 就绪判断 |

所有资源表格共用全局 `RowActions`：`详情` 是 link 类型按钮，写操作收在 `更多` 下拉里（`forceMenu` + `nowrap`），因此操作列与列表页、告警页等保持一致。详情统一用 `DetailSheet` 抽屉，头部固定、单一滚动区；抽屉内容用「详细信息 / YAML」两个 Tab 切换：详细信息是分组后的结构化字段（状态徽章、可复制标识、标签、容器/端口/路由列表），YAML 是真实 manifest 的代码块（懒加载、铺满 Tab 下方剩余高度、内容超出时只在代码块内部滚动、右上角复制图标带对勾动画；Secret / ConfigMap 的值以 `<redacted>` 呈现）。

## 3. 视觉与交互规范

### 布局与滚动

- 沿用固定高度壳：页面根节点 `flex min-h-0 flex-1 flex-col gap-4`，主体为 `grid min-h-0 flex-1 overflow-hidden lg:grid-cols-[15.5rem_minmax(0,1fr)]`。
- 每个视图只有一个主滚动容器：
  - 概况：卡片区 `overflow-y-auto`（仪表盘滚动当前面板）。
  - 列表：保留 `ListFrame` + `Table` 内层 `overflow-auto`，分页控件固定可达。
  - 左侧导航：`aside` 内层 `overflow-y-auto`，页面级不滚动。
- 窄屏（<1024px）折叠左侧导航为工具栏内的分组 `Select`；表格横向滚动限制在自己的视口内。
- 名称筛选默认收起；展开后输入框与「清空 / 应用」保持同一行（`items-end` + nowrap，输入 `flex-1 min-w-0`，按钮 `shrink-0`），不会出现输入框与按钮错行。

### 间距与层级

- 页面块间距 16px（`gap-4`），卡片间距 12px（`gap-3`），卡片内边距 12px（`Card size="sm"`）。
- 层级顺序：页头（2xl）→ 元数据条（11–12px）→ 工具栏（sm）→ 内容标题（sm）→ 辅助说明（11px）。分组标题用 10px、`tracking-[0.14em]`、大写、`muted-foreground`。
- 标识符、版本、端口、路径统一 `font-mono`；计数、比例、时间统一 `tabular-nums`。

### 状态与反馈

- 状态一律使用共享 `StatusBadge`；Pod 阶段色固定：Running 绿、Pending 琥珀、Succeeded 天蓝、Failed 红、Unknown 灰。
- 加载态：概况卡片用 `Skeleton`；列表用 `ClusterResourceState state="loading"`（旋转图标 + 文案）。
- 错误态：卡片内独立降级（图标 + 说明 + 重试），不会因为一个接口失败让整页报错；四个概况接口全部失败时在概况顶部追加整体提示。集群列表请求遇到网关/后端不可达（网络错误、5xx）时显示标准 502/503 错误页，不再把浏览器的 `Failed to fetch` 原文留在列表里；502/503 页在应用外壳内容区展示，可重试恢复。
- 空态：`ClusterResourceState state="empty"`，包含资源相关的下一步提示（切换命名空间 / 清空名称筛选）。
- 操作反馈：cordon/uncordon、调整副本、滚动重启、删除 Pod 成功后 `toast.success` 并刷新；失败保留对话框内错误。连接测试沿用探测结果对话框。
- 一切操作按钮保持 32px 高（`h-8`），输入框与 `SelectTrigger` 不改高度。

### 可访问性

- 导航是 `nav` + 按钮，当前项 `aria-current="page"`，键盘可聚焦并有焦点环；窄屏下拉带 `aria-label`。
- 状态分布条 `role="img"` 并给出各阶段读数；副本进度 `role="progressbar"` 带 `aria-valuenow`；筛选开关 `aria-controls` 指向真实面板 id，折叠时保留 DOM。
- 颜色不作为唯一信息：状态同时有文字标签；连接状态同时有文案与色点。

### 权限

`cluster:*` 只负责集群接入与连接；资源操作按资源独立授权，页面按权限过滤导航与写入口。

| 权限 | 页面行为 |
| --- | --- |
| `cluster:read` | 查看集群列表/详情、测试连接、进入控制台；无权限时跳 `/forbidden` |
| `cluster:write` / `cluster:delete` | 接入、更新、探测 kubeconfig / 删除集群 |
| `k8s:node:read` / `k8s:node:write` | 节点列表与详情 / cordon、uncordon |
| `k8s:workload:read` / `k8s:workload:write` | 工作负载列表、详情、镜像与历史版本 / 调整副本、滚动重启、设置镜像、版本回滚 |
| `k8s:pod:read` / `k8s:pod:write` / `k8s:pod:logs` | Pod 列表与详情 / 删除 Pod / 读取容器日志 |
| `k8s:ingress:read`、`k8s:service:read` | Ingress、Service 列表与详情 |
| `k8s:configmap:read`、`k8s:secret:read` | ConfigMap、Secret 列表与详情（仅键名） |
| `k8s:pv:read`、`k8s:pvc:read` | PV、PVC 列表与详情 |
| `k8s:job:read` / `k8s:job:write` | Job 列表与详情 / 删除 Job |
| `k8s:cronjob:read` / `k8s:cronjob:write` | CronJob 列表与详情 / 暂停、恢复、立即运行、删除 CronJob |

预置角色：`owner`、`admin` 全部；`member`、`viewer` 拿到全部 `k8s:*:read` 与 `k8s:pod:logs`，没有写权限。缺某个 `k8s:*:read` 时对应导航、Tab 和 YAML 入口不显示，直接访问 URL 回落到概况。服务端仍独立校验，页面只负责隐藏入口。

## 4. 详情抽屉

| 资源 | 详情字段 |
| --- | --- |
| 节点 | 状态、可调度、角色、内网 IP、kubelet 版本、OS 镜像、CPU/内存、Pod 数、运行时长、创建时间 |
| 工作负载 | 命名空间、类型、名称、期望/就绪/已更新/可用副本、镜像、运行时长、创建时间 |
| Pod | 命名空间、阶段、就绪、重启、节点、IP、所属控制器、容器（就绪/重启/镜像）、运行时长、创建时间 |
| Ingress | 命名空间、IngressClass、主机、TLS、规则（host/path → service:port）、运行时长、创建时间 |
| Service | 命名空间、类型、ClusterIP、外部 IP、负载均衡 IP、端口（name/protocol/port/targetPort/nodePort）、选择器、运行时长、创建时间 |
| ConfigMap | 命名空间、data 键名、binaryData 键名、是否不可变、运行时长、创建时间（值不返回） |
| Secret | 命名空间、类型、data 键名、是否不可变、运行时长、创建时间（值永不返回） |
| PV | 名称、容量、StorageClass、访问模式、绑定的 PVC、阶段、回收策略、运行时长、创建时间 |
| PVC | 命名空间、名称、容量、StorageClass、访问模式、绑定的 PV、阶段、运行时长、创建时间 |
| Job | 命名空间、状态、完成数、并行数、成功/失败/活跃、所属 CronJob、开始时间、完成时间、镜像、运行时长、创建时间（关联 Pod 可跳转） |
| CronJob | 命名空间、调度、时区、是否挂起、并发策略、最近调度、活跃 Job、镜像、运行时长、创建时间（关联最近 10 个 Job 可跳转） |

详情抽屉还会按资源加载关联信息（打开时懒加载，加载失败可重试）：

| 资源 | 关联区块 | 交互 |
| --- | --- | --- |
| 工作负载 | 副本集（名称、期望/就绪/当前副本、镜像、创建时间）；Pod 列表（阶段、就绪、重启、节点） | 点击 Pod 跳到 Pod 页签并按 Pod 名称过滤 |
| Job | Pod 列表（阶段、就绪、重启、节点） | 点击 Pod 跳到 Pod 页签并过滤 |
| CronJob | 最近 Job（名称、状态、成功/失败、创建时间） | 点击 Job 跳到 Job 页签并过滤 |
| Ingress | 关联 Service（类型、ClusterIP、端口） | 点击 Service 跳到 Service 页签并过滤 |
| Service | Endpoints（就绪状态、地址、节点、目标对象、端口）；选择器 | 选择器以 `key=value` 标签展示（可复制） |
| ConfigMap / Secret / PVC | 关联工作负载（类型、名称、命名空间），覆盖 volume、projected、envFrom、env.valueFrom、imagePullSecrets（Secret）引用 | 点击跳转到工作负载页签并过滤 |

抽屉标题不再放复制按钮；需要复制的字段（IP、镜像、键名、选择器标签等）在字段旁提供 1.6 秒对勾动画的复制图标。

抽屉内「YAML」Tab 打开时按需请求 `GET /v1/cluster/{cluster_uid}/yaml`，以代码块展示真实 manifest（去掉 `managedFields` 和 `last-applied-configuration`）：代码块铺满 Tab 下方剩余高度，长内容只在代码内容区滚动，外层抽屉不产生第二个滚动条；右上角是复制图标，点击后切换为对勾（1.6 秒后恢复）；Secret / ConfigMap 的 `data`、`binaryData` 值统一显示为 `<redacted>`，并给出「敏感值已隐藏」提示。YAML 超过 512 KiB 时不返回内容，只显示可读错误。切换资源时 Tab 重置回详细信息。

### 全站代码块

- 统一使用 [CodeBlock](../web/src/components/code-block.tsx)：`card`（标题栏 + 复制）与 `plain` 两种形态，支持 `fill`、最大高度、折行与占位符。
- 语法高亮由 [code-highlight.ts](../web/src/lib/code-highlight.ts) 按需加载 Shiki 的 `github-light` / `github-dark`，只输出 `--shiki-light` / `--shiki-dark` 变量，由 `.dark` 切换，不重新高亮、不闪烁；未知语言与超长内容退化为纯文本。
- 已覆盖 YAML、Pod 日志、通知模板预览、告警表达式/摘要/描述、事件原始数据、SSH 命令与输出、数据源 JSON 结果等所有代码展示；表格单元格里的日志片段与字段预览不属于代码块。

## 5. 概况数据与降级

概况只使用现有列表接口，在浏览器侧聚合：

| 区块 | 数据来源 | 降级 |
| --- | --- | --- |
| 节点就绪 / 容量 | `GET /nodes`（一次返回全部） | 失败时该卡片显示重试；容量解析支持 `8`、`7500m`、`16384Mi`、`16Gi`、纯字节 |
| Pod 状态分布 / 重启 | `GET /pods`，每页 100，**最多采样 3 页（300 个）** | 总数超过采样量时标注「统计 300/总数」；单个分页失败保留已取到的样本 |
| 工作负载健康 | `GET /workloads`（一次返回全部） | 按 Deployment/StatefulSet/DaemonSet 分行；`desired=0` 视为健康；点击行跳到对应类型 |
| Ingress 数量 | `GET /ingresses?pageSize=1` 取 `total` | 失败时该 KPI 显示 `—` |
| CPU / 内存容量 | 节点 `capacity` 汇总 | 只展示容量；用量标注「无指标数据源」 |
| 基础组件 | 无接口 | API Server 由「资源查询是否成功」推断；kubelet 由「节点 Ready 数」推断；scheduler / controller-manager / etcd 明确标注未接入 |
| Top 消耗 | 无指标 | 展示节点 Pod 数与 Pod 重启次数两个可推断榜单，并标注 CPU/内存用量需要指标数据源 |

Pod 采样而不全量拉取的原因：`ListPods` 在服务端先列全量再内存分页，单页上限 100；全量统计在大集群上要重复拉取多次全量列表。采样上限是显式常量，后续接口就绪后可整体替换。

## 6. 已实现的后端接口

| 方法 | 路径 | 关键字段 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/v1/cluster/{cluster_uid}/services` | `namespace`、`name`、`type`、`cluster_ip`、`external_ips[]`、`load_balancer_ip`、`ports[]`、`selector`、`created_at` | 支持 `namespace`/`name`/`page`/`pageSize` |
| `GET` | `/v1/cluster/{cluster_uid}/configmaps` | `namespace`、`name`、`data_keys[]`、`binary_data_keys[]`、`immutable`、`created_at` | 只返回键名 |
| `GET` | `/v1/cluster/{cluster_uid}/secrets` | `namespace`、`name`、`type`、`data_keys[]`、`immutable`、`created_at` | 只返回键名，永不返回值 |
| `GET` | `/v1/cluster/{cluster_uid}/persistentvolumes` | `name`、`capacity`、`storage_class`、`access_modes[]`、`claim`、`phase`、`reclaim_policy`、`created_at` | 集群级，无 `namespace` |
| `GET` | `/v1/cluster/{cluster_uid}/persistentvolumeclaims` | `namespace`、`name`、`capacity`、`storage_class`、`access_modes[]`、`volume_name`、`phase`、`created_at` | 支持 `namespace`/`name` 过滤 |
| `GET` | `/v1/cluster/{cluster_uid}/yaml` | `kind`、`namespace`、`name` → `yaml`、`redacted` | 详情 YAML 预览；11 类资源；Secret/ConfigMap 值脱敏；>512 KiB 拒绝 |
| `GET` | `/v1/cluster/{cluster_uid}/workload-detail` | `kind`、`namespace`、`name` → `workload`、`replicasets[]`、`pods[]` | Deployment 通过 ReplicaSet 归属匹配 Pod |
| `GET` | `/v1/cluster/{cluster_uid}/ingress-detail` | `namespace`、`name` → `ingress`、`services[]` | 规则/defaultBackend 引用的 Service |
| `GET` | `/v1/cluster/{cluster_uid}/service-detail` | `namespace`、`name` → `service`、`endpoints[]` | 含 ready/notReady、端口与 targetRef |
| `GET` | `/v1/cluster/{cluster_uid}/configmap-detail` | `namespace`、`name` → `configmap`、`relatedWorkloads[]` | volume/projected/envFrom/env.valueFrom |
| `GET` | `/v1/cluster/{cluster_uid}/secret-detail` | `namespace`、`name` → `secret`、`relatedWorkloads[]` | 含 imagePullSecrets；值仍不返回 |
| `GET` | `/v1/cluster/{cluster_uid}/persistentvolumeclaim-detail` | `namespace`、`name` → `claim`、`relatedWorkloads[]` | volumeMount 引用 |

六类接口的读权限按资源校验（Service、ConfigMap、Secret、PV、PVC 各自 `k8s:*:read`，YAML 按 `kind` 映射），并先按当前租户校验 `cluster_uid`；分页在服务端完成，排序为 namespace/name（PV 按 name）。

### M16 新增接口

| 方法 | 路径 | 关键字段 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/v1/cluster/{cluster_uid}/workload/images` | `kind`、`namespace`、`name` → `images[]`（容器、是否 init、当前镜像、仓库、tag、历史版本） | 历史版本来自该工作负载的 ReplicaSet（Deployment）或 ControllerRevision（StatefulSet/DaemonSet），按版本倒序去重 |
| `PUT` | `/v1/cluster/{cluster_uid}/workload/image` | `kind`、`namespace`、`name`、`container`、`init`、`image` → `workload` | 只改目标容器镜像并返回更新后的工作负载 |
| `GET` | `/v1/cluster/{cluster_uid}/workload/revisions` | `kind`、`namespace`、`name` → `revisions[]`（版本号、历史对象名、当前标记、副本数、镜像、创建时间） | 新版本在前；当前版本不可回滚 |
| `POST` | `/v1/cluster/{cluster_uid}/workload/rollback` | `revision` → `workload` | 把目标版本的 Pod 模板写回工作负载；回滚到当前版本或不存在的版本返回客户端错误 |
| `GET` | `/v1/cluster/{cluster_uid}/jobs` | `namespace`、`name`、`page`、`pageSize` → `items[]`、`total`、`page`、`pageSize` | 状态取 Suspended/Failed/Complete/Running/Unknown/Pending |
| `GET` | `/v1/cluster/{cluster_uid}/job-detail` | `namespace`、`name` → `job`、`pods[]` | Pod 按名称升序 |
| `DELETE` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/job/{name}` | — | 级联删除其 Pod |
| `GET` | `/v1/cluster/{cluster_uid}/cronjobs` | `namespace`、`name`、`page`、`pageSize` → `items[]`、`total`、`page`、`pageSize` | 调度、时区、挂起、并发策略、最近调度、活跃 Job |
| `GET` | `/v1/cluster/{cluster_uid}/cronjob-detail` | `namespace`、`name` → `cronjob`、`jobs[]` | 最近 10 个 Job，创建时间倒序 |
| `PUT` | `/v1/cluster/{cluster_uid}/cronjob/suspend` | `namespace`、`name`、`suspend` → `cronjob` | 暂停/恢复调度 |
| `POST` | `/v1/cluster/{cluster_uid}/cronjob/trigger` | `namespace`、`name` → `job` | 用 `jobTemplate` 立即创建 Job，名称 `<cronjob>-manual-<unix秒>` |
| `DELETE` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/cronjob/{name}` | — | 删除 CronJob（默认级联删除它创建的 Job） |

- 读接口要求对应资源的 `k8s:*:read`（工作负载、Job、CronJob），写接口要求对应资源的 `k8s:*:write`；分页参数与十一类列表一致，排序为 namespace/name。
- YAML 接口的 `kind` 扩展到 `Job`、`CronJob`；Secret/ConfigMap 脱敏规则不变。
- 以上接口都不落库、不新增表与迁移。

## 7. 后续可增强

- `GET /v1/cluster/{cluster_uid}/overview`：一次返回概况聚合，替代前端 5–7 次列表调用。
- `Cluster.metrics_datasource_uid` + `GET /v1/cluster/{cluster_uid}/metrics/summary`：绑定指标数据源，补齐 CPU/内存用量、控制面真实状态与 Top 消耗；前端接入点已留在概况 hook/卡片。
- 节点/Pod/工作负载详情接口（条件、污点、事件、卷、ReplicaSet 历史）。
- HPA、StorageClass、NetworkPolicy 等二级资源。

## 8. 验证

- 真实集群（v1.35.5）浏览器回归：九类资源列表与详情抽屉、详情「详细信息 / YAML」Tab 切换、YAML 铺满剩余高度且内部滚动、GitHub light/dark 高亮切换、复制图标动画、工作负载副本集/Pod 关系与跳转、Ingress→Service 跳转、Service Endpoints 与选择器标签、ConfigMap/Secret/PVC 关联工作负载跳转、RowActions 的详情/更多菜单、Pod 日志共享代码块主题、名称筛选展开/收起与对齐、集群级资源隐藏命名空间选择器、命名空间切换、旧深链、390×844 窄屏无横向溢出、无 console error。
- 请求去重：浏览器实测集群列表 `/v1/clusters` 只请求一次，详情页各接口每个 URL 只请求一次（并发 GET 共享同一个解析结果；settle 后不缓存，后续刷新仍会请求）。
- Go：`go build / vet / test`（新增五类接口的转换、fake clientset、权限门禁测试）；Secret/ConfigMap 值不泄漏由单测与真实接口测试覆盖。
- M16：镜像历史与设置、Deployment/StatefulSet/DaemonSet 回滚（含回滚到当前/不存在版本的客户端错误）、Job 状态判定、CronJob 暂停/立即运行、Job/CronJob YAML 都有 fake clientset 单测；Job/CronJob 页签、镜像与回滚对话框、URL `tab=jobs`/`cronjobs` 与旧深链兼容在浏览器回归。
- Web：`npm run typecheck` / `lint` / `build` 通过。
