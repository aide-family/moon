# M26 探测类型扩展（TCP / ICMP）

M26 在 M18（HTTP 探测）与 M21（探测能力按类型细分）之上补齐 **TCP** 与 **ICMP** 两类探测。目标组、规则、调度、领取、判定、事件、通知与告警处置链路全部复用 M18 的既有实现，只按驱动增加目标归一化、执行器与观测指标约束；探测仍然只从集群内执行，不引入边缘执行点。

## 目标与依赖

操作者可以：

1. 在「探测目标」页新建 TCP 目标组（组内多个 `host:port`）和 ICMP 目标组（组内多个 host / IP），与 HTTP 目标组一样共用超时、标签、启停与立即探测；
2. 在「告警规则」页按探测类型新建 TCP / ICMP 探测规则，规则只匹配同类型的目标组，观测指标只能选 `UP` 或 `LATENCY_MS`；
3. 命中后产生与 HTTP 探测同格式的事件，标签 `probe_kind` 分别为 `TCP` / `ICMP`，走同一套订阅、OnCall、通知与告警处置；
4. 在目标组详情与规则详情里看到 TCP 连接耗时 / ICMP 往返时延与失败原因。

依赖：M18（目标组、规则、调度、领取、判定、事件）、M21（探测能力按驱动细分）、M3 / M4（事件写入与通知链路）。

## 数据模型与校验

不新增表，复用 `probe_target_groups` / `probe_targets` / `probe_results`；`driver` 枚举追加 `PROBE_TCP`、`PROBE_ICMP`。现有 `request` JSON 继续保存请求参数，TCP / ICMP 只使用其中的 `timeout_ms`，保存与读取时其余 HTTP 字段不参与执行。

目标地址归一化按驱动区分（`pkg/probe.NormalizeTarget`）：

| 驱动 | 目标格式 | 归一化与校验 |
| --- | --- | --- |
| `PROBE_HTTP` | `http(s)://host[:port][/path?query]` | 维持 M18 行为：去 userinfo、默认端口、片段，host 小写 |
| `PROBE_TCP` | `host:port` | 必须带端口（1–65535），host 小写，IPv6 用 `[addr]:port`；拒绝 scheme、路径与保留网段 |
| `PROBE_ICMP` | `host` 或 IP | 不能带端口与 scheme，host 小写，IPv6 直接写地址；拒绝保留网段 |

- 保留网段沿用 M18：`169.254.0.0/16`、`fe80::/10`、`ff00::/8`、`224.0.0.0/4`、未指定地址；私网默认允许。
- 目标组内按归一化后的 spec 去重；TCP / ICMP 的请求参数只校验 `timeout_ms`（100–60000，默认 3000）。
- 探测规则的 `probe_metric` 按驱动校验：HTTP 允许 `UP`、`LATENCY_MS`、`STATUS_CODE`、`CERT_DAYS`；TCP / ICMP 只允许 `UP`、`LATENCY_MS`。

## 能力与目录

- `perm` 能力目录新增 `probe_tcp`、`probe_icmp`，与 `probe_http` 同属 `probe` 模块、共用 `probe:read/write/delete`；驱动与能力映射 `probeCapabilityDrivers` 追加两行。
- 基础套餐覆盖 `probe` 模块全部能力，因此新租户与升级后的存量租户都获得三种探测类型；套餐仍可只勾选其中一种或几种。
- 旧模块级能力码 `probe_target` 的展开结果改为三种探测能力；「探测目标」菜单由任一探测能力开通。

## 调度、领取与执行

- 调度与领取的「探测规则」判定从 `driver = 'PROBE_HTTP'` 改为「驱动属于三种探测类型」，查询规则仍排除全部探测驱动；其余槽位推进与幂等逻辑不变。
- 领取时按规则的驱动匹配同驱动目标组（既有 `ListProbeTargetGroupsByMatchers` 已按 driver 过滤），一条任务仍是一个规则 × 一个目标组 × 一批目标。
- `megrez` 的探测执行器按 `rule.driver` 分派：
  - `PROBE_TCP`：`net.Dialer` 连接 `host:port`，连接成功即 `ok=1`，`duration_ms` = 建连耗时，`resolved_ip` 记录解析地址；失败分类 `dns` / `connect` / `timeout` / `permission`；
  - `PROBE_ICMP`：解析目标后发送一个 ICMP Echo Request，收到 Echo Reply 即 `ok=1`，`duration_ms` = 往返时延；失败分类 `dns` / `timeout` / `connect` / `permission`；
  - HTTP 行为不变。
- `Result.Value(metric)` 语义不变：`UP` 取 1/0，`LATENCY_MS` 取耗时；TCP / ICMP 的 `status_code`、`cert_days` 恒为 0。
- 执行器自身故障（包括 ICMP 权限不足）**不回写结果**，只记日志并提示检查 `net.ipv4.ping_group_range` 或 `NET_RAW`；这样缺少权限不会被误判成目标不可用、也不会产生误告警。用户在「立即探测」与「试探测」里仍能看到真实的 `permission` 失败原因。

## 事件与判定

