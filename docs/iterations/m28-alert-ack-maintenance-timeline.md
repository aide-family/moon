# M28 告警认领、事件时间线与维护窗口

M28 在实时告警与告警事件之上补三件事：值班人员可以**认领**一条事件（ack），每条事件有完整的**处置时间线**，计划维护期间可以用**维护窗口**静默匹配的告警。不改评测、指纹、恢复与投递语义。

## 目标与依赖

操作者可以：

1. 在实时告警列表和事件详情里认领一条 FIRING 事件，记录认领人与认领时间；可以取消认领。
2. 在事件详情里按时间顺序看到这条事件发生了什么：触发、通知、认领、取消认领、介入、抑制、维护窗口抑制、升级、恢复。
3. 新建一个维护窗口（名称、描述、标签/级别/数据源/规则匹配条件、起止时间）。窗口生效期间，匹配的新 FIRING 事件直接以「已抑制」状态落库，不生成任何通知任务；窗口开始前已经 FIRING 的匹配事件也被置为抑制，直到窗口结束。
4. 维护窗口到期后，抑制自动到期，事件按既有规则回到实时列表或按既有恢复流程归档。

依赖：M3（事件与评测回写）、M4（通知任务与投递）、M1（租户与权限）。

## 数据模型

### `alert_events` 新增字段

| 字段 | 说明 |
| --- | --- |
| `acked_by` | 认领人用户 UID，空表示未认领 |
| `acked_at` | 认领时间，空表示未认领 |

认领是事件属性，不是状态机：恢复、抑制与归档都不清除认领信息；历史事件保留认领人展示名。

### `alert_event_actions`（新表）

事件时间线只追加，不修改，不可由租户删除。

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `event_uid` | 事件标识；`event_uid` 建索引 |
| `action` | `FIRING`、`NOTIFIED`、`ACKED`、`UNACKED`、`INTERVENED`、`SUPPRESSED`、`MAINTENANCE`、`ESCALATED`、`RECOVERED`、`RECOVERED_BY_MANUAL` |
| `actor_uid` / `actor_name` | 操作者；系统动作时为空或 `system` |
| `detail` | 可读说明，例如「认领」「维护窗口 `数据库维护` 抑制到 2026-11-03 02:00」 |
| `created_at` | 动作发生时间 |

时间线不落历史分库，事件归档后仍按 `event_uid` 查询主库时间线；保留策略由后续的数据治理里程碑统一处理。

### `alert_maintenance_windows`（新表）

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `name` | 租户内未删除记录唯一 |
| `description` | 备注 |
| `matcher` | JSON：`labels`（键到完整匹配正则）、`level_uids`、`datasource_uids`、`rule_uids`；全部为空表示匹配全部事件 |
| `starts_at` / `ends_at` | 生效区间，左闭右开；`ends_at` 必须晚于 `starts_at` |
| `created_by` | 创建人 UID |

维护窗口不是周期任务，只支持一次性区间；周期维护（每天/每周）留给后续里程碑。

## 接口

| 方法 | 路径 | 权限 |
| --- | --- | --- |
| `POST` | `/v1/alert-events/{uid}/ack` | `alert_realtime:intervene` |
| `POST` | `/v1/alert-events/{uid}/unack` | `alert_realtime:intervene` |
| `GET` | `/v1/alert-events/{uid}/actions` | 实时或历史告警的读权限（按事件所在视图判断） |
| `GET` | `/v1/alert-maintenance-windows` | `alert_maintenance:read` |
| `POST` | `/v1/alert-maintenance-windows` | `alert_maintenance:write` |
| `PUT` | `/v1/alert-maintenance-windows/{uid}` | `alert_maintenance:write` |
| `PUT` | `/v1/alert-maintenance-windows/{uid}/status` | `alert_maintenance:write` |
| `DELETE` | `/v1/alert-maintenance-windows/{uid}` | `alert_maintenance:delete` |

认领与时间线接口不新增能力码，复用实时告警的 `intervene` 权限；维护窗口作为「告警」模块下的新菜单，新增能力 `alert_maintenance`，权限为 `alert_maintenance:read/write/delete` 与菜单权限 `menu:alert_maintenance`。能力进入基础套餐；升级时基础套餐租户自动补上该能力，自定义套餐租户由超级管理员显式开通。

## 行为语义

