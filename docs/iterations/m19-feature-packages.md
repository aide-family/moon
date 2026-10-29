# M19 功能套餐与租户功能授权

M19 把菜单模块化成为「功能」，由超级管理员用「套餐」给租户开通功能；租户只能基于已开通的功能给角色授权。新租户默认使用「基础套餐」，只含告警、告警规则、探测、数据源、通知、成员与角色，Kubernetes、机器和 SSH 由超级管理员另行开通。这一版只叠加授权模型，登录、租户、角色和权限名不变。

## 目标与依赖

超级管理员可以：

1. 查看代码维护的功能目录（`GET /v1/system/features`），每个功能带它的权限名；
2. 新建、修改、删除功能套餐，给套餐勾选功能；
3. 在租户管理页看到租户当前套餐和有效功能，重新选择套餐并逐个增删功能；
4. 系统管理员只能查看套餐和租户功能，不能修改（`system:package:read` 与 `system:package:write` 分离）。

租户管理员可以：

1. 在角色授权页只看到当前租户已开通功能的权限；
2. 只能把已开通功能的权限授给角色，未开通功能的权限提交会被拒绝；
3. 当前用户权限（`/v1/self/permissions`）只返回已开通功能的权限，菜单与接口一致。

依赖：M1（身份、租户与 RBAC）、M5（SSH 功能）、M14（Kubernetes 功能）、M18（探测功能）。M19 只在这些里程碑已有的菜单和权限名之上加一层「功能是否开通」的判断。

## 功能目录

功能目录在代码里维护（`app/alioth/pkg/perm/feature.go`），不建表、不由接口增删：

| 功能 | 编码 | 包含的租户权限 |
| --- | --- | --- |
| 告警 | `alert` | `alert:read`、`alert:write`、`alert:delete`、`alert:intervene` |
| 告警规则 | `alert_rule` | `alert_rule:*` |
| 探测目标 | `probe` | `probe:*` |
| 数据源 | `datasource` | `datasource:*` |
| 通知 | `notification` | `notification:*` |
| 成员 | `member` | `member:*` |
| 角色 | `role` | `role:*` |
| Kubernetes | `cluster` | `cluster:*` 与全部 `k8s:<资源>:<动作>` |
| 机器 | `agent` | `agent:*` |
| SSH 与 Web Shell | `ssh` | `ssh:*`（含 `ssh:exec`） |

- `namespace:*` 不属于任何功能，始终可用；`system:*` 不参与功能。
- `FeatureOf(permission)` 把租户权限映射到功能，`namespace:*` 返回空表示不受功能限制。
- `BaseFeatures()` = `alert`、`alert_rule`、`probe`、`datasource`、`notification`、`member`、`role`。
- 功能列表顺序固定为上面的表格顺序；权限按 `Catalog()` 顺序列出，前端按 `code` 分组。

新增系统权限：

| 权限 | SUPER_ADMIN | SYSTEM_ADMIN | 作用 |
| --- | --- | --- | --- |
| `system:package:read` | 有 | 有 | 查看功能目录、套餐、租户功能 |
| `system:package:write` | 有 | 无 | 新建、修改、删除套餐，设置租户功能 |

## 数据模型

新增三张表：

| 表 | 关键字段 | 约束 |
| --- | --- | --- |
| `feature_packages` | `uid`、`name`、`remark`、`is_base`、`status`、审计与软删除 | `name` 在 `deleted_at IS NULL` 下唯一；`is_base = true AND deleted_at IS NULL` 的部分唯一索引保证基础套餐只有一个 |
| `feature_package_items` | `uid`、`package_uid`、`feature` | `(package_uid, feature)` 唯一 |
| `namespace_features` | `uid`、`namespace_uid`、`feature`、`enabled` | `(namespace_uid, feature)` 唯一 |

`namespaces` 增加 `package_uid`（默认空）。

约束与语义：

