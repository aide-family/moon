# M11 VictoriaLogs LogsQL 告警闭环

M11 在 M3 指标评测、M4 通知投递与 M10 通用告警处置基础上增加 VictoriaLogs 专属告警能力。Prometheus / VictoriaMetrics 的 PromQL/MetricsQL 路径、Loki 的 LogQL 路径保持不变；Elasticsearch 与 Jaeger 仍不参与规则评测。

## 目标与依赖

操作者可以在 VictoriaLogs 查询页和告警规则表单中编写 LogsQL，使用 VictoriaLogs 数据源验证查询；规则由 Megrez 定时评测，FIRING/恢复进入通用事件存储与通知订阅路由，接收人可查看告警、收到通知并处置。

依赖：M3、M4（M10 的事件与通知闭环复用同一套公共流程）。

## 查询与规则约束

- `LOGS_VLOGS` 独占 `logsql` 编辑语言、语法高亮和 VictoriaLogs 字段/字段值补全；字段名调用 `/select/logsql/field_names`，字段值调用 `/select/logsql/field_values?field=<name>`。不得使用 PromQL / LogQL language definition 或 Prometheus / Loki 查询路径。
- 普通日志查询继续走 `/select/logsql/query`，返回 JSON Lines，在数据源页展示日志行和原文。
- VictoriaLogs 告警表达式必须是带 `| stats ...` 的 LogsQL 查询。`stats by (...)` 的分组字段成为序列标签，stats 输出字段成为 `__name__`，例如：
  ```
  level:error AND _msg:"request timeout" | stats by (app, level) count() as value
  ```
- 时间桶由 Megrez 通过 `stats_query_range` 的 `step` 参数切分，表达式里不要写 `stats by (_time:<步长>)`，也不要加 `_time:5m` 这类相对时间过滤；它们会和引擎窗口冲突。
- 原始日志查询没有 stats 语义。告警评测严格要求 Prometheus matrix 且每个采样点都是有限数值：原始日志、非 matrix 响应、非数值 stats（例如 `values(level)`）、无效 JSON 和上游失败不得回写空报告，不得使既有 FIRING 事件错误恢复。有效的空 matrix 才表示本次没有采样，用于恢复旧事件。
- datasource matcher 的 driver 必须为 `LOGS_VLOGS` 且数据源 ACTIVE。不得因为 labels 相同而把规则扩展到 Prometheus、Loki 或 Elasticsearch。
- `stats_query_range` 的桶按 step 对齐，结束时间落在桶中间时最后一个桶只统计部分日志。引擎把结束时间向下对齐到桶边界；默认步长 15 秒，FOR 短于 15 秒时改用与 FOR 相同的桶宽（例如 10 秒），避免短规则被放大成两个 15 秒桶。规则使用 FOR 时再多查一个桶，让桶跨度覆盖配置的持续秒数。
- FOR 要求时间桶连续。VictoriaLogs 不会为没有命中的桶返回 0，缺失桶会让 FOR 不满足，不能只凭首末两桶的跨度判定持续命中。

## 评测和事件

1. 规则继续由 Alioth 调度，Megrez 按 `(rule_uid, datasource_uid)` 分布式加锁。
2. Megrez 使用 VictoriaLogs `/select/logsql/stats_query_range`，`query` 为规则表达式，`start` / `end` 为 Unix 秒，`step` 为 `<秒>s`；即时模板查询使用 `/select/logsql/stats_query`，范围取查询时刻之前的一个步长。密钥沿用内部规则快照，不写入日志。
3. Prometheus matrix 解码为现有 `Sample`，并把 step 记录到 `Sample.Step`。复用 `FOR` / `MAX` / `MIN`、阈值、告警模板、fingerprint、锁续期与幂等回写；FOR 额外检查相邻采样点是否缺桶。
4. 命中和恢复写入通用 `AlertEvent`；通知订阅按事件标签、级别、数据源筛选并生成 `NotificationTask`。
5. VictoriaLogs 告警事件详情图表调用同一个 `stats_query_range`，按事件标签收敛 matrix 序列；实时/历史查询与介入、抑制、手工恢复复用现有接口。