- **认领**：只接受 `FIRING` 状态。已被他人认领时，其他人仍可再次认领（覆盖认领人并留时间线），不阻塞值班交接；取消认领同样只接受 `FIRING`。重复认领同一人是幂等操作，不重复写时间线。
- **时间线写入**：触发与恢复在 `ApplyReport` / 探测回写的事务内写；通知任务生成成功时写 `NOTIFIED`；认领、介入、抑制、维护窗口在对应操作的事务内写。时间线写入失败与主操作同事务回滚，避免出现「认领了但没有记录」。
- **维护窗口匹配**：事件标签按窗口 `matcher.labels` 的完整匹配正则判断，级别、数据源、规则按 UID 精确匹配，条件之间是 AND；空对象表示不限制。匹配只看事件自身字段与注入后的标签，与通知订阅的匹配实现共用标签匹配函数。
- **窗口生效时**：新事件以 `suppressed=true`、`suppressed_until=ends_at` 落库，写 `MAINTENANCE` 时间线并跳过通知任务；窗口创建/启用时把该窗口内已 FIRING 且匹配的事件同样抑制到 `ends_at`（已经发过的通知不撤回）。窗口停用或删除时立即解除由它设置的抑制（`suppressed=false`，清空 `suppressed_until`），写时间线。
- **恢复**：抑制中的事件恢复时正常走恢复与归档；因为 FIRING 没有生成任务，恢复通知不会补发（沿用现有「没有在途 FIRING 任务就不发恢复通知」规则）。
- **实时列表**：抑制中的事件不出现在实时列表，与现有抑制行为一致；历史分析里仍能按「抑制中」终态看到。

## Web

- 实时告警列表：新增「认领人」列（可自定义列机制，不需要另建列配置）；行操作在现有「介入/抑制/恢复」之外加入「认领 / 取消认领」，认领后行内显示认领人。
- 事件详情：顶部显示认领状态与认领按钮；底部新增「时间线」区块，按时间倒序展示动作、操作者与说明，空态提示「暂无处置记录」。
- 新增页面「维护窗口」（告警菜单下）：列表展示名称、匹配摘要、起止时间、状态（未开始/进行中/已结束/已停用）、创建人；新建/编辑对话框包含名称、描述、匹配条件（标签键值行、级别、数据源、规则多选）与起止时间；行内启停与删除。禁止选择 `ends_at` 早于 `starts_at`。
- i18n 中英文、错误码映射与权限判断按现有页面惯例实现。

## 验收

1. 认领一条 FIRING 事件后，列表与详情显示认领人与认领时间；`alert_event_actions` 新增一条 `ACKED` 记录，操作者是当前用户。取消认领后显示未认领并新增 `UNACKED`。
2. 已恢复或已归档事件调用认领/取消认领返回参数错误；`viewer` 与没有 `alert_realtime:intervene` 的角色返回 403。
3. 事件从 FIRING 到恢复的完整时间线包含触发、通知、认领、恢复记录，顺序与时间一致；服务端不可删除时间线。
4. 新建一个进行中的维护窗口，匹配某个标签。让一条匹配规则在该窗口内 FIRING：事件落库 `suppressed=true`、`suppressed_until=窗口结束`，`notification_tasks` 不新增任务，时间线有 `MAINTENANCE`。
5. 窗口开始前已 FIRING 且匹配的事件在窗口创建后变为抑制；窗口删除或停用后立即解除抑制并回到实时列表。
6. 窗口不匹配的事件（标签/级别/数据源/规则任一不满足）照常通知。
7. 能力开关：从租户能力快照移除 `alert_maintenance` 后，维护窗口接口返回 403 且菜单不出现；恢复后正常。
8. 后端 `make vet` / `make lint` / `make test` 通过；Web `npm run typecheck` / `lint` / `build` 通过；浏览器至少验证认领、时间线、维护窗口 CRUD 与抑制效果。

## 不做什么

- 不做周期维护窗口（cron 式重复）、不做一次性「静默某条告警」之外的临时静默规则复用（现有单事件抑制已覆盖）。
- 不做认领通知（推送「已被认领」消息）、不做工单系统与评论。
- 不改动通知订阅、OnCall、重复通知间隔与恢复通知的既有语义。

## 风险

- 维护窗口与手动抑制共用 `suppressed` 字段：必须用时间线区分来源，解除抑制时只解除由对应窗口设置的事件（以 `suppressed_until` 等于窗口结束时间为判断依据）。
- 时间线表随事件量增长：先建索引与查询分页，不引入分库；容量治理在待做计划书中跟踪。