- 套餐功能不存数组，明细落在 `feature_package_items`，重复功能由唯一约束挡住。
- `namespace_features` 保存租户相对套餐的差异：`enabled = true` 表示在套餐之外加开，`false` 表示从套餐里减掉；没有记录表示跟随套餐。
- 有效功能 = 套餐功能 + 租户差异，按功能目录顺序排列，不落表。每次读取时计算，读的是两张小表。
- `package_uid` 为空只出现在升级前的存量租户，由 seed 迁移；新租户创建时必写基础套餐。

## 接口

全部加到 `NamespaceService`，走 `RequireSystem`，和新 operation 一起加入 auth 的 `namespaceOptional`，因此不要求 `X-Namespace`。错误文案用英文，由前端本地化。

| 方法与路径 | 权限 | 说明 |
| --- | --- | --- |
| `GET /v1/system/features` | `system:package:read` | 功能目录 |
| `GET /v1/system/packages?page&pageSize&name` | `system:package:read` | 套餐分页 |
| `GET /v1/system/package/{uid}` | `system:package:read` | 套餐详情 |
| `POST /v1/system/package` | `system:package:write` | 新建套餐 |
| `PUT /v1/system/package/{uid}` | `system:package:write` | 修改套餐 |
| `DELETE /v1/system/package/{uid}` | `system:package:write` | 删除套餐 |
| `GET /v1/system/namespace/{uid}/features` | `system:package:read` | 租户功能开通情况 |
| `PUT /v1/system/namespace/{uid}/features` | `system:package:write` | 原子设置租户套餐与有效功能 |

JSON 形状：

```json
// GET /v1/system/features
{"items":[{"code":"alert","name":"告警","description":"查看告警事件、处理与干预告警。","permissions":["alert:read","alert:write","alert:delete","alert:intervene"]}]}

// GET /v1/system/packages
{"items":[{"uid":"...","name":"基础套餐","remark":"...","features":["alert","alert_rule"],"isBase":true,"status":"ACTIVE","createdAt":1700000000}],"total":1,"page":1,"pageSize":20}

// GET /v1/system/package/{uid}、POST、PUT
{"package":{"uid":"...","name":"...","remark":"...","features":["alert"],"isBase":false,"status":"ACTIVE","createdAt":1700000000}}

// GET /v1/system/namespace/{uid}/features、PUT 同形状
{"packageUid":"...","packageName":"基础套餐","features":["alert","alert_rule"],"packageFeatures":["alert","alert_rule"]}
```

`PUT /v1/system/namespace/{uid}/features` 的请求体：

```json
{"packageUid":"...","features":["alert","alert_rule","cluster"]}
```

`features` 是目标有效功能，服务端和套餐功能对比后换算差异：目标有、套餐没有写 `enabled = true`；目标没有、套餐有写 `enabled = false`；其余不写。请求在事务里换套餐、重写差异；提交后返回最新状态。

`/v1/self/permissions` 的响应增加 `features`：

```json
{"permissions":["alert_rule:read"],"readOnly":false,"namespace":{"uid":"..."},"features":["alert","alert_rule"]}
```

没有选择租户时 `permissions` 和 `features` 都为空数组。

## 权限执行点

1. `access.Checker` 增加 `Features repository.FeatureRepo`。`require` 在权限通过之后检查 `perm.FeatureOf(code)` 对应的功能是否已开通，未开通返回 403 `feature not enabled`。系统管理员的跨租户只读同样受限，避免看到点进去就 403 的菜单。
2. `role.Authority` 增加 `Features` 仓储：
   - `ListPermissions` 只返回已开通功能的权限；
   - `UpdateRolePermissions` 拒绝新授予未开通功能的权限（`feature not enabled`），但允许保留角色已经持有的旧授权：关闭功能后编辑角色不能因为前端原样回传旧权限而保存失败；
   - `Mine` 按已开通功能过滤角色权限和平台管理员的只读目录，同时把有效功能交给 `/self/permissions`。
