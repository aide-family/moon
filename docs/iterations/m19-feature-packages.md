# M19 功能能力化与租户授权

M19 把菜单从「一个模块一个开关」细化成两级：模块（菜单分组）下面的能力（可单独开通的最小单位）。超级管理员用「套餐」给租户开通能力，套餐带默认授权时长；租户在套餐之上只能**加开**能力，不能减掉套餐能力。能力化同时引入三件事：

1. 菜单权限 `menu:<能力>`：角色持有它才能看到对应菜单入口，接口权限仍由能力里的权限决定；
2. 套餐有效期：`namespaces.package_expires_at` 到期后能力失效，但配置保留；
3. 租户生命周期：平台可以停用（带原因、邮件通知负责人）、启用和归档租户；新增系统只读角色 `SYSTEM_READONLY`。

依赖：M1（身份、租户与 RBAC）、M5（SSH）、M14/M16（Kubernetes）、M18（探测）。登录、租户、角色和权限名本身不变，只在其上叠加能力与生命周期。

## 两级目录

目录全部在 `app/alioth/pkg/perm` 用代码维护，不建表、不由接口增删：

- 模块（`feature.go`）：10 个，决定菜单分组和模块级功能检查，编码为 `alert`、`alert_rule`、`probe`、`datasource`、`notification`、`member`、`role`、`cluster`、`agent`、`ssh`。
- 能力（`capability.go`）：30 个，是套餐勾选和租户差异的唯一单位，`Capability{Code, Module, Name, Description, Permissions}`。
- 权限（`perm.go`）：`resource:action`，挂在能力上；`namespace:*` 不属于任何能力，始终可用；`system:*` 不参与。

### 模块

| 模块 | 编码 | 包含的权限 |
| --- | --- | --- |
| 告警 | `alert` | `alert:read/write/delete/intervene` |
| 告警规则 | `alert_rule` | `alert_rule:*` |
| 探测目标 | `probe` | `probe:*` |
| 数据源 | `datasource` | `datasource:*` |
| 通知 | `notification` | `notification:*` |
| 成员 | `member` | `member:*` |
| 角色 | `role` | `role:*` |
| Kubernetes | `cluster` | `cluster:*` 与全部 `k8s:<资源>:<动作>` |
| 机器 | `agent` | `agent:*` |
| SSH 与 Web Shell | `ssh` | `ssh:*`（含 `ssh:exec`） |

### 能力

| 模块 | 能力 | 名称 | 包含的权限 |
| --- | --- | --- | --- |
| alert | `alert_realtime` | 实时告警 | `alert:read`、`alert:intervene` |
| alert | `alert_history` | 历史告警 | `alert:read` |
| alert | `alert_page` | 告警页面 | `alert:read`、`alert:write`、`alert:delete` |
| alert_rule | `alert_rule` | 告警规则 | `alert_rule:read/write/delete` |
| alert_rule | `alert_dictionary` | 告警字典 | `alert_rule:read/write/delete` |
| probe | `probe_target` | 探测目标 | `probe:read/write/delete` |
| datasource | `datasource` | 数据源 | `datasource:read/write/delete` |
| notification | `notify_channel` | 通知渠道 | `notification:read/write/delete` |
| notification | `notify_template` | 消息模板 | `notification:read/write/delete` |
| notification | `notify_group` | 通知组 | `notification:read/write/delete` |
| notification | `notify_subscription` | 告警订阅 | `notification:read/write/delete` |
| notification | `notify_oncall` | OnCall 值班 | `notification:read/write/delete` |
| notification | `message_log` | 消息日志 | `notification:read/write/delete` |
| member | `member` | 成员 | `member:read/write/delete` |
| role | `role` | 角色 | `role:read/write/delete` |
| cluster | `cluster_access` | 集群接入 | `cluster:read/write/delete` |
| cluster | `k8s_node` | 节点 | `k8s:node:read/write` |
| cluster | `k8s_workload` | 工作负载 | `k8s:workload:read/write` |
| cluster | `k8s_pod` | Pod | `k8s:pod:read/write`、`k8s:pod:logs` |
| cluster | `k8s_ingress` | Ingress | `k8s:ingress:read` |
| cluster | `k8s_service` | Service | `k8s:service:read` |
| cluster | `k8s_configmap` | ConfigMap | `k8s:configmap:read` |
| cluster | `k8s_secret` | Secret | `k8s:secret:read` |
| cluster | `k8s_pv` | PV | `k8s:pv:read` |
| cluster | `k8s_pvc` | PVC | `k8s:pvc:read` |
| cluster | `k8s_job` | Job | `k8s:job:read/write` |
| cluster | `k8s_cronjob` | CronJob | `k8s:cronjob:read/write` |
| agent | `agent_machine` | 机器 | `agent:read/write/delete` |
| ssh | `ssh_command` | SSH 命令 | `ssh:read/write/delete` |
| ssh | `web_shell` | Web Shell | `ssh:exec` |

