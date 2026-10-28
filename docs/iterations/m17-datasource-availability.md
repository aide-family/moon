# M17 数据源可用性告警

M17 在 M3 的规则评测回写与失败批次、M4 的通知任务与消息日志之上，为数据源增加可用性告警。可用性不来自额外的探测任务，而是从 `megrez` 已有的规则评测请求结果推导；事件沿用 `AlertEvent`、实时/历史告警、介入、抑制、手工恢复与历史分片归档，通知则直接投递给数据源绑定的通知组，不经过 `NotificationSubscription` 或 `NotificationOnCall`。

## 目标与依赖

操作者可以在数据源的新建/编辑对话框里为单个数据源开启可用性告警，设置统计窗口、失败次数阈值并直接绑定通知组。窗口内 `megrez` 对该数据源的规则评测请求失败达到阈值时，`alioth` 生成一条数据源级别的 FIRING 事件；事件出现在实时/历史告警中并可处置，绑定通知组收到 FIRING 与恢复通知。数据源恢复后事件自动恢复、归档到历史分片；没有绑定通知组时事件仍然记录和归档，只是不发通知。

依赖：M3（规则评测回写、`alert_eval_batches` 与 `AlertEvent`）、M4（`NotificationTask`、`merak` 投递与 `MessageLog`）。

## 数据模型与校验

`Datasource` 新增以下字段，随数据源新建/编辑一起保存：

| 字段 | 存储 | 默认 | 说明 |
| --- | --- | --- | --- |
| `availability_alert_enabled` | bool | `false` | 是否启用可用性告警 |
| `failure_window_minutes` | int | `5` | 统计窗口，单位分钟；取值 0..1440，启用时至少 1 |
| `failure_threshold` | int | `3` | 窗口内失败次数阈值；取值 0..1000，启用时至少 1 |
| `notification_group_uids` | jsonb 数组 | `[]` | 直接绑定的通知组 UID，不经过订阅或 OnCall |

- 窗口与阈值的范围始终校验为 0..1440 与 0..1000；启用可用性告警时业务层再要求窗口不小于 1、阈值不小于 1。关闭时允许 0，保存时把 0 归一为默认值 `5` 与 `3`，不产生告警。
- 超出 0..1440 / 0..1000 的输入一律返回客户端参数错误。
- 通知组绑定按当前 Namespace 解析，编辑对话框只列出本 Namespace 的通知组；列表和详情回填已绑定通知组的名称用于展示。

`AlertEvalBatch` 增加 `failed`（bool）与 `failure_message`（string，默认空）。`(rule_uid, datasource_uid, eval_slot)` 唯一约束不变：同一批次重复回写只保留一条记录，不重复计数。

可用性事件与其他事件共用 `AlertEvent`。当前 FIRING 的 `(namespace_uid, fingerprint)` 唯一约束照样生效；可用性事件的 fingerprint 由数据源 UID 稳定推导，因此同一数据源最多一条未恢复的可用性事件。

## 评测失败分类与上报

- 可用性失败只包括数据源请求本身的上游或服务端错误：连接失败、超时、HTTP 状态码不低于 300、响应不可读或无法解析。判定发生在 `megrez` 向数据源发出查询之后，与具体驱动无关。
- 表达式或请求本身不合法（`merr.Invalid`，例如规则表达式语法错误、参数不合法）是客户端错误，不是可用性失败：`failed` 保持 false，继续按原有评测错误与重试语义处理，也不进入失败计数。
- `megrez` 对可用性失败的批次回写 `ReportEval`，置 `failed=true` 并带 `failure_message`，`hits` 留空；`failure_message` 取最近一次失败原因，进入事件描述并用于排查。成功批次维持 `failed=false`。
- 未实现的驱动（`TRACE_JAEGER`）回写 `skipped=true`。跳过既不是失败也不是成功，不参与可用性计数；启用了可用性告警的 TRACE 数据源在驱动实现前不会因评测产生可用性事件。
- 失败批次与成功批次一样被记为「该 `(rule, datasource, eval_slot)` 已回写」：`AllBatchesRecorded` 满足后清除规则的 `scheduled_slot`，失败的任务不会被调度器反复领取。这里清除的是本次评测的调度标记，不是失败证据。

