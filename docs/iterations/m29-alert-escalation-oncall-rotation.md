# M29 告警升级策略与值班轮换

M29 在 M28 的认领之上补「没人处理就升级」，并把 OnCall 的接收人从固定名单扩展为按**值班轮换**自动取当班成员。升级只生成通知任务，复用既有订阅、通知组、模板与 `merak` 投递链路，不改投递实现。

## 目标与依赖

操作者可以：

1. 新建升级策略：名称、描述、匹配条件（标签、级别、数据源、规则，空表示不限）、启停；策略下配置有序的升级步骤，每步有「事件触发后第几分钟」和一组通知组。
2. 一条 FIRING 事件命中策略后，在步骤到期且事件仍未认领、未恢复时，向该步骤的通知组生成通知任务；被认领或恢复后，后续步骤不再发送。
3. 新建值班轮换：名称、时区、轮换周期（每天/每周）、起始时刻、参与成员的有序列表；系统按时间计算当前当班成员。
4. OnCall 的通知组可以绑定一个值班轮换；投递时接收方追加当班成员的邮箱（与手填接收方合并去重）。

依赖：M28（认领状态与时间线）、M13（OnCall 路由）、M4（通知任务与投递）、M8（通知组与接收方）。

## 数据模型

### `alert_escalation_policies`（新表）

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `name` | 租户内未删除记录唯一 |
| `description` | 备注 |
| `matcher` | JSON：`labels`、`level_uids`、`datasource_uids`、`rule_uids`，语义与维护窗口一致 |
| `status` | `ACTIVE` / `DISABLED`，默认 `ACTIVE` |
| `created_by` | 创建人 |

### `alert_escalation_steps`（新表）

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `policy_uid` | 所属策略 |
| `step_order` | 从 0 开始，同一策略内唯一 |
| `delay_minutes` | 距事件 `starts_at` 的累计分钟数，必须递增；最后一步不得为 0（至少一次升级） |
| `notification_group_uids` | 本步投递的通知组 UID JSON 数组，至少一个 |

### `alert_event_escalations`（新表）

记录「某事件在某个策略的某一步」的发送状态，唯一键 `(event_uid, policy_uid, step_order)`。

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `event_uid` / `policy_uid` / `step_order` | 定位 |
| `due_at` | 事件 `starts_at + delay_minutes` |
| `status` | `PENDING` / `SENT` / `SKIPPED` / `CANCELLED` |
| `sent_at` | 发送时间，可空 |

### `oncall_rotations` 与 `oncall_rotation_members`（新表）

| 字段 | 说明 |
| --- | --- |
| `uid` / `namespace_uid` / `name` | 轮换名称，租户内唯一 |
| `timezone` | IANA 时区名 |
| `period` | `DAILY` / `WEEKLY` |
| `starts_at` | 轮换起点（UTC 保存），第一班从该时刻开始 |
| `handoff_hour` | 交班时刻（0–23，本地时区），默认 0 |
| `status` | `ACTIVE` / `DISABLED` |
| `oncall_rotation_members` | `rotation_uid`、`member_uid`、`position`，同一轮换内成员唯一 |

### `notification_oncall_groups` 新增字段

| 字段 | 说明 |
| --- | --- |
| `rotation_uid` | 可空；非空时该通知组在投递时追加当班成员的邮箱与手机号 |

## 接口

| 方法 | 路径 | 权限 |
| --- | --- | --- |
| `GET/POST` | `/v1/alert-escalation-policies` | `alert_escalation:read` / `alert_escalation:write` |
| `GET/PUT/DELETE` | `/v1/alert-escalation-policies/{uid}` | 读/写/`alert_escalation:delete` |
| `PUT` | `/v1/alert-escalation-policies/{uid}/status` | `alert_escalation:write` |
| `GET/POST` | `/v1/oncall-rotations` | `notification_rotation:read` / `notification_rotation:write` |
| `GET/PUT/DELETE` | `/v1/oncall-rotations/{uid}` | 读/写/`notification_rotation:delete` |
| `PUT` | `/v1/oncall-rotations/{uid}/status` | `notification_rotation:write` |
| `GET` | `/v1/oncall-rotations/{uid}/schedule?days=14` | `notification_rotation:read` |
| `GET` | `/v1/event-escalations?eventUid=` | `alert_realtime:read` |

新增能力：`alert_escalation`（告警模块，菜单「升级策略」）与 `notify_rotation`（通知模块，菜单「值班轮换」），权限分别 `alert_escalation:*`、`notification_rotation:*` 与对应 `menu:*`。两者进入基础套餐，升级时基础套餐租户自动补上。

## 调度与投递语义

