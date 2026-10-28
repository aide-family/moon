# M16 工作负载运维：镜像版本、Job/CronJob 与版本回滚

通用规范见 [conventions.md](conventions.md)。本文在 M14 的 Kubernetes 控制台上补三类运维能力：
工作负载镜像版本查看与设置、Job/CronJob 管理、工作负载版本回滚。控制面仍然是 `alioth`，浏览器仍然只访问 Web。

## 目标与依赖

依赖 M14（集群接入、工作负载/Pod 列表与详情抽屉、`cluster:*` 权限）与
[kubernetes-console-design.md](../kubernetes-console-design.md) 定义的控制台结构与 URL 契约。

做：

- 镜像版本：按工作负载列出租户可见的每个容器当前镜像与历史版本（Deployment 取 ReplicaSet、
  StatefulSet/DaemonSet 取 ControllerRevision），支持按容器设置新镜像。
- Job/CronJob：列表、详情抽屉、YAML 预览；Job 删除；CronJob 暂停/恢复、立即运行、删除。
- 版本回滚：列出工作负载的历史版本（版本号、镜像、副本数、创建时间），回滚到指定版本。

不做：

- 不查询镜像仓库，不拉取 tag 列表；「历史版本」只来自集群里已存在的历史对象。
- 不做 Job/CronJob 的创建与模板编辑、不做任意 YAML 编辑、不做 CronJob 调度规则解析。
- 不做 Helm/GitOps、不做 Pod exec、不做指标图表。
- 不回滚 Pod 本身，也不回滚不属于 Deployment/StatefulSet/DaemonSet 的控制器。

## 数据与业务对象

数据层只读集群对象，不落库，不新增表和迁移；`bo` 新增以下业务对象
（`app/alioth/pkg/bo/cluster.go`）：

| 业务对象 | 关键字段 |
| --- | --- |
| `KubeWorkloadImage` | `Container`、`Init`、`Image`、`Repository`、`Tag`、`Versions []KubeImageVersion` |
| `KubeImageVersion` | `Image`、`Revision`、`CreatedAt` |
| `KubeWorkloadRevision` | `Revision`、`Name`、`Current`、`Replicas`、`Images`、`CreatedAt` |
| `KubeJob` | `Namespace`、`Name`、`Status`、`Completions`、`Parallelism`、`Active`、`Succeeded`、`Failed`、`OwnerKind`、`OwnerName`、`StartTime`、`CompletionTime`、`CreatedAt`、`Images` |
| `KubeJobDetail` | `Job KubeJob`、`Pods []KubePod` |
| `KubeCronJob` | `Namespace`、`Name`、`Schedule`、`TimeZone`、`Suspend`、`ConcurrencyPolicy`、`LastScheduleTime`、`ActiveCount`、`Images`、`CreatedAt` |
| `KubeCronJobDetail` | `CronJob KubeCronJob`、`Jobs []KubeJob` |

- 时间统一 Unix 秒，未设置时为 0；`Images` 按容器顺序稳定输出。
- Job 状态判定顺序：`spec.suspend=true` → `Suspended`；`failed > backoffLimit` → `Failed`；
  `succeeded >= completions`（`completions=0` 时 `succeeded>0`）→ `Complete`；`active>0` → `Running`；
  有开始时间 → `Unknown`；否则 `Pending`。
- Job 的 `OwnerKind`/`OwnerName` 取直接控制器的 ownerReference（通常是 CronJob）；
  `Completions`/`Parallelism` 未设置时为 0。
- 工作负载历史版本只返回属于该工作负载的历史对象：Deployment 用 `ownerReferences`（UID）匹配
  ReplicaSet，并取 `deployment.kubernetes.io/revision` 注解；StatefulSet/DaemonSet 用
  `ownerReferences`（UID）匹配 ControllerRevision，版本号取 `revision` 字段。
  没有 ownerReference 命中的历史对象不返回。
- 版本列表的「当前版本」：Deployment 取自身 `deployment.kubernetes.io/revision` 注解；
  StatefulSet 取 `status.currentRevision`（缺失时回退 `updateRevision`）；DaemonSet 没有
  当前版本字段，取最高 ControllerRevision 号。ControllerRevision 不记录副本数，
  StatefulSet/DaemonSet 的版本副本数展示工作负载当前期望副本数。

## 接口

全部路径中的 `{cluster_uid}` 按当前租户校验，跨租户的 `cluster_uid` 按不存在处理。