- 判定逻辑完全复用 M18：按目标取窗口内的结果序列，用规则级别条件判定；点数不足不判定，未命中调用 `ApplyProbeHits` 恢复旧事件。
- 事件标签 `probe_kind` 由驱动映射为 `HTTP` / `TCP` / `ICMP`；`probe_target`、`probe_target_uid`、`datasource`（目标组名）等键不变，指纹仍为 `Of(rule_uid, target_uid, labels)`。
- 事件描述里的失败原因保留执行器的 `error_message`；目标组 / 规则删除或停用、匹配器与周期变更的收口逻辑不变。

## 接口

- `CreateProbeTargetGroupRequest.driver` / `UpdateProbeTargetGroupRequest.driver` 白名单追加 `PROBE_TCP`、`PROBE_ICMP`。
- `ProbeTrialRequest` 末尾追加 `driver` 字段（字段号 3，留空按 `PROBE_HTTP`）；原有 `spec` / `request` 字段号不变，旧客户端不传 driver 时行为不变。
- 规则创建 / 更新请求的 `driver` 白名单追加两个探测驱动；`probe_metric` 的枚举不变，按驱动在业务层校验。
- 其余目标组 CRUD、立即探测、探测记录与规则接口不变；`EvalService.ClaimProbeTasks` / `ReportProbeResults` 协议不变，执行器从规则快照的 `driver` 分派。

## Web

- 探测目标页：类型筛选与新建下拉从「仅 HTTP」改为已开通的探测类型；第 2 步按类型校验地址（URL / host:port / host）、批量粘贴与占位文案同步；第 3 步 HTTP 显示完整请求参数，TCP / ICMP 只显示超时；试探测与详情表格按类型隐藏「状态码」「证书剩余天数」。
- 告警规则页：探测类型驱动可选 TCP / ICMP；目标组匹配只列出同类型目标组，匹配预览按类型计算；观测指标下拉按驱动裁剪，切换驱动时非法指标回落到 `UP`；类型分段筛选、列表「观测对象」与详情抽屉都按三种探测驱动识别。
- 能力目录、状态徽标与中英文文案补齐 `probe_tcp` / `probe_icmp` / `PROBE_TCP` / `PROBE_ICMP`。

## 部署与运行前提

- **TCP 探测**无额外权限。
- **ICMP 探测**需要 ICMP socket 权限，二选一：
  1. Linux 非特权 ping socket：节点 / 容器 `net.ipv4.ping_group_range` 覆盖 `megrez` 运行用户（部署清单在 Pod 级设置 `sysctls: net.ipv4.ping_group_range = 0 2147483647`）；
  2. `NET_RAW` capability：`docker-compose.yaml` 与 `deploy/k8s/megrez.yaml` 已为 `megrez` 增加 `cap_add: NET_RAW` / `capabilities.add: ["NET_RAW"]`。
- 两者都不可用时，探测结果记录 `permission` 失败：试探测与立即探测直接显示原因，规则执行不产生结果、不误报，日志提示补权限。
- macOS 本地开发没有 Linux 非特权 ping socket，需要 root 才能跑通 ICMP 成功路径；单测在无权限时跳过成功断言。

## 验收

- [x] 新建 TCP 目标组（多个 `host:port`），重复地址被拒绝，无端口 / 带 scheme / 保留网段被拒绝。
- [x] 新建 ICMP 目标组（host 与 IP 混合），带端口或保留网段被拒绝；组内去重生效。
- [x] TCP 试探测：本机监听端口返回成功与耗时，关闭端口返回 `connect` 失败原因；不落库、不产生事件。
- [x] ICMP 试探测：有权限时回显成功与往返时延；无权限时返回 `permission` 原因。
- [x] TCP / ICMP 探测规则一个周期内按槽位落库，`scheduled_slot` / `next_eval_at` 正常推进；停用目标后下一周期产生失败记录。
- [x] 命中后事件 `probe_kind` 为 `TCP` / `ICMP`，`expr` 为 `host:port` / host，通知订阅按标签命中，恢复后归档。
- [x] TCP / ICMP 规则保存 `STATUS_CODE` / `CERT_DAYS` 指标被拒绝；HTTP 规则仍可使用四个指标。
- [x] 规则只匹配同类型目标组：HTTP 规则不会匹配 TCP / ICMP 目标组，反之亦然。
- [ ] `probe.enabled=false` 时探测 worker 不领取任务的行为不变。
- [x] `make vet`、`make lint`、`go test ./...` 与 Web 的 typecheck / lint / build 通过；`pkg/probe` 覆盖 TCP 成功 / 失败、目标归一化、权限失败分类与 ICMP 无权限跳过。

## 验证记录

### 2026-11-01 本地代码级验证