3. `data.NamespaceRepo.CreateNamespace` 在创建租户的事务里写入基础套餐 `package_uid`，预置角色权限按基础套餐当前的功能明细裁剪（`RolePermissionsForFeatures`）；明细缺失时回落到 `BaseFeatures()`。
4. seed：
   - 基础套餐不存在才创建（名称、备注、功能为 `BaseFeatures()`），已存在时不覆盖；
   - 把 `package_uid` 为空的存量租户迁到基础套餐，并写 `cluster`、`agent`、`ssh` 的 `enabled = true` 差异，保持升级前能力；
   - `syncTenantRoles` 只补有效功能内的权限；已有但不在有效功能内的授权保留，由权限检查拒绝。
5. 修改租户功能（`SetNamespaceFeatureState`）在同一个事务里补齐预置系统角色缺失的权限，只加不删。这样超级管理员开通功能后，租户管理员无需等下一次 `migrate` 就能在角色授权页看到并授予新功能的权限。

## 迁移

1. 发布前执行 `alioth migrate`。ent 会新建三张表并给 `namespaces` 加 `package_uid` 列（默认空）。
2. 同一次 migrate 的 seed 会创建基础套餐、迁移存量租户并加开 `cluster`/`agent`/`ssh`。迁移可重复执行：已迁移的租户 `package_uid` 非空，不再命中。
3. 多副本同时启动时，基础套餐的 `is_base` 部分唯一索引兜底：冲突的一方回滚后读取已存在的套餐，不覆盖配置。
4. 升级后新租户默认只有基础套餐的功能；存量租户行为不变（有 Kubernetes、机器和 SSH）。
5. 回滚代码不会删除新表；旧版本忽略 `package_uid`、新表，行为回到升级前。

## Web

- 系统管理增加「功能套餐」入口（`system:package:read` 可见，`system:package:write` 才能编辑）：套餐列表、新建/编辑对话框（名称、备注、功能多选）、删除。基础套餐的改名和删除按钮禁用，功能与备注可编辑；被租户使用的套餐删除按钮禁用。
- 租户管理增加「功能」入口：展示当前套餐、套餐功能与有效功能，选择套餐并逐个勾选/取消功能，保存即提交目标有效功能。
- 角色授权页按功能分组展示权限，只展示已开通功能的权限。
- 菜单按 `/self/permissions` 的 `permissions` 与 `features` 生成，未开通功能不显示入口。
- 功能名称和说明直接使用接口返回的 `name`、`description`；错误文案按错误码本地化。

## 验收

1. `alioth migrate` 后基础套餐存在且 `is_base = true`，只有一个；`feature_package_items` 含七个基础功能；存量租户的 `package_uid` 指向基础套餐，并有 `cluster`/`agent`/`ssh` 的 `enabled = true` 差异。
2. 新建租户的 `package_uid` 是基础套餐；预置角色权限不含 `cluster:*`、`k8s:*`、`agent:*`、`ssh:*`；租户管理员在角色授权页看不到这些权限。
3. 超级管理员新建套餐并给租户勾选 `cluster` 后，该租户 `GET /v1/system/namespace/{uid}/features` 的 `features` 含 `cluster`；预置系统角色在同一事务里补上 `cluster:*` 与 `k8s:*`，租户成员按角色获得 Kubernetes 访问，角色授权页与 `/self/permissions` 出现 `k8s:*`。
4. 取消租户的 `alert_rule` 后，即使成员角色持有 `alert_rule:read`，`GET /v1/alert-rules` 返回 403 `feature not enabled`；`/self/permissions` 不再返回该权限。
5. 系统管理员（`SYSTEM_ADMIN`）可以读套餐和租户功能，调用写接口返回 403；超级管理员可以写。
6. 基础套餐改名、删除返回 400；被租户使用的套餐删除返回 400；删除未使用的套餐成功。
7. `PUT /v1/system/namespace/{uid}/features` 同时修改套餐和有效功能：一次请求后套餐、差异、有效功能三者一致；功能列表按目录顺序返回。
8. 未带 `X-Namespace` 调用 `/v1/system/*` 不返回租户头缺失错误。
9. 后端 `go build ./...`、`go test ./...`、`buf generate`、`gofmt -l`、`scripts/check-naming.sh`、`buf lint`、`make wire` 全部通过。