## 计数窗口与去重

- 计数按失败批次：同一数据源上所有规则的可用性失败都计入该数据源；每条 `(rule_uid, datasource_uid, eval_slot)` 只算一次，重复回写不增加次数。同一评测周期多条规则同时失败会更快达到阈值。
- 窗口按失败批次的写入时间（`created_at`，即失败回写到达 `alioth` 的时刻）向前回溯 `failure_window_minutes` 分钟，不用 `eval_slot`；窗口内失败数达到 `failure_threshold` 时生成 FIRING 事件。
- 同一数据源最多一条未恢复的可用性事件：再次达到阈值时，若事件已存在，只更新 `sample_value`（失败次数）、`summary` 和 `labels`，不新建事件、不重复发送 FIRING 通知。
- 计数与「事件是否存在」都以 PostgreSQL 为准，不放 Redis 或进程内缓存；多副本由唯一约束和事务保证不重复。

## 事件指纹与格式

可用性事件与普通规则事件共用 `AlertEvent` 的表、字段和展示格式：

| 字段 | 取值 |
| --- | --- |
| `status` | `FIRING`；恢复后为 `RECOVERED`，人工恢复为 `RECOVERED_BY_MANUAL` |
| `fingerprint` | 由数据源 UID 稳定推导，数据源不变则指纹不变 |
| `rule_uid` | 空。可用性事件没有关联规则 |
| `datasource_uid` | 失败的数据源 |
| `level_uid` | 空。可用性事件没有告警级别，数据源级别名称只出现在 `labels.datasource_level` |
| `labels` | 数据源名称、数据源自身的 `labels`、数据源级别名称（`datasource_level`） |
| `summary` | 例如「数据源 X 在 n 分钟内请求失败 c 次（阈值 m）」 |
| `description` | 最近一次失败原因（`failure_message`） |
| `sample_value` | 窗口内的失败次数 |
| `starts_at` | 事件首次触发时间 |
| `eval_slot` | 触发该事件的那次失败评测批次 |

- 实时/历史列表的展示、详情沿用普通事件；规则名称为空，详情不展示规则表达式。`level_uid` 为空，按告警级别筛选时不会命中可用性事件；数据源级别只用于标签匹配与展示。
- 介入、抑制、手工恢复沿用实时告警的现有流程；人工恢复把状态置为 `RECOVERED_BY_MANUAL` 并归档。

## 通知路由

- FIRING 通知直接发给数据源 `notification_group_uids` 绑定的通知组，恢复通知也走同一绑定关系；两者都不参与 `NotificationSubscription` 与 `NotificationOnCall` 的标签、级别、数据源匹配，也不受订阅或 OnCall 的重复通知间隔限制。
- 没有绑定通知组时事件照常记录、可在告警页面处置并归档，但不生成 `NotificationTask`，也没有消息日志。
- 通知任务沿用 `NotificationTask`、`merak` 投递、`MessageLog` 和现有任务去重约束；任务带事件 fingerprint 与 FIRING / 恢复阶段。
- 重复的失败回写只更新事件，不生成第二条 FIRING 通知。恢复通知沿用普通任务链路的条件：只有该事件此前实际生成过 FIRING 任务时，才会生成 `RECOVERED` 任务并由 `merak` 投递；没有绑定通知组或 FIRING 任务没有建立时，恢复只更新事件状态并归档，不发恢复通知。

## 恢复与关闭