- `alioth` 的调度器新增升级 tick，锁 `alioth:lock:schedule:escalation`，每 15 秒执行一次，不续期。一次 tick：
  1. 对仍为 `FIRING` 且未认领、未被抑制的事件，按启用策略的匹配条件补齐 `alert_event_escalations` 的 `PENDING` 行（同一事件同一策略同一步只建一次；补齐失败下一轮重试）。
  2. 取 `due_at <= now` 且状态为 `PENDING` 的行，按步骤通知组生成通知任务并置 `SENT`；事件已认领/恢复/删除时置 `CANCELLED`，不再发送。
  3. 写 `ESCALATED` 时间线，说明策略、步骤与通知组数量。
- 任务生成走与订阅相同的通知组/模板/接收方解析与去重：同一 `(event_uid, phase, group_uid, channel_uid)` 已存在 FIRING 任务时该步骤置 `SKIPPED`，不重复打扰同一渠道。
- 认领或恢复时把该事件所有 `PENDING` 行置 `CANCELLED`；认领后再次取消认领不恢复已取消的步骤（升级只前进不回溯）。
- 事件标签或级别在后续评测中被刷新时，不再重新匹配策略；升级关系在事件首次建立时确定。
- 升级任务与订阅任务共用 `NotificationTask` 与消息日志，页面上按事件指纹与事件 UID 可查；重复通知间隔不作用于升级步骤（步骤自身有延迟）。

### 值班轮换计算

- 当前班次序号：从 `starts_at` 到目标时刻按 `period` 计整周期数，取模成员数得到当班成员；`handoff_hour` 在本地时区生效，交班前的时刻属于上一班。
- OnCall 投递时：`notification.oncall` 判断时段通过后，若绑定 `rotation_uid`，把当班成员的邮箱与手机号并入该通知组接收方；轮换停用、无成员或成员无邮箱时只使用手填接收方，不报错。
- 值班轮换只决定接收人，不改变 OnCall 的时段匹配与订阅路由；没有绑定轮换的 OnCall 行为不变。

## Web

- 新增页面「升级策略」（告警菜单）：列表 + 新建/编辑对话框（基础信息、匹配条件、步骤编辑：延迟分钟与通知组多选，支持增删与排序校验）；行内启停与删除。
- 新增页面「值班轮换」（通知菜单）：列表 + 对话框（基础信息、成员排序、周期与时区）；详情展示未来 14 天排班与「当前值班」标记。
- OnCall 编辑对话框的通知组行新增「值班轮换」选择（可空），保存后列表展示绑定的轮换名。
- 事件详情时间线展示 `ESCALATED` 记录。
- i18n、错误码映射与权限判断按现有惯例。

## 验收

1. 建立一条 5/15 分钟两步升级策略，匹配一条 FIRING 规则。第 5 分钟（允许调度误差）生成第一步通知任务并写 `ESCALATED` 时间线；第 15 分钟生成第二步；同一事件同一策略同一步只有一行 `alert_event_escalations`。
2. 第 5 分钟前认领事件：所有 `PENDING` 行变 `CANCELLED`，第 5/15 分钟都不生成任务。
3. 第 5 分钟前事件恢复：不生成后续步骤任务；恢复通知按既有规则执行。
4. 步骤通知组与订阅已通知的通知组/渠道相同时，该步骤记 `SKIPPED`，消息日志不新增重复任务。
5. 停用策略后不再补齐新事件；已存在 `PENDING` 行按停用时间取消（下次 tick）。
6. 值班轮换：周轮换 3 名成员，当前时刻解析出的当班成员与手工计算一致；绑定到 OnCall 通知组后，FIRING 任务接收方包含当班成员邮箱且与手填邮箱去重；成员无邮箱时只发手填接收方。
7. 存量租户升级后基础套餐租户能看到两个新菜单；自定义套餐租户未开通时接口 403、菜单隐藏。
8. `make vet` / `make lint` / `make test`、Web typecheck/lint/build 通过；浏览器验证两个新页面与 OnCall 轮换选择。

## 不做什么

- 不做值班成员手机短信（短信发送仍未实现）、不做换班申请与告警确认 SLA 报表。
- 不做多策略叠加的优先级仲裁：同一事件可命中多条策略，各自独立升级；同一步骤内通知组去重。
- 不做轮换的节假日/临时换班覆盖（后续里程碑）。

## 风险

- 升级任务与订阅任务共用唯一约束，`SKIPPED` 语义必须在文档与页面上解释清楚，避免用户误以为升级没生效。
- 事件量大时每 15 秒补齐行需要 `(namespace_uid, status, starts_at)` 索引与批量查询；先只处理最近 24 小时 FIRING 事件，超出范围的补行跳过。