- `gofmt -l pkg api app` 无输出；`make vet` 与 `make lint`（gofmt、`scripts/check-naming.sh`、`buf lint`）通过。
- `make test` 通过：`pkg`、`api`、`app/alioth`、`app/megrez`、`app/merak`、`app/alkaid` 全部包通过；`app/alioth/internal/data` 的 55 个 DB 集成测试因未设置 `ALIOTH_TEST_DSN` 跳过。
- 重点包的 `-race` 通过：`pkg/probe`、`app/megrez/internal/biz/probe`、`app/alioth` 的 `pkg/bo`、`biz/probe`、`biz/alertevent`、`biz/alertrule`。
- `pkg/probe` 新增单测覆盖：HTTP 行为回归；TCP 本机监听成功 / 关闭端口 `connect` 失败；TCP / ICMP 目标归一化与非法写法；`permission` 分类；ICMP loopback 在无权限环境按 `permission` 跳过成功断言。
- `app/megrez` 执行器单测覆盖驱动分派、TCP 成功 / 失败、HTTP 回归、`permission` 不回报与告警日志。
- `app/alioth` 单测覆盖按驱动的目标归一化、指标白名单、`probe_kind` 映射、`$externalURL`、调度/领取驱动集合与 `ruleDriver` 白名单。
- Web：`pnpm typecheck`、`pnpm lint`、`pnpm build` 通过（仅既有 chunk 体积提示）；`buf generate` 幂等。
- 部署清单：`docker compose config -q` 与 `kubectl apply --dry-run=client -f deploy/k8s/megrez.yaml` 通过。

### 2026-11-01 本地端到端（独立库 `alioth_probe_e2e_m26` + 独立 Redis 6380）

做法：本机起 alioth（HTTP 18000 / gRPC 19000）、megrez（评测 + 探测 worker）与 merak；PostgreSQL 与 Redis 与开发实例隔离；用 SQL 造租户与负责人后，全部操作经 HTTP API 完成（Bearer JWT + `X-Namespace`）。

- [x] TCP、ICMP 目标组创建成功并落库正确的 `driver`；TCP 带 URL、ICMP 带端口、组内重复地址均返回 400。
- [x] TCP / ICMP 试探测成功（`ok=true`、解析 IP）；TCP DNS 失败分类 `dns`，ICMP 对 TEST-NET 超时分类 `timeout`。
- [x] 探测规则按 15 秒槽位落库，`scheduled_slot` / `next_eval_at` 连续推进；TCP、ICMP 规则只匹配各自类型的目标组。
- [x] 停掉 TCP 目标后 `ok=false`、`error_kind=connect`、`value=0`，UP EQ 0 规则产生 FIRING 事件：`probe_kind=TCP`、`expr=127.0.0.1:18080`、`datasource` 为目标组名、`sample_value=0`。
- [x] 恢复 TCP 目标后结果恢复 `ok=true`，事件归档到租户历史库：`status=RECOVERED`、`ends_at` 有值、标签保留。
- [x] ICMP 规则（UP EQ 1）产生 FIRING 事件（`probe_kind=ICMP`、`expr=127.0.0.1`、`sample_value=1`）；删除目标组后事件恢复并归档。
- [x] TCP / ICMP 规则保存 `STATUS_CODE` / `CERT_DAYS` 指标均返回 400；HTTP 试探测回归成功（200、`value=1`）。
- [x] 「立即探测」写入 `rule_uid` 为空的结果；目标组探测记录 API 返回目标地址、成功状态与规则名。
- [x] 能力开关：从租户能力快照移除 `probe_tcp` 后创建 TCP 目标组返回 403 `feature not enabled`，TCP 目标组列表为空；恢复能力后正常。
- [x] 通知闭环：只配 `probe_kind=TCP` 标签的订阅收到 FIRING 与 RECOVERED 各一条任务，merak 投递 `SENT`，Mailpit 收到邮件；模板中 `probe_kind`、`externalURL` 渲染出 TCP 目标；消息日志归档到租户库。
- [x] 两个 megrez 副本同时运行：同一槽位仍只有一条结果（29 行 / 29 个唯一 `target+slot`），无重复探测或重复回写；进程同时上报评测与探测两个实例心跳。
- [x] 本机 macOS 上 ICMP 非特权路径实际回显成功（`127.0.0.1`），`TEST-NET` 超时符合预期。

端到端发现并已修复的问题：

- TCP / ICMP 目标组或试探测请求只传 `timeout_ms` 时，proto 校验因 `method` 不允许空、`timeout_ms` 要求 ≥100 而返回 400，与「TCP / ICMP 只使用超时」的接口说明不符。已放宽为 `method` 允许空、`timeout_ms` 允许 0（业务层补默认 3000），并重生成 API 代码；修复后同一请求成功。

### 未执行

- 浏览器页面验收（登录态、1366×768 / 390×844、交互与滚动）未执行，没有可用的真实登录会话；目标组/规则页仅通过 Web 构建与类型检查验证。
- `probe.enabled=false` 时不再领取探测任务由 M18 已验证，本次未重复执行。
- 真实公网 ICMP / TCP 目标的探测未执行，本机用回环地址与 TEST-NET 覆盖成功与超时路径。

## 文档同步

- `docs/roadmap.md` 增加 M26 与依赖边。
- `docs/design.md` 的探测章节从「一期只有 HTTP」改为「支持 HTTP / TCP / ICMP」。
- `docs/probe-rule-design.md` 标注后续项已由 M26 实现。
- `docs/iterations/conventions.md` 的能力数量与基础套餐说明同步。
