# M12 Elasticsearch 时间桶告警闭环

M12 在 M3 指标评测、M4 通知投递、M10 通用告警处置、M10/M11 的 Loki 与 VictoriaLogs 数值型告警基础上，
补上 Elasticsearch 的告警评测。Prometheus / VictoriaMetrics、Loki、VictoriaLogs 的查询路径保持不变；
`TRACE_JAEGER` 仍不参与规则评测。

## 目标与依赖

操作者可以在 Elasticsearch 数据源上创建告警规则，表达式沿用日志查询页面的 Query DSL 或检索语句；
规则由 Megrez 定时评测，FIRING/恢复进入通用事件存储与通知订阅路由，接收人可查看告警、收到通知并处置。

依赖：M3、M4（M10 的事件与通知闭环复用同一套公共流程）。

## 为什么需要独立的查询路径

告警判定吃的是「每个时间桶一个有限数值」。Prometheus / VictoriaMetrics 有 `query_range`，Loki 有数值型
LogQL，VictoriaLogs 有 `stats_query_range`，三者都直接返回 Prometheus matrix。Elasticsearch 的 `_search`
返回的是命中文档，没有等价接口，所以由 Megrez 侧把命中数变成采样值。

## 查询与规则约束

- 表达式兼容两种输入：完整 Query DSL 对象（以 `{` 开头，原样作为 `bool.must`）或普通检索语句
  （包装成 `query_string`）。空表达式在构造请求阶段就被拒绝，不会发出上游请求。
- 请求固定 `size: 0`、`track_total_hits: false`，只取聚合；聚合为时间字段上的 `date_histogram`，
  `fixed_interval` 取规则的 step，`min_doc_count: 0` 且 `extended_bounds` 覆盖整个窗口。
- 时间范围是半开区间 `[start, end)`，用 epoch 毫秒传给上游；`extended_bounds.max` 取 `end-1` 毫秒。
- 桶边界由 step 决定，Megrez 把结束时间向下对齐到桶边界（与 VictoriaLogs 相同）。FOR 规则多查一个桶，
  让采样点跨度覆盖配置的持续秒数；默认步长 15 秒，持续秒数短于 15 秒时改用与 FOR 相同的桶宽。
- 空桶返回 0 而不是缺席，FOR 因而不需要额外的「缺桶」容忍；这与 Loki `count_over_time` 的语义一致。
- 索引模式和时间字段来自数据源，随内部规则快照下发。索引模式只允许索引名字符与通配符，避免把额外
  路径拼到上游；空索引模式表示搜索全部索引，空时间字段表示 `@timestamp`。
- 响应里没有聚合结果（索引不存在、查询语法错误、上游 4xx、非 JSON）一律按评测失败处理且不回写，
  不会把故障误判成「零命中」而恢复既有 FIRING 事件。有效的空桶列表表示本次没有命中，用于恢复。

## 评测和事件

1. 规则继续由 Alioth 调度，Megrez 按 `(rule_uid, datasource_uid)` 分布式加锁。
2. Megrez 按 `LOGS_ELASTICSEARCH` driver 调用 `POST /<index>/_search`，认证沿用内部规则快照里的
   Basic / Bearer，不写入日志。
3. `date_histogram` 的每个桶解码成一个 `bo.Point`，`doc_count` 同时作为数值和原始文本；整个响应是
   单个序列，`Sample.Step` 记录桶宽。复用 `FOR` / `MAX` / `MIN`、阈值、告警模板、fingerprint、
   锁续期与幂等回写。
4. 命中和恢复写入通用 `AlertEvent`；通知订阅按事件标签、级别、数据源筛选并生成 `NotificationTask`。
5. 事件详情图表用同一套时间桶计数，并把结果编码成 Prometheus matrix，前端复用指标图表。

## 通知接收闭环

验收环境须有可用 Elasticsearch、Alioth、Megrez、Merak、PostgreSQL、Redis 和 Web。按顺序配置：

1. 在 Namespace 创建 `LOGS / LOGS_ELASTICSEARCH` 数据源，填地址、索引模式和时间字段并保存凭据。
   自签 HTTPS 集群需要把 CA 注入 Pod 信任库（见 [deploy/k8s/README](../../deploy/k8s/README.md#私有-caelasticsearch)），
   且端点要写成证书 SAN 里的主机名。
2. 在数据源页面运行一次日志查询，确认索引模式和时间字段能取到文档。
3. 创建告警字典级别与规则组；创建 `LOGS_ELASTICSEARCH` 规则，填写表达式、阈值、持续时间、
   摘要和描述，并启用规则。
4. 配置通知渠道、模板、通知组和 ACTIVE `NotificationSubscription`；subscription 指向通知组并匹配该
   规则使用的 datasource/级别/标签。
5. 向匹配索引写入符合条件的文档。至少一个评测周期后，实时告警出现 FIRING，样本值和规则摘要可见；
   命中订阅后产生并投递通知任务。
6. 在实时告警详情查看表达式与数值曲线；执行介入、抑制与手工恢复，并在历史告警中确认状态。
7. 停止写入文档，确认后续成功的评测把事件自动恢复并按订阅发送恢复通知。

## 验收标准

- [ ] `LOGS_ELASTICSEARCH` 规则产出数值采样并参与 FIRING 判定；`numericDriver` 不再把它当未实现。
- [ ] 评测只请求 `POST /<index>/_search`，保留 endpoint 路径前缀、认证、epoch 毫秒起止和 step 步长；
      不请求 Prometheus / Loki / VictoriaLogs 的查询接口。
- [ ] 请求体只取聚合（`size: 0`），`date_histogram` 使用数据源的时间字段和规则步长，`min_doc_count: 0`。
- [ ] 空桶返回 0，使 FOR 能正确判定持续不命中；FOR 的采样点跨度覆盖配置的持续秒数。
- [ ] 索引不存在、查询语法错误、非 JSON 或缺少聚合结果的响应不会回写空报告，也不会错误恢复事件。
- [ ] 告警命中后写入通用 FIRING 事件、匹配订阅、生成通知任务并由 Merak 投递；重复评测保持幂等。
- [ ] 告警详情图表可查询 Elasticsearch 并按事件时间窗返回曲线；实时/历史列表、介入、抑制、手工恢复可用。
- [ ] 告警恢复会生成恢复通知；`TRACE_JAEGER` 仍按未实现驱动跳过，不回退到 Prometheus。
- [ ] 相关 Go 测试、Web typecheck/build/lint 通过。

## 推荐验证命令

```bash
go test -C pkg ./esagg/...
go test -C app/megrez ./internal/data ./internal/biz/eval
go test -C app/alioth ./internal/data ./internal/biz/alertevent ./internal/service
cd web && npm run typecheck && npm run lint && npm run build
```
