# M18 探测规则（HTTP）

M18 在 M3 规则评测、M4 通知链路与 M17 事件口径之上增加「探测规则」。探测规则与查询规则共用告警规则入口、规则组、级别条件、`AlertEvent`、实时/历史告警与通知链路；区别只在观测方式：探测规则按周期对一个或多个**探测目标组**里的地址发 HTTP 请求，把结果变成同样格式的事件。

一期只做 HTTP，只从集群内探测；探测执行是 `megrez` 里新增的一个 worker，不是新进程。

## 目标与依赖

操作者可以：

1. 在「探测目标」页按探测类型管理目标组（组内多个地址、共用请求参数、带标签、可启停）；
2. 在「告警规则」页新建探测规则，用标签匹配一个或多个目标组，选择观测指标（`up` / 耗时 / 状态码 / 证书天数）与级别条件；
3. 命中条件时在实时告警里看到与查询规则同样格式的事件，并走同样的介入、抑制、恢复、归档与通知；
4. 在规则详情与目标组详情里看到最近探测记录。

依赖：M3（`ClaimRules` / `ReportEval` / `AlertEvent`）、M4（`NotificationTask`、`merak`、`MessageLog`）、M17（无规则事件的先例与事件口径）。

## 数据模型与校验

新增三张表：

| 表 | 关键字段 | 约束 |
| --- | --- | --- |
| `probe_target_groups` | `uid`、`namespace_uid`、`name`、`driver`（`PROBE_HTTP`）、`status`、`labels`、`executor_labels`、`request`（JSON）、`remark` | `(namespace_uid, name)` 在 `deleted_at IS NULL` 下唯一；名称 ≤64 |
| `probe_targets` | `uid`、`group_uid`、`spec`、`labels`、`status`、`sort` | `(group_uid, spec)` 在 `deleted_at IS NULL` 下唯一 |
| `probe_results` | `rule_uid`、`group_uid`、`target_uid`、`slot`、`ok`、`value`、`duration_ms`、`status_code`、`error_kind`、`error_message`、`dns_ms/tcp_ms/tls_ms/ttfb_ms`、`cert_days`、`resolved_ip`、`executor` | `(target_uid, slot)` 唯一；`(rule_uid, group_uid, target_uid, probed_at)` 普通索引 |

`AlertRule` 新增：

| 字段 | 说明 |
| --- | --- |
| `driver` | 枚举追加 `PROBE_HTTP` |
| `probe_group_labels` | 目标组标签匹配器 JSON，默认 `{}`；键含固定键 `probe_group`（组名）与目标组自定义标签键 |
| `probe_interval_seconds` | 15 / 30 / 60 / 300，默认 30；探测规则必填 |
| `probe_metric` | `UP` / `LATENCY_MS` / `STATUS_CODE` / `CERT_DAYS`，默认 `UP` |

校验规则：

- 目标组 `request`：`method` ∈ GET/HEAD；`expect_status` 可解析为区间集合；`timeout_ms` ∈ 100..60000；`headers` 拒绝 `Host`、`Content-Length`、`Transfer-Encoding` 等受控头并过滤换行；`body_keyword` ≤256。
- 目标 `spec`：必须 http/https 且带 host；保存时归一化（去默认端口、末尾 `#`、userinfo）；组内去重；每组至少 1 个 ACTIVE 目标。
- 探测规则：`probe_interval_seconds`、`probe_metric` 必填；探测驱动忽略 `expr` / `datasource_labels` / `variables`，保存时归一为空。
- 匹配器值为完整匹配正则，规范化复用 `pkg/labels`。

## 调度与领取

- `alioth` 新增探测调度 tick（与查询调度同一 cron，15s），锁 `alioth:lock:schedule:probe`，扫描 `status=ACTIVE AND driver='PROBE_HTTP' AND next_eval_at <= now`；写 `scheduled_slot = slot`、`next_eval_at = slot + interval`，`slot` 按 interval 对齐。
- `scheduled_slot` 对探测规则是「当前窗口」而不是待清除的批次标记：领取条件为 `scheduled_slot` 非空且 `next_eval_at > now`，窗口内重复领取由 `(target_uid, slot)` 唯一键与单目标锁兜底，因此不需要像查询规则那样在回写后清除。
- 现有 `ScheduleDue` 只处理查询驱动，行为与 15s 节奏不变。
- `EvalService` 追加 `ClaimProbeTasks(executor{kind, labels})` 与 `ReportProbeResults(rule_uid, slot, results[])`，仍只接受 service key `megrez`。
- 领取时按 `probe_group_labels` 匹配 `ACTIVE` 目标组，再展开 `ACTIVE` 目标；一条任务 = 规则 × 目标组 × 目标。无匹配目标组不产生任务。
- 一期只接受 `executor.kind=CLUSTER`，且只返回 `executor_labels` 为空的组。