- 数据源下一次规则评测成功（`failed=false` 的批次回写）后，`alioth` 生成恢复事件：状态 `RECOVERED`，`ends_at` 为恢复时刻，并归档到历史分片；若该事件此前存在 FIRING 任务，再按绑定的通知组生成恢复通知。
- 数据源停用（`DISABLED`）或删除时，仍处于 FIRING 的可用性事件直接恢复，不再等待后续评测；删除后数据源不再可见，历史事件仍按归档保留。
- 在数据源编辑里只把 `availability_alert_enabled` 关掉时，仍处于 FIRING 的可用性事件同样立即恢复；普通规则事件不受影响。这是与停用（`DISABLED`）、删除数据源并列的第三条关闭路径。
- 数据源没有规则，或规则全部停用/删除后已没有评测请求时，不会产生新的可用性失败；已经 FIRING 的事件也不会自动恢复，只能等待重新出现成功评测、停用/删除数据源或人工恢复。
- 恢复后事件进入历史列表；实时列表只保留仍 FIRING 且未被抑制的事件，与普通事件一致。

## 历史归档

- 可用性事件恢复后走与普通事件相同的历史归档：在事务提交后写入租户历史库周表，删除主库行，`alert_history_events` 记录事件 UID 与分片位置。
- 归档失败时保留主库行，历史列表合并主库兜底；升级前遗留的已恢复事件仍由每小时迁移处理。详情按 UID 补名称时数据源名称正常显示，规则名称为空。

## 边界情况

- 无规则的数据源：没有评测请求，不会产生可用性失败；保存可用性配置不报错，也不产生事件。
- 没有绑定通知组：事件记录、可处置、会归档，但不发通知。
- 停用或删除数据源：恢复当前 FIRING 事件；停用期间不再评测，不会再产生新的失败计数。
- 未实现的 `TRACE_JAEGER` 驱动：`skipped=true` 不计失败；可以保存可用性配置，但驱动实现前不会因评测产生事件。
- 重复或迟到回写：同一 `(rule_uid, datasource_uid, eval_slot)` 只记一次；重复上报不增加计数、不新建事件、不重复发 FIRING 通知。`scheduled_slot` 已清除后再回写同一批次仍按重复处理。
- 失败原因变化：新失败只把描述更新为最近一次 `failure_message`，不新建事件。
- 多条规则同时失败：每个失败批次各计一次，因此同一窗口内更容易达到阈值。

## Web

- 数据源新建/编辑对话框增加「可用性告警」区域：启用开关、统计窗口（分钟）、失败次数阈值、通知组多选。输入提示 1..1440 / 1..1000；关闭时允许 0 并在保存时归一为默认 `5` 与 `3`，超出 0..1440 / 0..1000 的输入拦下。
- 通知组多选只列出当前 Namespace 的通知组；保存后重新打开数据源，窗口、阈值与绑定关系保持不变。
- 数据源列表/详情展示是否启用可用性告警。
- 实时/历史告警中可用性事件使用现有列（摘要、级别、数据源、样本值、开始时间）；`level_uid` 为空，级别列不显示告警级别；详情展示描述（最近失败原因），规则列与表达式为空。介入、抑制、手工恢复沿用现有交互。
- 中文与英文语言包同时补齐。

## 验收步骤

验收环境须有可用的数据源（Prometheus / VictoriaMetrics、Loki、VictoriaLogs 或 Elasticsearch）、至少一条会请求该数据源的已启用规则、通知渠道/模板/通知组，以及 Alioth、Megrez、Merak、PostgreSQL、Redis 和 Web。按顺序配置：

