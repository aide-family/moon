# 集群内数据源与告警链路验证

2026-10-25 在 `moon` 集群实测。**五类数据源全部接进来并验证到"事件产生 → 推送 → 恢复"**，
Elasticsearch 需要先解决自签 TLS 才能接入（见下）。这里记录怎么验的、结果如何、
留在集群里的对象怎么清理。

## 数据源矩阵

| 数据源 | 驱动 | 端点（集群内） | 连通 | 查询 | 告警触发 | 告警恢复 |
| --- | --- | --- | --- | --- | --- | --- |
| 集群 Prometheus | `METRICS_PROMETHEUS` | `http://prometheus-server.monitoring.svc.cluster.local:80` | ✅ `up` | ✅ 指标名 / `up` | ✅ | ✅ |
| 集群 VictoriaMetrics | `METRICS_VICTORIA_METRICS` | `http://vmcluster-…-vmselect.monitoring.svc.cluster.local:8481/select/0/prometheus` | ✅ `up` | ✅ 1396 个指标名 | ✅ | ✅ |
| 集群 Loki | `LOGS_LOKI` | `http://loki.monitoring.svc.cluster.local:3100` | ✅ `up` | ✅ LogQL | ✅ | ✅ |
| 集群 VictoriaLogs | `LOGS_VLOGS` | `http://vlogs-cluster-…-vlselect.monitoring.svc.cluster.local:9471` | ✅ `up` | ✅ LogsQL | ✅ | ✅ |
| 集群 Elasticsearch | `LOGS_ELASTICSEARCH` | `https://es-es-http.es.svc:9200` | ✅ `green` | ✅ 普通检索 | ✅ | ✅ |
| （`TRACE_JAEGER`） | — | 集群里没有 Jaeger | — | — | — | — |

### Elasticsearch 的几个坑

**1. 自签证书需要注入信任库。** ECK 默认 TLS，证书由集群内 CA 签发，而数据源客户端用的是
Go 默认 `http.Client`（没有跳过校验的开关）。做法：把「系统 CA + ECK CA」合成一个 bundle，
放进 `alioth` 命名空间的 Secret `alioth-ca-bundle`，再挂到 Pod 的
`/etc/ssl/certs/ca-certificates.crt`（`alioth.yaml` 里已配好，Secret 不存在时该卷是可选的）。
合成命令：

```bash
kubectl -n alioth exec deploy/webhook-sink -- cat /etc/ssl/certs/ca-certificates.crt > sys-ca.crt
kubectl -n es get secret es-es-http-certs-public -o jsonpath='{.data.ca\.crt}' | base64 -d > es-ca.crt
cat sys-ca.crt es-ca.crt > ca-bundle.crt
kubectl -n alioth create secret generic alioth-ca-bundle --from-file=ca-certificates.crt=ca-bundle.crt
kubectl -n alioth rollout restart deploy/alioth
```

**2. 端点必须用证书里的主机名。** ECK 签发的证书 SAN 是 `es-es-http.es.svc`、`es-es-http.es`，
**不含** `.cluster.local`。写 `https://es-es-http.es.svc.cluster.local:9200` 会报
`x509: certificate is valid for … not es-es-http.es.svc.cluster.local`。

**3. Elasticsearch 参与告警评测需要 megrez 也能信任这个 CA。** 评测进程才是直连数据源的
那个，只给 `alioth` 挂 `alioth-ca-bundle` 不够：`megrez.yaml` 也挂了同一个卷，否则评测会报
`x509: certificate signed by unknown authority`，表现为规则一直被调度但从不产生事件。

**4. Elasticsearch 的告警是「时间桶计数」。** `LOGS_ELASTICSEARCH` 已纳入
`numericDriver()`（`app/megrez/internal/biz/eval/judge.go`），表达式就是日志检索语句或
Query DSL，引擎把它翻成 `date_histogram` 聚合，每个桶的 `doc_count` 作为一个采样点，
阈值比较的是**命中条数**。空桶返回 0，所以 MAX/FOR 的语义和其它驱动一致。