`Capabilities()` 按上表顺序返回；`CapabilityOf`、`IsCapability`、`CapabilitiesOfModule`、`CapabilityCodesOfModule`、`CapabilityPermissions`（多个能力的权限并集）、`EffectiveCapabilities`（套餐能力 ∪ 额外能力，按目录顺序去重）、`ModulesOfCapabilities`（能力反推模块）是配套查询函数。

### 预置角色

`RolePermissionsForCapabilities(role, capabilities)` 返回预置角色在能力范围内应持有的权限：模块权限与能力权限取交集，再补每个能力的菜单权限；`namespace:*` 始终保留。

`BaseCapabilities()` = 前七个模块（alert、alert_rule、probe、datasource、notification、member、role）的全部能力，共 15 个。基础套餐固定使用这份清单，Kubernetes、机器和 SSH 能力由超级管理员另行开通。

## 菜单权限

`Catalog()` 在租户权限目录末尾追加 30 条菜单权限：`Item{Name: "menu:<能力>", Resource: "menu", Action: <能力编码>}`，因此 `Name` 仍是 `Resource+":"+Action`，前端可以直接按资源分组。

- `MenuPermission(cap)`、`MenuPermissionCapability(permission)`、`IsMenuPermission(name)`、`MenuPermissions()` 是配套函数；`FeatureOf` 对 `menu:<能力>`、能力编码和普通权限都返回所属模块，`namespace:*` 返回空。
- `IsReadOnly` 把 `menu:*` 视为只读；`ReadOnlyNames()` 含全部菜单权限，供平台管理员/系统只读角色的只读视角使用。
- 预置角色默认持有全部菜单权限（owner/admin 来自目录全集，member/viewer 由 `readPermissions` 追加）；实际可见入口 = 角色持有的 `menu:<能力>` ∩ 租户已开通能力 ∩ 角色持有该能力里的权限。菜单权限不是接口权限，`access` 在接口检查里不拿它当普通权限名。

## 系统权限与平台角色

新增 `system:namespace:write`：停用、启用和归档租户，只授予超级管理员。

| 权限 | SUPER_ADMIN | SYSTEM_ADMIN | SYSTEM_READONLY | 作用 |
| --- | --- | --- | --- | --- |
| `system:user:read` | 有 | 有 | 有 | 平台用户列表与详情 |
| `system:user:write` | 有 | 有 | 无 | 封禁/解封普通用户 |
| `system:role:read` | 有 | 有 | 有 | 查看平台角色 |
| `system:role:write` | 有 | 无 | 无 | 授权/撤销系统角色 |
| `system:invite:read` | 有 | 有 | 有 | 平台邀请 |
| `system:namespace:read` | 有 | 有 | 有 | 平台租户列表与只读切入 |
| `system:namespace:write` | 有 | 无 | 无 | 停用/启用/归档租户 |
| `system:package:read` | 有 | 有 | 有 | 功能目录、套餐、租户能力 |
| `system:package:write` | 有 | 无 | 无 | 新建/修改/删除套餐，分配能力与到期时间 |

平台角色固定四档：`SUPER_ADMIN`、`SYSTEM_ADMIN`、`SYSTEM_READONLY`、`USER`。

- `SYSTEM_READONLY` 是系统只读角色：持有全部 `system:*:read`，没有任何写权限；可以查看平台全部信息，也可以跨租户只读切入。
- `IsSystemReadOnly(role)`、`CanReadAnyTenant(role)`（三种平台角色都返回 true）供 access 判断只读通道。
- `IsSystemAdmin(role)` 仍然只认 `SUPER_ADMIN` 和 `SYSTEM_ADMIN`：`SYSTEM_READONLY` 不能管理用户或授权系统角色，避免只读角色被当成管理员。
- `namespace:*` 是租户身份权限，不属于任何能力；租户侧停用接口已移除，租户不能自行解封。

