# M10 Loki LogQL 告警闭环

M10 在 M3 指标评测、M4 通知投递与现有通用告警处置上增加 Loki 专属告警能力。Prometheus / VictoriaMetrics 的 PromQL/MetricsQL 路径不变；Elasticsearch、VictoriaLogs 和 Jaeger 仍不参与规则评测。

> M11 之后 VictoriaLogs 已加入数值型评测，见 [m11-vlogs-alerts.md](m11-vlogs-alerts.md)；M12 之后 Elasticsearch 也已加入，见 [m12-es-alerts.md](m12-es-alerts.md)。本文保留 M10 当时的验收范围，其中「VictoriaLogs / Elasticsearch 未实现」的表述不再是当前状态。

## 目标与依赖

操作者可以在 Loki 查询页和告警规则表单中编写 LogQL，使用 Loki 数据源验证查询；规则由 Megrez 定时评测，FIRING/恢复进入通用事件存储与通知订阅路由，接收人可查看告警、收到通知并处置。

依赖：M3、M4。

## 查询与规则约束

- `LOGS_LOKI` 独占 LogQL 编辑语言 ID、语法高亮和 Loki 标签/标签值补全；补全调用 `/loki/api/v1/labels` 与 `/loki/api/v1/label/{name}/values`。不得使用 PromQL language definition 或 Prometheus `/api/v1/*` 查询路径。
- 普通日志查询可返回 `streams`，继续在 Loki 数据源页展示日志行。
- Loki 告警表达式必须是返回数值序列的 LogQL 指标查询，例如 `sum by (app) (count_over_time({app="api"} |= "error" [5m]))`。规则级别条件将采样值与阈值比较；可设置持续时间。
- 原始日志 selector 返回 `resultType=streams`，不具备数值阈值语义。告警评测严格要求 Loki matrix 响应；stream、无效 JSON、上游失败不得回写空报告，不得使既有 FIRING 事件错误恢复。
- datasource matcher 的 driver 必须为 `LOGS_LOKI` 且数据源 ACTIVE。不得因为 labels 相同而把规则扩展到 Prometheus、VictoriaLogs 或 Elasticsearch。

## 评测和事件

1. 规则继续由 Alioth 调度，Megrez 按 `(rule_uid, datasource_uid)` 分布式加锁。
2. Megrez 使用 Loki `/loki/api/v1/query_range`，查询开始/结束使用 Unix 纳秒，`step` 使用 duration（例如 `15s`）；即时模板查询走 Loki `/loki/api/v1/query`。密钥沿用内部规则快照，不写入日志。
3. numeric matrix 解码为现有 `Sample`，复用 `FOR` / `MAX` / `MIN`、阈值、告警模板、fingerprint、锁续期与幂等回写。
4. 命中和恢复写入通用 `AlertEvent`；通知订阅按事件标签、级别、数据源筛选并生成 `NotificationTask`。
5. Loki 数值告警事件详情图表查询相同 LogQL 表达式的 Loki matrix，并只显示与事件标签匹配的序列。实时/历史查询与介入、抑制、手工恢复复用现有接口。

## 通知接收闭环

验收环境须有可用 Loki、Alioth、Megrez、Merak、PostgreSQL、Redis 和 Web。按顺序配置：

1. 在 Namespace 创建 `LOGS / LOGS_LOKI` 数据源，验证 `/ready` 连通并保存凭据。
2. 在 Loki 数据源页面运行普通 LogQL 日志查询；标签/标签值补全和日志结果来自 Loki。
3. 创建告警字典级别与规则组；创建 `LOGS_LOKI` 规则，设置 numeric LogQL 表达式、数据源 matcher、阈值、持续时间、摘要和描述，并启用规则。
4. 配置通知渠道、模板、通知组和 ACTIVE `NotificationSubscription`；subscription 指向通知组并匹配该规则使用的 datasource/级别/标签。订阅只决定接收路由，不绑定 Prometheus 语义。
5. 向 Loki 写入符合表达式的日志。至少一个评测周期后，实时告警出现 FIRING，样本标签、数据源标签、样本值和规则摘要可见；命中订阅后产生并投递通知任务，接收人可收到邮件或 Webhook。
6. 在实时告警详情检查 LogQL 表达式、数值曲线和原始响应；执行介入、抑制与手工恢复，并在历史告警中确认状态。
7. 停止产生日志，确认后续成功的 Loki 数值评测将事件自动恢复并按订阅发送恢复通知。

## 验收标准

- [ ] Loki 查询面板和规则表单使用 `logql` Monaco 语言；可完成标签键和值。组件不从 Prometheus 补全接口读取 Loki 标签，也不把 LogQL 作为 PromQL 高亮。
- [ ] Prometheus 规则仍请求 `/api/v1/query_range`；Loki 规则只请求 `/loki/api/v1/query_range`，保留 endpoint 路径前缀、认证、纳秒时间戳和 duration 步长。
- [ ] Loki numeric matrix 生成数值 samples/hits；无数据的有效 matrix 能恢复旧事件。Loki `streams` 响应、上游错误和畸形响应不会当作空矩阵回写或恢复事件。
- [ ] 只评测匹配的 ACTIVE `LOGS_LOKI` 数据源；同名标签的 Prometheus/其他日志源不被 Loki 规则查询。
- [ ] 告警命中后写入通用 FIRING 事件、匹配订阅、生成通知任务并由 Merak 投递；重复评测保持 fingerprint 和任务幂等。
- [ ] Loki 告警的详情图表可查询并按事件标签过滤序列；实时/历史列表、通知接收、介入、抑制、手工恢复和自动恢复可用。
- [ ] Loki 告警恢复会生成恢复通知；Elasticsearch、VictoriaLogs、Jaeger 仍按未实现驱动处理，不回退到 Prometheus。
- [ ] 相关 Go 测试、Web typecheck/build/lint 通过；数据库集成测试使用现有 `ALIOTH_TEST_DSN` 测试环境执行。

## 推荐验证命令

```bash
go test -C app/megrez ./internal/biz/eval ./internal/data
go test -C app/alioth ./internal/biz/alertevent ./internal/data ./internal/service
cd web && npm run typecheck && npm run lint && npm run build
```