## 执行（megrez）

新增 `app/megrez/internal/biz/probe`，作为 `megrez` 里与评测并列的 kratos server：

- 轮询 5s；单目标锁 `alioth:lock:probe:{target_uid}:{slot}`（30s TTL、10s 续期）；闸门默认 8 并发；单轮 30s；失败退避上限 30s。
- HTTP 执行：连接 + TLS 校验（未跳过时）+ 期望状态码 + 关键字；记录 `dns_ms/tcp_ms/tls_ms/ttfb_ms`、`status_code`、`cert_days`、`resolved_ip`。
- 失败分类写入 `error_kind`：`dns` / `connect` / `tls` / `timeout` / `status` / `keyword` / `redirect` / `protocol`。
- 响应体最多读 64 KiB；重定向 ≤5 跳且逐跳校验协议。
- 执行器自身故障（配置错误、上下文取消、进程退出）不回写，只记日志；不回写不参与判定，也不触发恢复。
- 每个目标一条结果，逐条回写；一个目标超时不拖累同批其他目标。
- 上报：复用 `ReportStatus`，身份仍是 `MEGREZ`，task 名写探测任务；不新增进程端口。

配置（`conf.proto` + `config.yaml`）：

```yaml
probe:
  enabled: ${MEGREZ_PROBE_ENABLED:true}
  poll_interval: 5s
  concurrency: 8
  max_timeout_ms: 30000
```

## 判定与事件

- 回写落 `probe_results`（唯一键幂等），`alioth` 取该目标 `(now - 窗口, now]` 的结果序列判定：
  - 持续：窗口内每个点都命中，相邻点间隔 ≤ `interval × 1.5`；
  - 最大值 / 最小值：窗口内极值比较；
  - 点数不足 `持续秒数 / 周期` 时不判定。
- 判定函数上移到 `pkg`，与 `megrez` 共用，避免两处语义漂移。
- 事件写入复用与 `applyHits` 同一套辅助：`datasource_uid` 为空、`expr` 为目标 URL、fingerprint `Of(rule_uid, target_uid, labels)`、labels 注入 `datasource`（目标组名）/ `alert_level` / `probe_kind` / `probe_target`，`sample_value` 为观测值，`eval_slot` 为探测槽位；不写 `datasource_level`。
- 通知 `externalURL` 取目标 URL。
- 读取兜底：事件列表/详情/筛选在 `datasource_uid` 为空时用 labels 的 `datasource` 作为 `DatasourceName`。
- 目标增删改、目标组停用/删除、规则匹配器变更都要恢复对应指纹的旧事件，逻辑并入 `recoverChangedRule` 同类路径。
- 订阅不做改动：探测事件靠标签命中；订阅里若设置了数据源条件，不会命中探测事件（文档与表单提示说明）。

## 权限与接口

- `perm.go` 的 resources 增加 `probe`，自动派生 `probe:read` / `probe:write` / `probe:delete`；内置角色按现有规则自动覆盖。
- 目标组：`GET/POST/PUT /v1/probe-target-groups`、`DELETE /v1/probe-target-groups/{uid}`、`GET /v1/probe-target-groups/{uid}`、`PATCH .../{uid}/status`、`POST .../{uid}/probe-now`、`GET .../{uid}/probe-results`、`GET .../match-keys`、`GET .../match-values`、`POST /v1/probe-trial`。
- 规则：`driver` 白名单加 `PROBE_HTTP`；创建/更新请求追加 `probe_group_labels`、`probe_interval_seconds`、`probe_metric`（新 tag，避开已 reserved 的字段号）。

## Web