## 数据模型

| 实体 | 关键字段 | 说明 |
| --- | --- | --- |
| `FeaturePackage` | `uid`、`name`（未删除唯一）、`remark`、`is_base`（部分唯一索引保证只有一个基础套餐）、`status`、`duration_days` | 平台级套餐；`duration_days` 是默认授权天数，0 表示永久 |
| `FeaturePackageItem` | `uid`、`package_uid`、`feature`，`(package_uid, feature)` 唯一 | `feature` 存能力编码，不存数组 |
| `NamespaceFeature` | `uid`、`namespace_uid`、`feature`、`enabled`，`(namespace_uid, feature)` 唯一 | 只存租户**额外**能力，`enabled` 恒为 true；旧的 `enabled=false` 记录由 seed 清理且不再参与计算 |
| `Namespace` | 增加 `package_uid`、`package_expires_at`（可空）、`disable_reason`（默认空） | `package_expires_at` 为 NULL 表示永久；到期不清空配置，只让能力失效；停用时写原因，启用清空 |
| `User` | `system_role` 枚举增加 `SYSTEM_READONLY` | 见「迁移」：该列在 PostgreSQL 里是 varchar，不是原生 enum |

约束与语义：

- **有效能力** = 套餐能力 ∪ 租户额外能力，按能力目录顺序排列，不落表。租户停用、已归档、套餐已到期时有效能力为空。
- **已配置能力** = 套餐能力 ∪ 额外能力（不看停用/到期），用于平台管理端展示和 seed 补权限；`NamespaceFeatureState` 返回这一份，并带 `Expired` 标记。
- **只增不减**：租户目标能力必须包含套餐全部能力，差异只存额外能力。`SetNamespaceFeatureState` 校验失败返回 `merr.Client("features must include the package features")`，非法能力编码返回 `merr.Client("invalid capability")`。
- **归档不可恢复**：`ArchiveNamespace` 在事务里软删除租户，再按 `information_schema` 找出所有同时含 `namespace_uid` 与 `deleted_at` 列的业务表逐表软删除（表名白名单校验后引用，分区父表覆盖子分区）。

`bo` 对应字段：`FeaturePackage.DurationDays`；`Feature.Module`；`Namespace.PackageUID/PackageName/PackageExpiresAt/DisableReason`；`NamespaceFeature.PackageFeatures/ExtraFeatures/Features/ExpiresAt/Expired/DurationDays`。时间字段零值表示永久。

## 仓储接口

`repository.FeatureRepo`：

```go
ListPackages(ctx, page, pageSize int, name string) ([]bo.FeaturePackage, int, error)
GetPackage(ctx, uid string) (*bo.FeaturePackage, error)
CreatePackage(ctx, name, remark string, durationDays int32, features []string) (*bo.FeaturePackage, error)
UpdatePackage(ctx, uid, name, remark string, durationDays int32, features []string) (*bo.FeaturePackage, error)
DeletePackage(ctx, uid string) error
NamespaceFeatureState(ctx, namespaceUID string) (*bo.NamespaceFeature, error)
EffectiveCapabilities(ctx, namespaceUID string) ([]string, error)
EffectiveModules(ctx, namespaceUID string) ([]string, error)
EffectiveFeatures(ctx, namespaceUID string) ([]string, error) // Deprecated: EffectiveModules 的旧名字
SetNamespaceFeatureState(ctx, namespaceUID, packageUID string, capabilities []string, expiresAt time.Time) error
ArchiveNamespace(ctx, namespaceUID string) error
CountNamespacesUsingPackage(ctx, packageUID string) (int, error)
```

`repository.NamespaceRepo` 增加 `SetNamespaceStatus(ctx, uid, status, reason string) (*bo.Namespace, error)`；`ListUserNamespaces`/`SelectNamespaces` 过滤停用租户，`ListAllNamespaces`/`GetNamespace` 补齐套餐名、到期时间和停用原因。

## 接口