## 通知接收闭环

验收环境须有可用 VictoriaLogs、Alioth、Megrez、Merak、PostgreSQL、Redis 和 Web。按顺序配置：

1. 在 Namespace 创建 `LOGS / LOGS_VLOGS` 数据源，地址指向 VictoriaLogs 的 `/health` 所在根地址并保存凭据。
2. 在 VictoriaLogs 数据源页面运行普通 LogsQL 日志查询；字段名/字段值补全和日志结果来自 VictoriaLogs。
3. 创建告警字典级别与规则组；创建 `LOGS_VLOGS` 规则，设置 stats 表达式（例如 `level:error | stats by (app, level) count() as value`）、数据源 matcher、阈值、持续时间、摘要和描述，并启用规则。
4. 配置通知渠道、模板、通知组和 ACTIVE `NotificationSubscription`；subscription 指向通知组并匹配该规则使用的 datasource/级别/标签。
5. 向 VictoriaLogs 写入符合表达式的日志。至少一个评测周期后，实时告警出现 FIRING，样本标签、数据源标签、样本值和规则摘要可见；命中订阅后产生并投递通知任务，接收人可收到邮件或 Webhook。
6. 在实时告警详情检查 LogsQL 表达式、数值曲线和原始响应；执行介入、抑制与手工恢复，并在历史告警中确认状态。
7. 停止产生日志，确认后续成功的 VictoriaLogs 数值评测将事件自动恢复并按订阅发送恢复通知。

本地可用 `make vlogs-demo` 每 5 秒写入 `api/error` 与 `web/info` 示例日志，配套表达式见 [README](../../README.md#本地全栈)。

## 验收标准

- [ ] VictoriaLogs 查询面板和规则表单使用 `logsql` Monaco 语言；可完成字段名和字段值。组件不读取 Prometheus / Loki 补全接口，也不把 LogsQL 作为 PromQL 或 LogQL 高亮。
- [ ] VictoriaLogs 规则只请求 `/select/logsql/stats_query_range`，保留 endpoint 路径前缀、认证、Unix 秒起止和 duration 步长；即时模板查询只请求 `/select/logsql/stats_query`；不请求 Prometheus 或 Loki endpoint。
- [ ] 有效的 stats matrix 生成数值 samples/hits；空 matrix 能恢复旧事件。原始日志、非 matrix 响应、非数值 stats 点、无效 JSON 或畸形响应不会当作空结果回写或恢复事件。
- [ ] 只评测匹配的 ACTIVE `LOGS_VLOGS` 数据源；同名标签的 Prometheus / Loki / Elasticsearch 数据源不被 VictoriaLogs 规则查询。
- [ ] FOR 规则在时间桶上的跨度判定覆盖配置的持续秒数；缺失桶不会被视为持续命中；MAX / MIN 规则仍按完整桶的值比较。
- [ ] 告警命中后写入通用 FIRING 事件、匹配订阅、生成通知任务并由 Merak 投递；重复评测保持 fingerprint 和任务幂等。
- [ ] VictoriaLogs 告警的详情图表可查询并按事件标签过滤序列；实时/历史列表、通知接收、介入、抑制、手工恢复和自动恢复可用。
- [ ] VictoriaLogs 告警恢复会生成恢复通知；Elasticsearch 与 Jaeger 仍按未实现驱动处理，不回退到 Prometheus。
- [ ] 相关 Go 测试、Web typecheck/build/lint 通过；数据库集成测试使用现有 `ALIOTH_TEST_DSN` 测试环境执行。

## 推荐验证命令

```bash
go test -C app/megrez ./internal/data ./internal/biz/eval
go test -C app/alioth ./internal/data ./internal/biz/alertevent ./internal/service
cd web && npm run typecheck && npm run lint && npm run build
```