- 侧边栏「告警」加「探测目标」（`probe:read`，`Target` 图标）；`App.tsx` 加路由；`last-page.ts` 的 knownPaths 加 `/probe-targets`；i18n zh/en 补 nav 与页面文案。
- 探测目标页照数据源页：页头 + 名称/类型筛选 + 左列表（名称、类型、状态、更多菜单）+ 右详情卡（基本信息、目标表、最近探测记录、启停）。对话框为三步：基础 → 目标地址 → 请求与试探测。
- 探测类型下拉一期只提供 `HTTP 探测`；`PROBE_TCP` / `PROBE_ICMP` 不在 proto 白名单里，界面也不出现这两个选项。
- 规则表单选择探测驱动后，在第 1 步把「数据源标签匹配 + 表达式」替换成「匹配目标组」（新增 `probeTargetMatchKeys/Values`，不复用 `datasource-label-keys.ts`，含「将匹配 N 个组 / M 个目标」预览与「按目标组选择」快捷填充）+ 探测周期 + 只读执行点；后续步骤的观测指标、级别条件、文案/标签与查询规则共用。
- 规则列表加类型分段筛选，探测规则的「观测对象」显示匹配到的目标组与目标数；详情抽屉显示目标组列表与探测记录。
- 新类型与驱动补 `enum` 中英文与 `StatusBadge` 色调映射。

## 验证记录

### 2026-10-29 端到端（本地 compose 的 Postgres/Redis + 独立库 `alioth_probe_e2e`）

做法：起 `alioth`（HTTP 18000 / gRPC 19000）与 `megrez`（探测 worker 开启），用 SQL 造一条 15 秒周期、`EQ 0 / FOR 0 秒` 的探测规则与一个目标组，目标指向本地 `python3 -m http.server 18080`。

已验证：

- [x] 目标可达时 `probe_results` 按槽位落库（`ok=t`、`value=1`、`status_code=200`），`scheduled_slot` / `next_eval_at` 每轮推进 15 秒。
- [x] 停掉目标后下一轮 `ok=f`、`value=0`、`error_kind=connect`，同一 `(target, slot)` 没有重复行。
- [x] 判定命中后写入 FIRING 事件：`datasource_uid` 为空、`expr` 是目标地址、`sample_value=0`，labels 为 `{"datasource":"E2E 目标组","env":"e2e","probe_kind":"HTTP","probe_target":"http://127.0.0.1:18080/","probe_target_uid":"e2e-target","team":"platform"}`。
- [x] 目标恢复后事件恢复并归档到租户历史库 `alioth_alert_e2ens` 的周表 `alert_events_20261026`：状态 `RECOVERED`、`ends_at` 有值、标签原样保留，主库行与 `alert_event_firings` 占位都已清掉。
- [x] `TestProbeEventFiringAndRecovery` 集成测试覆盖事件字段、列表数据源名兜底与恢复归档。
- [x] 通知闭环（到任务层）：只配标签 `{"probe_kind":"HTTP"}` 的订阅在 FIRING 时生成一条 `PENDING` 通知任务（订阅、通知组、收件人、邮件渠道都正确），目标恢复后再生成一条 `RECOVERED` 任务；投递本身仍由 `merak` 走既有链路。
- [x] 多副本：同时起两个 megrez 探测 worker，`probe_results` 的 `(target, slot)` 无重复（23 行 / 23 个不同组合），单目标锁生效。
- [x] `probe.enabled=false`：worker 打一行 `probe worker disabled` 后不领任务，等待 20 秒后 `probe_results` 行数没变。

端到端跑出来并已修复的问题：

- service key 白名单没有放行探测 RPC，worker 全部 `PermissionDenied`；
- 判定窗口按最长 `for_seconds` 统一取窗口，`FOR 0` 也会回看两个周期，窗口里只要有一个旧成功点就永不命中；
- 探测事件写入时 `expr` 为空（应在写入时置为目标地址）。

### 2026-10-29 浏览器验收（真实登录态，1366×768 与 390×844）

做法：用临时库建租户与管理员，签一个 access token 注入 `localStorage`（登录页要图形验证码，没有走登录表单），`vite` 起在 8082 反代到探测分支的 `alioth`，用 CDP 驱动真实 Chrome。

- [x] `/probe-targets`：列表（2 个目标组）、名称/类型筛选、详情卡（目标表含地址/状态/最近结果/状态码/耗时/证书天数、最近探测记录、停用/立即探测）都正常；页面只有卡片一个滚动容器。
- [x] 新建目标组三步：名称必填被拦下并显示「请填写必填项」；第 2 步目标地址（添加目标/批量粘贴/单条启用）；第 3 步请求设置与试探测；试探测返回真实失败原因 `dial tcp 127.0.0.1:18080: connect: connection refused`；保存成功后对话框关闭、列表出现新行。
- [x] `/alert-rules`：类型分段筛选只留下探测规则；驱动下拉有「HTTP 探测」；切换后数据源标签匹配与表达式消失，出现「匹配目标组」「按目标组选择」「将匹配 2 个启用目标组，共 4 个目标」「探测周期」「集群内执行器」。
- [x] 窄视口 390×844：无横向溢出，卡片纵向堆叠，详情卡滚到底部后停用/立即探测仍可见。
- [x] 全程无 console 错误。