全部走 `RequireSystem`，JSON 为 camelCase，错误文案英文由前端本地化（实现细节见 proto）：

| 方法与路径 | 权限 | 说明 |
| --- | --- | --- |
| `GET /v1/system/features` | `system:package:read` | 30 个能力，含 `module` 与 `permissions` |
| `GET /v1/system/packages?page&pageSize&name` | `system:package:read` | 套餐分页，含 `durationDays` |
| `GET/POST/PUT/DELETE /v1/system/package[/{uid}]` | 读/写分别校验 | 套餐 CRUD，基础套餐不可改名/删除，被使用的套餐不可删 |
| `GET /v1/system/namespaces` | `system:namespace:read` | 平台租户列表，含 `packageName`、`packageExpiresAt`、`disableReason` |
| `PUT /v1/system/namespace/{uid}/status` | `system:namespace:write` | `{status: ACTIVE\|DISABLED, reason}`；DISABLED 必填原因并邮件通知负责人，启用清空原因 |
| `PUT /v1/system/namespace/{uid}/archive` | `system:namespace:write` | 软删除租户及业务数据，不可恢复 |
| `GET /v1/system/namespace/{uid}/features` | `system:package:read` | `{packageUid, packageName, features, packageFeatures, extraFeatures, expiresAt, expired, durationDays}` |
| `PUT /v1/system/namespace/{uid}/features` | `system:package:write` | `{packageUid, features, expiresAt}`；`expiresAt` 0 表示按套餐时长计算，仍为 0 则永久 |
| `GET /v1/self/permissions` | — | 增加 `features`（有效能力）与 `readOnly`；`permissions` 含 `menu:<能力>`，租户停用/归档/到期时返回空 |

## 权限执行点

1. `access.Checker.require` 顺序：角色持有权限（或平台管理员只读放行）→ `perm.FeatureOf(code)` 在 `EffectiveModules` 中（否则 403 `feature not enabled`）→ 租户成员还要求持有覆盖该权限的 `menu:<能力>`（否则 403 `menu permission denied`）。`namespace:*` 跳过模块与菜单检查；平台管理员只读通道跳过菜单检查，模块检查仍然生效。
2. `role.Authority.ListPermissions` 返回模块权限（按 `EffectiveModules`）与菜单权限（按 `EffectiveCapabilities`）；`UpdateRolePermissions` 拒绝新授予未开通能力的权限，但允许保留角色已持有的旧授权；`Mine` 返回过滤后的权限（含菜单）、有效能力和 `readOnly`。
3. `namespace.Tenancy.CreateNamespace` 写基础套餐与按 `duration_days` 计算的到期时间，预置角色权限取 `RolePermissionsForCapabilities(role, BaseCapabilities())`。
4. `Provisioning`（biz/feature）在设置租户能力时校验套餐存在、目标能力 ⊇ 套餐能力，并按套餐 `durationDays` 计算默认到期时间。
5. 系统只读角色在 `user` 服务里只能由超级管理员授予；`/self/info` 原样返回 `SYSTEM_READONLY` 与只读系统权限。

## 迁移与 seed

`alioth migrate` 的 seed 按以下顺序执行，全部幂等：

1. 权限目录：`Catalog()` 现在包含 30 个菜单权限，缺失的 `permissions` 行自动补建。
2. 基础套餐：不存在则按 `BaseCapabilities()` 创建；已存在只补缺失能力，不删除管理员加的能力，不改名称/备注/时长。
3. 旧模块编码展开：`feature_package_items` 和 `namespace_features` 里的模块级编码展开成该模块全部能力，再删除旧行；`enabled=false` 的旧差异记录整体删除（能力只增不减，不再有减法）。
4. 存量租户：`package_uid = ''` 的租户迁到基础套餐，并额外写入 cluster（12）+ agent（1）+ ssh（2）共 15 个能力，保持升级前能力；`package_expires_at` 保持 NULL（永久）。
5. 预置系统角色：按 `RolePermissionsForCapabilities(role, 已配置能力)` 补齐权限与菜单权限，只加不删。
6. 自定义角色菜单补齐：角色持有某能力的任一权限就补 `menu:<能力>`，只补缺失项。
7. 旧 K8s 权限展开、超级管理员补齐、通知预设等既有迁移保持原样。

