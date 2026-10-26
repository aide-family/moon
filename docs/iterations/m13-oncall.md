# M13 OnCall 值班路由

通用规范见 [conventions.md](conventions.md)。

## 目标与依赖

在「通知」下增加 OnCall：按事件标签、告警级别、数据源和数据源级别订阅告警，绑定多个通知组，并让每个通知组只在配置的时段接收告警；同时增强告警订阅与 OnCall 的重复通知间隔和接收方去重。依赖 M4 的通知任务与 M8 的订阅路由。

## 范围与不做

做：

- 新增 `NotificationOnCall`，按标签、告警级别、数据源和数据源级别匹配告警事件。
- 一个 OnCall 绑定多个通知组，每个绑定可以配置多条接收时段。
- 时段由星期几和一天内的起止时间组成，支持跨午夜；没有时段表示全天候接收。
- 事件命中 OnCall 时，只给当前时刻处于接收时段的绑定生成通知任务。
- 同一事件同时命中订阅和 OnCall 时按接收方去重：同一邮箱/手机号只发一次，同一 Webhook 渠道只安排一次。
- 告警订阅和 OnCall 都支持 `repeat_interval_seconds`（0～43200，最大 12 小时）：同一告警指纹在命中来源的最大间隔内不再生成 FIRING 任务。
- 恢复通知不受重复间隔限制，但只在同一指纹存在仍在投递链路上的 FIRING 任务时发送。
- 管理台新增「通知 → OnCall」页面，支持增删改查、启停和时段编辑；通知组、数据源和数据源级别使用可搜索、可全选的下拉框。
- 保留现有 `NotificationSubscription` 路由；两套路由独立匹配、合并去重。

不做：

- 不替换或迁移现有告警订阅。
- 不做节假日、按日期区间或轮班自动交接；时段只按星期几和时刻匹配。
- 不新增通知渠道、模板或任务类型。
- 不按 OnCall 单独配置接收人或渠道；接收能力沿用绑定的通知组。

## 数据

| 实体 | 要点 |
| --- | --- |
| `NotificationOnCall` | Namespace 内唯一名称，可启停；标签匹配器、告警级别、数据源、数据源级别、IANA 时区、重复通知间隔、备注 |
| `NotificationOnCallGroup` | OnCall 到通知组的绑定；`time_windows` 保存该组的时段 JSON，唯一约束 `(oncall_uid, group_uid)` |
| `NotificationSubscription` | 新增数据源级别条件和重复通知间隔字段 |
| `NotificationTask` | 新增 `fingerprint`，用于同一告警的重复通知判断；索引 `(namespace_uid, fingerprint, phase)` |

时段 JSON 形态：

```json
[
  {"weekdays": ["MONDAY", "TUESDAY"], "start": "09:00", "end": "18:00"},
  {"weekdays": ["FRIDAY"], "start": "22:00", "end": "06:00"}
]
```

- `weekdays` 取 `MONDAY`…`SUNDAY`，保存时按周一到周日排序去重。
- 时间为 `HH:mm`，边界左闭右开。`start` 晚于 `end` 表示跨午夜，星期几归属于开始那天：`FRIDAY 22:00-06:00` 覆盖周五 22:00 到周六 06:00，周六 22:00 不在窗口内。
- `start` 等于 `end` 拒绝保存；用空时段表达全天候。
- 绑定没有任何时段时，该通知组始终接收。
- 时区保存 IANA 名称，空值按 `Asia/Shanghai` 处理；判断时先把当前时刻转换到该时区。

时间语义参考 moon 的 `magicbox/timer`（星期范围、时刻范围、跨边界判断），实现放在 `app/alioth/pkg/oncall`，按本仓库的错误码和测试规范落地。

## 路由

评测回写 FIRING 或恢复时，`alioth` 把命中的订阅和 OnCall 合并成一组通知组后生成任务：