## 告警是怎么可控地触发和恢复的

不是等真实故障，而是给每个驱动注入可控样本，触发和恢复都能精确复现：

| 驱动 | 手段 | 触发 | 恢复 |
| --- | --- | --- | --- |
| Prometheus | 向 `prometheus-prometheus-pushgateway` 推 `smoke_alert_metric` | 推 `1` | 推 `0`（约 45s 后规则恢复） |
| Loki | `POST /loki/api/v1/push`，流标签 `job=alioth-smoke`，正文含 `SMOKE_LOKI_MARKER` | 写入后约 15s | 日志滑出 15s 评测窗口后自动恢复 |
| VictoriaLogs | `POST /insert/jsonline?_stream_fields=app`，正文含 `SMOKE_VLOG_MARKER` | 写入后约 15s | 同上 |
| VictoriaMetrics | 用集群里真实 down 的目标 `up{job="victorialogs"} == 0` | 一直 FIRING | 修好抓取配置后才会恢复 |
| Elasticsearch | `POST /alioth-smoke/_doc?refresh=true` 写入含 `SMOKE_ES_MARKER` 的文档 | 15s 内 | 文档滑出统计窗口后自动恢复 |

对应规则（都在 `集群内验证组`，级别 `P1-严重`）：

- `验证-Prometheus指标`：`smoke_alert_metric`，`MAX > 0`
- `验证-Loki日志`：`sum(count_over_time({job="alioth-smoke"} |= "SMOKE_LOKI_MARKER" [15s]))`，`MAX > 0`
- `验证-VictoriaLogs日志`：`"SMOKE_VLOG_MARKER" | stats count() as value`，`MAX > 0`
- `集群目标不可用`：`up == 0`，`FOR 30s`（真实规则）
- `ES-错误日志`：`LOGS_ELASTICSEARCH`，表达式 `level:error`，`MAX > 0`

日志规则的表达式都要求返回 Prometheus matrix：Loki 必须是数值型 LogQL（原始日志流响应会被
当成查询错误，不是"没命中"），VictoriaLogs 必须是 `| stats …` 查询。

## 通知

一条订阅 `全部告警→验证组`（`labels: {}`，不限级别和数据源）把事件路由到 `验证通知组`，
组里绑了两个渠道：

- `验证 Webhook` → `http://webhook-sink.alioth.svc.cluster.local/`（Vector `http_server`，
  `kubectl -n alioth logs deploy/webhook-sink` 看实际报文；只接受根路径，写 `/hook` 会 404）
- `运维邮箱` → `smtp.example.com:25`，收件人 `ops@example.com`

实测每个驱动的 FIRING / RECOVERED 都在两个渠道各投递一次，`notification_tasks` 全部 `SENT`
（9 FIRING + 7 RECOVERED + 1 TEST）。规则**只在事件创建和恢复时投递**，持续 FIRING 不重复推送。

## 遗留问题

监控里 `victorialogs` 这个抓取目标一直是 `up = 0`：

```
job=victorialogs  instance=vlogs-victoria-logs-single-server.monitoring.svc.cluster.local:9428
```

集群跑的是集群版 VictoriaLogs（`vlogs-cluster-victoria-logs-cluster-vlselect`），单机版 Service
不存在。改抓取配置或删掉这条 job 之前，`集群目标不可用` 会一直保持 FIRING（不会重复通知）。

## 清理

```bash
kubectl -n alioth delete -f deploy/k8s/smoke/webhook-sink.yaml
```

平台的验证对象（5 个数据源、3 条验证规则、模板/渠道/通知组/订阅）按依赖顺序删：订阅 → 通知组
→ 渠道 → 模板 → 规则 → 规则组 → 数据源。`集群目标不可用` 和 `集群 VictoriaMetrics` 是长期要用的，
建议保留。