### 真库验证

本次在临时库上真实执行了升级路径，验证后清理：

- 新建库：`migrate` 成功；`feature_package_items` 为 15 个基础能力；`permissions` 83 条（53 个原有 + 30 个菜单）；`namespaces` 新列与 `feature_packages.duration_days` 均创建成功。
- 存量库（复制本地开发库：12 个租户、7 条模块级基础套餐明细、33 条模块级租户差异、53 条权限）：`migrate` 成功；基础套餐明细变为 15 个能力；租户差异变为 165 条能力记录（11 个租户 × 15），`enabled=false` 为 0；12 个租户全部写上 `package_uid`；菜单权限 30 条、角色菜单授权 1397 条；重复执行 3 次行数稳定（`role_permissions` 3249 总行、无重复授权）。
- `User.system_role`：ent v0.14.6 的 PostgreSQL 方言把 `field.Enum` 落成 **varchar**（源码 `dialect/sql/schema/postgres.go` 明确保留旧行为），新建库与存量库里都没有原生 enum 类型。因此这次不需要 `ALTER TYPE ... ADD VALUE`，也不存在 enum 事务兼容问题；直接插入 `SYSTEM_READONLY` 用户验证通过。若未来改成原生 enum，需要单独评估加值的迁移窗口。
- 能力仓储集成测试（真库）覆盖：基础能力、套餐/额外能力分配、只增不减拒绝、非法能力拒绝、菜单权限补齐、到期失效、停用原因、`ListUserNamespaces` 过滤、平台租户列表套餐字段、启用恢复、归档级联软删除。

## 回滚

- 回滚代码不会删除新表和新列；旧版本忽略 `duration_days`、`package_expires_at`、`disable_reason` 和菜单权限。
- 回滚前如果套餐明细已经写成能力编码，旧代码按模块编码读取会得到空功能；需要保留旧版本行为时不要回滚到能力化之前的代码，或先导出套餐配置。

## Web

- 套餐管理改为两级勾选（模块 → 能力），增加「授权时长（天，0=永久）」；基础套餐不可改名/删除但可改能力与时长；被使用的套餐删除失败显示后端错误。
- 租户管理展示套餐、到期时间、停用原因；「功能授权」弹窗里套餐自带能力锁定不可取消，只能额外勾选，并选择到期时间；停用/启用/归档是独立高风险操作。
- 菜单入口改为 `session.can("menu:<能力>") && session.hasFeature(<能力>) && session.can(<接口权限>)`；系统只读角色只读横幅与角色徽标同步展示。
- 平台只读视角（含 `SYSTEM_READONLY`）只显示只读入口与只读横幅；DISABLED 租户禁用只读查看与功能授权。

## 验收

1. 新库 `alioth migrate` 后基础套餐含 15 个基础能力、30 个菜单权限存在、`SYSTEM_READONLY` 可写入用户。
2. 存量库迁移后行为不变：原租户仍可使用 cluster/agent/ssh 全部能力；基础套餐原有配置不被删除；重复执行不产生重复数据。
3. 目标能力缺少套餐能力时 `PUT /v1/system/namespace/{uid}/features` 返回 400 `features must include the package features`；保存后 `extra_features` 只含套餐之外的能力，`expires_at` 与 `expired` 正确。
4. 套餐到期或租户停用后：`EffectiveCapabilities`/`EffectiveModules` 为空，业务接口 403 `feature not enabled`，`/self/permissions` 返回空；启用并续期后恢复。
5. 停用租户必须带原因并写 `disable_reason`，租户列表可见；启用清空原因；停用租户不出现在成员租户列表和切换列表。
6. 归档后租户与业务数据软删除、无法再读取；归档不可重复。
7. 角色只持有接口权限但没有 `menu:<能力>` 时调用业务接口返回 403 `menu permission denied`；补上菜单权限后通过。
8. `SYSTEM_READONLY` 用户可读平台信息并跨租户只读，调用任一写接口 403；`SYSTEM_ADMIN` 调 `PUT /v1/system/namespace/{uid}/status` 返回 403。
9. `cd app/alioth && go build ./... && go test ./...`、`cd api && go test ./...`、`gofmt -l`、`scripts/check-naming.sh`、`buf lint`、`make wire` 全部通过。