1. 标签、告警级别、数据源、数据源级别条件之间是 AND；空条件表示不限制。标签值为完整匹配正则；数据源级别取事件数据源自身的 `level_uid`。
2. 订阅和 OnCall 停用、通知组停用、渠道或模板停用时不生成任务。
3. 对 OnCall 的每个绑定，按 OnCall 时区取当前时刻，判断是否落在任一时段内；没有时段视为命中。FIRING 和恢复各自按事件发生时的时刻判断。
4. 订阅命中的通知组先处理；两套路由接收方重叠时，同一邮箱/手机号只保留先处理的一条任务，同一 Webhook 渠道只安排一次。`(event_uid, phase, group_uid, channel_uid)` 唯一约束作为最后一道兜底。
5. FIRING 按 `fingerprint` 查最近一条 PENDING/SENDING/SENT 的 FIRING 任务；命中来源的最大 `repeat_interval_seconds` 仍覆盖当前时刻时，整条事件不生成任务。
6. 恢复通知不受间隔限制，但同一指纹没有仍在投递链路上的 FIRING 任务时不补发；旧任务没有指纹时按 `event_uid` 兜底。

## 接口

| 方法 | 路径 |
| --- | --- |
| 创建 | `POST /v1/notification-oncall` |
| 更新 | `PUT /v1/notification-oncall/{uid}` |
| 启停 | `PUT /v1/notification-oncall/{uid}/status` |
| 删除 | `DELETE /v1/notification-oncall/{uid}` |
| 详情 | `GET /v1/notification-oncall/{uid}` |
| 列表 | `GET /v1/notification-oncalls` |

权限沿用 `notification:read`、`notification:write`、`notification:delete`。

OnCall 详情包含 `groups` 数组，每项带 `group_uid`、只读的 `group_name` 和 `windows` 数组。创建和更新请求只需要 `group_uid` 与 `windows`。

## Web

「通知 → OnCall」页面：

- 列表展示名称、订阅条件（含数据源级别）、通知组与时段摘要、状态和操作。
- 表单选择名称、标签、告警级别、数据源、数据源级别、时区、通知组和重复通知间隔；数据源、数据源级别和通知组使用支持搜索与全选的下拉框。
- 先选择通知组，再为每个选中的通知组展开时段编辑器；时段编辑器支持勾选星期几、填写起止时间、增删时段，不添加时段时提示“全天候接收”。
- 弹窗头部标题和底部确认按钮固定，只有中间内容区域滚动。
- 详情、启停、删除与告警订阅页面保持一致；告警订阅页使用同一套下拉框、数据源级别、重复间隔和固定头尾弹窗。

## 验收标准

- [ ] OnCall 能绑定多个通知组并保存，重复绑定同一通知组被拒绝或合并。
- [ ] 标签、告警级别、数据源、数据源级别条件按 AND 匹配；任一不满足时不发送。
- [ ] 通知组配置 `09:00-18:00` 时，18:00 整不接收，09:00 整接收。
- [ ] 跨午夜时段 `22:00-06:00` 在次日 02:00 接收，边界 06:00 不接收。
- [ ] 空时段的通知组全天候接收；停用 OnCall 或通知组后不再产生任务。
- [ ] 同一事件同时命中订阅和 OnCall、或两个通知组包含同一邮箱/手机号时，只生成一条包含该接收方的任务。
- [ ] 同一 Webhook 渠道被两个命中通知组绑定时只安排一条任务。
- [ ] 同一告警指纹在命中来源的最大重复间隔内不再生成 FIRING 任务；不同指纹不受影响。
- [ ] 重复间隔为 0 时不限制，最大允许 720 分钟；非法值在页面和服务端都被拒绝。
- [ ] 恢复通知不受重复间隔限制；从未成功生成 FIRING 任务或 FIRING 已失败的告警不补发恢复通知。
- [ ] 时区按 OnCall 配置生效，非法时区在保存时被拒绝。
- [ ] 数据源、数据源级别和通知组下拉框支持搜索、全选和清空。
- [ ] OnCall 弹窗头部和底部固定，内容区可滚动；暗黑模式下原生时间控件可见。
- [ ] 页面的增删改查、启停、时段编辑和详情可用；TypeScript 构建通过。
- [ ] `go test ./...`、`go vet`、`buf lint`、Web typecheck/lint/build 通过。