工作负载镜像与回滚：

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/v1/cluster/{cluster_uid}/workload/images` | `cluster:read` | `kind`、`namespace`、`name`；返回容器镜像与历史版本 |
| `PUT` | `/v1/cluster/{cluster_uid}/workload/image` | `cluster:write` | `kind`、`namespace`、`name`、`container`、`init`、`image`；返回更新后的工作负载 |
| `GET` | `/v1/cluster/{cluster_uid}/workload/revisions` | `cluster:read` | 返回历史版本列表，新版本在前 |
| `POST` | `/v1/cluster/{cluster_uid}/workload/rollback` | `cluster:write` | `revision` 必填且 ≥1；返回更新后的工作负载 |

Job/CronJob：

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/v1/cluster/{cluster_uid}/jobs` | `cluster:read` | `namespace`、`name`、`page`、`pageSize`；返回 `items`、`total`、`page`、`pageSize` |
| `GET` | `/v1/cluster/{cluster_uid}/job-detail` | `cluster:read` | `namespace`、`name`；返回 Job 与其 Pod |
| `DELETE` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/job/{name}` | `cluster:write` | 删除 Job（`propagationPolicy` 用默认后台级联） |
| `GET` | `/v1/cluster/{cluster_uid}/cronjobs` | `cluster:read` | `namespace`、`name`、`page`、`pageSize` |
| `GET` | `/v1/cluster/{cluster_uid}/cronjob-detail` | `cluster:read` | 返回 CronJob 与其最近 Job |
| `PUT` | `/v1/cluster/{cluster_uid}/cronjob/suspend` | `cluster:write` | `namespace`、`name`、`suspend`；返回更新后的 CronJob |
| `POST` | `/v1/cluster/{cluster_uid}/cronjob/trigger` | `cluster:write` | 用 `jobTemplate` 立即创建一个 Job，返回新 Job |
| `DELETE` | `/v1/cluster/{cluster_uid}/namespace/{namespace}/cronjob/{name}` | `cluster:write` | 删除 CronJob |

- 分页沿用 `page` / `pageSize`，按 namespace、name 升序；`name` 为子串过滤。
- `job-detail` 的 Pod 按 name 升序；`cronjob-detail` 的 Job 最多 10 条，按创建时间倒序。
- YAML 接口扩展 `kind=Job`、`kind=CronJob`；Secret/ConfigMap 的脱敏规则不变。
- 回滚：Deployment 把目标 ReplicaSet 的 Pod 模板写回 Deployment；StatefulSet/DaemonSet 把目标
  ControllerRevision 的模板写回工作负载。回滚到当前版本返回客户端错误；目标版本不存在返回客户端错误。
- 设置镜像：`container` 必须存在且按 `init` 区分普通容器与 init 容器；只改目标容器的镜像，其它字段不动。
- 立即运行：Job 名为 `<cronjob>-manual-<unix秒>`，长度超过 63 时截断前缀；复制 `jobTemplate`
  的标签并追加 `cronjob.kubernetes.io/instantiate=manual`；不设置 CronJob ownerReference，
  避免被历史清理策略回收。
- 401/403 沿用 `cluster` 现有映射：凭据被拒绝、凭据无权限分别返回客户端错误，不清平台登录态。

## 权限

不新增权限。读取新增接口用 `cluster:read`；设置镜像、回滚、删除 Job、暂停/恢复、立即运行、
删除 CronJob 用 `cluster:write`。预置角色不变，`member`/`viewer` 只读。

## Web

资源导航「负载」分组增加 Job / CronJob；URL 用 `tab=jobs`、`tab=cronjobs`，与 `namespace`、
`name`、`page`、`pageSize` 的语义一致；Job/CronJob 忽略 `kind` 参数。

- 工作负载行操作在「更多」里增加「镜像」与「回滚」：
  - 镜像对话框列出每个容器的当前镜像、历史版本，允许逐容器填写新镜像；只对改动过的容器调用设置接口，
    任一失败停止并保留对话框内的错误。
  - 回滚对话框列出历史版本（版本号、镜像、创建时间、当前标记），确认后调用回滚并刷新列表与详情。
- Job 页签：名称、命名空间、状态、完成数、成功/失败/活跃、所属 CronJob、开始时间、创建时间；
  行操作为「详情 + 更多（删除）」；详情抽屉展示字段与 Pod，Pod 可跳转到 Pod 页签。
- CronJob 页签：名称、命名空间、调度、时区、挂起、并发策略、最近调度、活跃 Job、创建时间；
  行操作为「详情 + 更多（暂停/恢复、立即运行、删除）」；详情抽屉展示字段与最近 Job。
- Job/CronJob 的详情抽屉同样有「详细信息 / YAML」两个 Tab；YAML 走 `/yaml` 接口。
- 写操作按钮按 `cluster:write` 隐藏，服务端仍独立校验；所有操作成功后 `toast.success` 并刷新。
- 沿用 `web/AGENTS.md` 的表格、滚动、32px 控件与 `RowActions`/`DetailSheet` 约束。

## 验收标准

- [ ] 工作负载镜像接口返回每个容器当前镜像与历史版本；历史版本只来自该工作负载的历史对象，去重后新版本在前。
- [ ] 设置镜像后工作负载 Pod 模板中的目标容器镜像变化，列表与详情刷新后可见；不存在的容器返回客户端错误。
- [ ] 修改镜像两次后回滚到上一个版本，Pod 模板恢复为目标版本的镜像，`kubectl rollout history` 出现新版本。
- [ ] 回滚到当前版本或不存在的版本返回客户端错误，工作负载保持不变。
- [ ] Job 列表分页、名称过滤、状态判定正确；Job 详情返回其 Pod；删除 Job 后列表刷新。
- [ ] CronJob 列表展示调度、时区、挂起状态与活跃 Job；暂停后 `suspend=true`，恢复后为 `false`。
- [ ] 立即运行生成独立 Job，名称符合 `<cronjob>-manual-<unix秒>`，带 `cronjob.kubernetes.io/instantiate=manual` 标签。
- [ ] 未登录、无 `cluster:read` 的账号访问新增接口被拒绝；`viewer` 不能设置镜像、回滚、删除 Job/CronJob、暂停或立即运行。
- [ ] Job/CronJob 的 YAML 预览返回真实 manifest；Secret/ConfigMap 值仍然不出现在响应里。
- [ ] `go test ./...`、`make lint`、Web 的 `typecheck`/`lint`/`build` 全部通过。