1. 新建数据源时开启可用性告警，窗口填 `5`、阈值填 `3`，绑定通知组；保存后重新打开，配置与绑定关系保持不变。启用时窗口填 `0` 或 `1441`、阈值填 `0` 或 `1001` 保存失败并提示范围；关闭开关后窗口与阈值填 `0` 可以保存并分别归一为 `5` 与 `3`，`1441` / `1001` 仍然拒绝。
2. 让数据源不可达（停掉上游服务，或把地址改为不可达端口），保持规则启用。等到若干评测周期后，实时告警出现一条 FIRING 可用性事件：摘要含窗口与失败次数，描述为最近失败原因，样本值为失败次数，规则列与级别列为空。
3. 继续制造失败：事件只有一条，样本值、摘要和标签随失败次数更新，不出现第二条事件，也不重复产生 FIRING 通知。
4. 查看通知任务与消息日志：绑定的通知组各生成一条 FIRING 任务并投递成功。解绑全部通知组后继续制造失败，事件仍然更新，但没有新任务和消息日志。
5. 重新绑定通知组并恢复数据源。下一次成功评测后事件自动恢复：实时列表移除、历史列表出现 `RECOVERED` 记录，并按绑定组生成恢复通知任务与消息日志（该事件此前已有 FIRING 任务）。再解绑通知组、重新制造一轮失败使事件重新 FIRING（此时没有 FIRING 任务），恢复后确认只更新事件状态、不发恢复通知。
6. 再次触发到 FIRING，分别在实时告警执行介入、抑制和手工恢复；状态、历史列表与详情一致，手工恢复为 `RECOVERED_BY_MANUAL`。
7. 让事件处于 FIRING 后先只关闭可用性告警开关，事件立即恢复且普通规则事件不受影响；重新开启并再次触发后停用数据源，事件恢复；再启用并触发后删除数据源，事件同样恢复并归档。
8. 把一条规则表达式改成非法语法，确认回写 `failed=false`、没有新增可用性失败批次、不产生可用性事件；表达式修正后原有评测行为不变。
9. 对 `TRACE_JAEGER` 数据源开启可用性告警并挂一条 TRACE 规则，确认进入跳过路径、`failed=false`，不产生可用性事件。
10. 重复回写同一 `(rule_uid, datasource_uid, eval_slot)`（模拟重复上报），确认失败次数只增加一次，事件与通知不重复。
11. 在历史告警明细按数据源、状态与时间范围查询，恢复事件可定位详情；主库行已删除、历史分片存在。

## 验收标准

- [ ] 数据源新建/编辑对话框可以开启可用性告警，保存窗口、阈值与通知组绑定；范围始终限制在 0..1440 / 0..1000，启用时拒绝 0，关闭时 0 归一为 `5` / `3`，且不产生事件。
- [ ] 连接失败、超时、HTTP 状态码不低于 300、响应不可读或不可解析的回写带 `failed=true` 与 `failure_message`；`merr.Invalid` 的表达式或请求错误不带 `failed`，保持原有重试语义。
- [ ] 失败批次写入 `alert_eval_batches`；同一 `(rule_uid, datasource_uid, eval_slot)` 重复回写不重复计数；失败批次同样清除 `scheduled_slot`，不会被反复领取。
- [ ] 窗口内失败数达到阈值时生成一条 FIRING 可用性事件，fingerprint 稳定、`rule_uid` 与 `level_uid` 为空，labels（含 `datasource_level`）、summary、description、sample_value、starts_at 与 eval_slot 符合约定。
- [ ] 同一数据源最多一条未恢复事件；后续失败只更新数值、摘要和标签，不重复发 FIRING 通知。
- [ ] FIRING 与恢复通知只发给数据源绑定的通知组，不经过订阅或 OnCall；没有绑定通知组时事件仍记录与归档，但不发通知。
- [ ] 数据源下一次成功评测后事件恢复；关闭可用性告警开关、停用或删除数据源也会恢复事件；恢复通知走 `NotificationTask` → `merak` → `MessageLog`，且仅在该事件已有 FIRING 任务时生成。
- [ ] 恢复事件进入历史分片并可在历史列表/详情查看；实时列表只保留 FIRING 事件。
- [ ] 无规则数据源、未绑定通知组、未实现的 TRACE 驱动、重复或迟到回写都不产生错误事件或重复通知。
- [ ] 实时告警的介入、抑制、手工恢复对可用性事件可用，结果与普通事件一致。
- [ ] 相关 Go 测试、Web typecheck/lint/build 通过；数据库集成测试使用现有 `ALIOTH_TEST_DSN` 测试环境执行。

## 推荐验证命令

```bash
go test -C app/megrez ./internal/biz/eval ./internal/data
go test -C app/alioth ./internal/biz/datasource ./internal/biz/alertevent ./internal/data ./internal/service
cd web && npm run typecheck && npm run lint && npm run build
```