浏览器验收发现并已修复的问题：数据层 `ruleDriver` 白名单没有 `PROBE_HTTP`，前端能选到 HTTP 探测，提交却一律 400（`fix(alert-rule): accept PROBE_HTTP when saving rules`）。

### 2026-10-29 告警链路投递与运行态收口（alioth + megrez + merak 三进程）

- [x] 通知投递：只配标签 `{"probe_kind":"HTTP"}` 的订阅，FIRING 任务被 `merak` 取走后 `message_logs` 记为 `SENT`，邮件在 Mailpit 里可见；模板 `{{ .summary }}`、`{{ .description }}`、`{{ .externalURL }}`、`{{ index .Labels "probe_kind" }}`、`{{ index .Labels "probe_target" }}`、`{{ index .Labels "datasource" }}` 全部渲染出目标地址与目标组名。
  注意模板字段是小写：`{{ .summary }}` / `{{ .description }}`，不是 `.Summary`（第一次用大写只渲染出 `<no value>`，属测试夹具写错，不是产品问题）。
- [x] 恢复通知：目标恢复后生成 `RECOVERED` 任务并投递，`message_logs` 同为 `SENT`。
- [x] 运行态收口：事件 FIRING 时调用 `DELETE /v1/probe-target-groups/{uid}`，事件立即从主库消失（恢复并归档到租户历史库），并生成一条 `RECOVERED` 通知任务。
- [x] 目标组校验（API 实测）：重复地址、无协议、非 http(s)、云元数据地址、受控请求头分别返回 400 `duplicate probe target` / `probe target must be http or https` / `probe target host is blocked` / `probe header is controlled by the client`。
- [x] 试探测：`POST /v1/probe-trial` 返回 200 与真实失败分类，不写 `probe_results`、不产生事件。

### 待补验

- [ ] 介入 / 抑制 / 手工恢复：探测事件走的是既有告警处置路径，本次没有逐项在运行态点过（由既有告警测试覆盖）。
- [ ] 规则匹配器变更后的旧事件收口：逻辑在 `recoverChangedRule` 里，只有数据层集成覆盖，未在运行态触发。
- [ ] 实时告警列表里探测事件的「数据源」列：标签兜底有集成测试覆盖，浏览器没有单独打开实时告警页核对。

## 验收

- [ ] 新建 HTTP 目标组：3 个地址 + 标签 `env=prod`，保存后出现在列表，详情可见 3 个地址与最近探测记录。
- [ ] 目标组地址重复保存被拒绝；URL 不带 scheme 或非 http(s) 被拒绝；受控请求头被拒绝。
- [ ] 试探测不写 `probe_results`、不产生事件，返回状态码、耗时分解、证书天数。
- [ ] 新建探测规则匹配 `probe_group=核心接口组`，周期 30s；一个周期后 `probe_results` 出现 3 条同 slot 记录，`scheduled_slot` 推进到本轮槽位。
- [ ] 停掉目标服务后，第二个周期 `up=0`；连续 2 次（持续 60s）后产生一条 FIRING 事件，级别取第一条命中的条件。
- [ ] 该事件出现在实时告警，可介入、抑制、手工恢复；恢复后进入历史告警与历史分片，历史分析口径不变。
- [ ] 事件 `datasource` 标签为目标组名，实时/历史列表「数据源」列显示同名，按数据源名可筛出该事件。
- [ ] 通知：按 `probe_kind` 或 `datasource` 标签命中的订阅收到 FIRING 与恢复通知，消息日志可查，模板 `externalURL` 为目标 URL。
- [ ] 规则标签或匹配器变更后旧事件恢复、新指纹建立新事件；删除目标组后其旧事件恢复。
- [ ] 3 个 megrez 副本同时运行：同一 `(target_uid, slot)` 只被一个副本探测，`probe_results` 无重复。
- [ ] 关掉探测 worker（`probe.enabled=false`）后不再产生新探测记录，查询规则评测节奏不受影响。
- [ ] `ICMP` / `TCP` 类型在界面为灰态「后续版本」，后端白名单不接受这些驱动。
- [ ] 浏览器验收：目标组页与规则表单在 1366×768 与 390×844 下可完成新建、保存、切换步骤；深色模式无硬编码颜色。
- [ ] `make vet`、`make lint`、`make test` 与 `web` 的 typecheck / lint / build 通过；`go test ./...` 覆盖新增的判定与 URL 归一化逻辑。
