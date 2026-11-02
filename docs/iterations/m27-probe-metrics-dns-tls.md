# M27 探测指标扩展、HTTP 断言与 DNS / TLS 探测

M27 在 M18 / M26 的探测链路上继续扩展三类能力：

1. **观测指标扩展**：HTTP / TCP / ICMP 已有结果里的耗时分解与 ICMP 统计（丢包、抖动、最小/最大 RTT）变成规则可选的观测指标；HTTP 追加响应头断言与响应体正则。
2. **新探测类型**：`PROBE_DNS`（域名解析）与 `PROBE_TLS`（TLS 握手与证书），复用目标组、调度、领取、判定、事件与通知链路。
3. **探测请求加密**：目标组 `request` JSON（含自定义请求头里的 token）不再明文入库，使用 `ALIOTH_ENCRYPT_KEY` 加密保存；读取兼容历史明文行。

目标组、调度、领取、判定、事件与通知语义不变。新增类型仍需实现执行器后才能上架能力，不能只卖能力。

## 目标与依赖

操作者可以：

1. 在 HTTP 规则里选择 `DNS_MS`、`TCP_CONNECT_MS`、`TLS_MS`、`TTFB_MS` 等耗时指标，并为 HTTP 目标组配置响应头断言与响应体正则；
2. 在 ICMP 规则里选择 `PACKET_LOSS`、`JITTER_MS`、`MIN_RTT_MS`、`MAX_RTT_MS`，并配置探测次数与 payload 大小；
3. 新建 DNS 目标组（域名）与 TLS 目标组（host:port），规则可按同类型匹配并选择各自指标；
4. 事件标签 `probe_kind` 分别为 `DNS` / `TLS`，走同一套告警、订阅、OnCall 与通知链路；
5. 目标组请求 JSON（含自定义请求头）以密文落库，接口在授权租户内保持明文回读以便编辑；迁移包依赖同一个 `ALIOTH_ENCRYPT_KEY` 才能解密。

依赖：M18（目标组、规则、调度、领取、判定、事件）、M21（能力按类型拆分）、M26（TCP / ICMP 驱动与指标校验模板）、M3 / M4（事件与通知）。

## 驱动与能力

| 驱动 | 目标格式 | 能力码 | 说明 |
| --- | --- | --- | --- |
| `PROBE_HTTP` | `http(s)://...` | `probe_http` | 已有 |
| `PROBE_TCP` | `host:port` | `probe_tcp` | 已有 |
| `PROBE_ICMP` | host / IP | `probe_icmp` | 已有 |
| `PROBE_DNS` | 域名（不带 scheme / 端口 / 路径） | `probe_dns` | 新增：按记录类型解析并可断言结果 |
| `PROBE_TLS` | `host:port` | `probe_tls` | 新增：TLS 握手、证书与协议信息 |

- 能力目录：`probe` 模块下五个能力，基础套餐覆盖全部五种（能力总数 55 → 57，基础套餐 28 → 30）；探测菜单由任一探测能力开通。
- 旧版 `probe_target` 模块级能力展开为五种探测能力。

## 指标目录

新增指标常量（proto 字符串 / bo 常量）：

| 指标 | 适用驱动 | 取值 |
| --- | --- | --- |
| `UP` | 全部 | 1 / 0 |
| `LATENCY_MS` | 全部 | 单次探测总耗时；TCP 为建连、ICMP 为平均 RTT、DNS 为解析、TLS 为整次握手 |
| `STATUS_CODE` | HTTP | HTTP 状态码 |
| `CERT_DAYS` | HTTP / TLS | 证书剩余天数 |
| `DNS_MS` | HTTP / TCP / DNS / TLS | 域名解析耗时 |
| `TCP_CONNECT_MS` | HTTP / TCP / TLS | TCP 建连耗时 |
| `TLS_MS` | HTTP / TLS | TLS 握手耗时 |
| `TTFB_MS` | HTTP | 首字节耗时 |
| `PACKET_LOSS` | ICMP | 丢包百分比 0–100 |
| `JITTER_MS` | ICMP | 相邻回包 RTT 平均绝对差（毫秒） |
| `MIN_RTT_MS` | ICMP | 最小 RTT |
| `MAX_RTT_MS` | ICMP | 最大 RTT |
| `ANSWER_COUNT` | DNS | 解析到的记录条数 |

`bo.ValidProbeMetric(driver, metric)` 按驱动校验；Web 的指标下拉使用同一张表：

- HTTP：`UP`、`LATENCY_MS`、`STATUS_CODE`、`CERT_DAYS`、`DNS_MS`、`TCP_CONNECT_MS`、`TLS_MS`、`TTFB_MS`
- TCP：`UP`、`LATENCY_MS`、`DNS_MS`、`TCP_CONNECT_MS`
- ICMP：`UP`、`LATENCY_MS`、`PACKET_LOSS`、`JITTER_MS`、`MIN_RTT_MS`、`MAX_RTT_MS`
- DNS：`UP`、`LATENCY_MS`、`DNS_MS`、`ANSWER_COUNT`
- TLS：`UP`、`LATENCY_MS`、`DNS_MS`、`TCP_CONNECT_MS`、`TLS_MS`、`CERT_DAYS`

## 请求参数（`ProbeRequestConfig` 追加字段）

HTTP 探测使用已有字段 + 新断言；TCP 继续只使用超时；ICMP / DNS / TLS 使用各自新增字段。所有新字段对不适用的驱动忽略，不参与校验。

| 字段（JSON） | 类型 | 适用 | 规则 |
| --- | --- | --- | --- |
| `expect_headers` | JSON 对象文本 | HTTP | 响应头名 → 完整匹配正则；头缺失或不匹配即失败 |
| `body_regex` | string ≤512 | HTTP | 响应体完整匹配正则；与 `body_keyword` 同时配置时两者都要满足 |
| `icmp_count` | int 0–10 | ICMP | 0 视为 1；单次探测连续发送的 echo 数量 |
| `icmp_packet_size` | int 0–1400 | ICMP | 0 视为 56；payload 字节数 |
| `dns_record_type` | enum | DNS | `A` / `AAAA` / `CNAME` / `MX` / `TXT` / `NS`，空视为 `A` |
| `dns_expect_values` | string ≤1024 | DNS | 逗号分隔的完整匹配正则；每个正则至少命中一条记录才算成功，空表示只要有记录 |
| `dns_server` | string ≤255 | DNS | 可选 `host:port`；留空使用系统 resolver；IP 目标沿用保留网段拦截 |
| `tls_server_name` | string ≤255 | TLS | 可选 SNI；留空取目标 host；不允许端口、scheme、空白 |

请求校验（`probe.ValidateRequestForDriver`）：

- HTTP：已有校验 + 响应头名/值不加换行、值可编译为正则、`body_regex` 可编译。
- TCP：只校验超时（100–60000，0 补 3000）。
- ICMP：超时；`icmp_count` 1–10（0 补 1）；`icmp_packet_size` 0–1400（0 补 56）。
- DNS：超时；记录类型白名单；`dns_expect_values` 每个正则可编译；`dns_server` 必须 `host:port` 且端口 1–65535、host 不能是保留网段地址。
- TLS：超时；`tls_server_name` 只能是主机名（无端口/路径/scheme）；SNI 默认取目标 host。

## 目标归一化

- DNS：只接受域名。转小写、去掉末尾点；要求由 1–253 个字符组成，标签 1–63、只含字母/数字/连字符且不以连字符开头或结尾；拒绝 scheme、端口、路径、空白、IP 字面量与保留网段地址。
- TLS：与 TCP 相同（`host:port`，端口 1–65535，IPv6 方括号，拒绝 scheme/路径/保留网段）。
- 其他驱动行为不变。

## 执行语义

### HTTP（追加断言）

成功条件在原有「连接 + TLS + 期望状态码 + 关键字」上追加：

1. `expect_headers`：每个响应头必须存在，且所有值用 `, ` 连接后完整匹配对应正则；
2. `body_keyword`：响应体包含关键字（已有）；
3. `body_regex`：响应体完整匹配正则。

失败分类新增 `assert`（断言不匹配，`error_message` 指明是哪个头/正文正则）。响应体仍最多读取 64 KiB。

### ICMP（多包与统计）

- 按 `icmp_count` 连续发送 echo，逐包等待回包；每包读超时取 `timeout_ms / count`（最少 50ms），总时长受 `timeout_ms` 约束。
- `OK` = 至少收到 1 个回包；全部丢失时 `error_kind=timeout`、`error_message` 带丢包率。
- `duration_ms` = 平均 RTT；`packet_loss` = 丢失百分比（四舍五入）；`jitter_ms` = 相邻 RTT 平均绝对差；`min_rtt_ms` / `max_rtt_ms` 为观测极值；`resolved_ip` 记录解析地址。

### DNS

- 按 `dns_record_type` 使用系统 resolver 或 `dns_server` 自定义 `net.Resolver` 解析；记录 `dns_ms`、`answer_count`、`dns_answers`（逗号分隔文本）、`resolved_ip`（A/AAAA 第一条）。
- 成功 = 解析成功 + 每个 `dns_expect_values` 正则至少命中一条记录；解析失败（含 NXDOMAIN）`error_kind=dns`；断言不匹配 `error_kind=assert`。
- CNAME/MX/TXT/NS 的答案文本：CNAME 取目标，MX 取 `priority host` 文本，TXT 取各字符串拼接，NS 取主机名。

### TLS

- 解析 host → 记录 DNS 耗时 → TCP 建连 → TLS 握手（SNI 取 `tls_server_name` 或 host，`InsecureSkipVerify` 取 `skip_tls_verify`）。
- 成功 = 握手成功；`verify_cert=true` 时校验证书未过期（即使跳过了链校验），过期记 `error_kind=tls` 且失败。
- 记录 `dns_ms` / `tcp_ms` / `tls_ms`、`cert_days`、`tls_version`（如 `TLSv1.3`）、`tls_cipher`、`cert_subject`、`cert_issuer`、`cert_sans`（逗号分隔）、`resolved_ip`。
- 失败分类：`dns` / `connect` / `tls` / `timeout` / `permission`。

## 结果存储与接口

`probe_results` 追加列（默认值均为 0 / 空串）：

| 列 | 类型 | 说明 |
| --- | --- | --- |
| `packet_loss` | int | ICMP 丢包百分比 |
| `jitter_ms` | int | ICMP 抖动 |
| `min_rtt_ms` / `max_rtt_ms` | int | ICMP RTT 极值 |
| `answer_count` | int | DNS 记录数 |
| `dns_answers` | string | DNS 答案文本 |
| `tls_version` / `tls_cipher` | string | TLS 协议与套件 |
| `cert_subject` / `cert_issuer` / `cert_sans` | string | 证书信息 |

`ProbeResultItem`（alioth HTTP）追加字段号 20–30：
`packet_loss`、`jitter_ms`、`min_rtt_ms`、`max_rtt_ms`、`answer_count`、`dns_answers`、`tls_version`、`tls_cipher`、`cert_subject`、`cert_issuer`、`cert_sans`。

`megrez.v1.ProbeRequestSnapshot` 追加字段号 9–16（与 `ProbeRequestConfig` 同名同序）；`megrez.v1.ProbeResult` 追加字段号 15–25（与 `ProbeResultItem` 同名同序，去掉 target/slot 等元数据）。

## 请求加密

- 目标组 `probe_target_groups.request` 整段 JSON 使用 `secretBox`（AES-GCM，密钥派生同 `ALIOTH_ENCRYPT_KEY`）加密后入库，密文带 `enc:v1:` 前缀。
- 读取：有前缀则解密；没有前缀按历史明文 JSON 解析，随下一次保存自动加密。
- 解密失败（密钥不一致/数据损坏）返回服务端错误，不静默当空请求。
- 迁移包导入依赖同一 `ALIOTH_ENCRYPT_KEY`（已有 `encryptKeyCheck` 校验）。
- megrez 快照仍通过集群内 gRPC 传递明文请求（与数据源密码同一边界），不新增外泄面。

## Web

- 探测目标页：类型下拉/筛选、目标占位与校验、批量粘贴按五种驱动区分；第 3 步按驱动显示参数：
  - HTTP：已有字段 + 响应头断言（键值行）+ 响应体正则；
  - TCP：超时；
  - ICMP：超时 + 探测次数 + payload 大小；
  - DNS：超时 + 记录类型 + 期望值 + DNS 服务器；
  - TLS：超时 + SNI + 跳过 TLS 校验 + 校验证书。
- 目标详情与探测记录表：HTTP 保留状态码/证书列；非 HTTP 增加「探测详情」列，展示 DNS 答案/记录数、ICMP 丢包/抖动/min/max、TLS 版本/套件/证书主体，代替 HTTP 专属列。
- 告警规则页：各驱动指标下拉使用上面的指标表；切换驱动时非法指标回落 `UP`。
- 能力目录、状态徽标、中英文文案补齐 `probe_dns` / `probe_tls` / `PROBE_DNS` / `PROBE_TLS` / 新指标。

## 验收

- [x] DNS 目标组：合法域名保存成功；IP 字面量、带端口/路径、非法标签被拒绝；`dns_server` 非法被拒绝。
- [x] TLS 目标组：`host:port` 保存成功；带 URL/无端口被拒绝。
- [x] DNS 试探测：系统 resolver 成功并返回记录数/答案/解析耗时；NXDOMAIN 返回 `dns`；期望值不匹配返回 `assert`。
- [x] TLS 试探测：自签证书 + 跳过校验成功并返回证书天数/协议/套件；不跳过校验返回 `tls`。
- [x] ICMP 探测次数 3 时返回丢包率、抖动、min/max RTT；全部丢包返回 `timeout` 且 `packet_loss=100`。
- [x] HTTP 响应头断言/响应体正则命中与不命中分别成功/`assert`。
- [x] 新指标可保存为规则条件并按 `probe_results.value` 正确判定（DNS_MS、ANSWER_COUNT、PACKET_LOSS、CERT_DAYS 等）。
- [x] 目标组保存后数据库 `request` 列带 `enc:v1:` 前缀且不含明文 token；接口回读为明文请求头；megrez 实际带解密后的请求头探测成功。
- [x] 历史明文 request 行可读，保存后自动加密（真实 Postgres 集成测试覆核）。
- [x] DNS / TLS 规则命中产生 `probe_kind=DNS` / `TLS` 事件；通知链路复用 M26 已验证路径，本次未重复启动 merak 投递。
- [x] 能力开关：未开通 `probe_dns` / `probe_tls` 时对应类型目标组创建返回 403，列表不显示。
- [x] `make vet`、`make lint`、`make test`、Web typecheck/lint/build 通过。

## 验证记录

### 2026-11-02 本地端到端（独立库 `alioth_probe_e2e_m27` + 独立 Redis 6381）

做法：本机起 alioth（18100/19100）与 megrez（探测 worker）；本地 HTTP 目标（18280）、自签 TLS 目标（18443）与自建 UDP DNS（15353）；用 SQL 造租户与负责人后全部经 HTTP API 操作。能力快照先保留全部五种探测能力。

- [x] 能力：`/v1/self/permissions` 返回 `probe_http`、`probe_tcp`、`probe_icmp`、`probe_dns`、`probe_tls`。
- [x] HTTP：带请求头 `X-Token`、响应头断言 `X-Health: ok`、响应体正则与关键字的目标组保存成功；试探测成功；断言不匹配返回 `assert`；请求 `/token` 时解密后的 `X-Token` 确实发出（成功 200，去掉头则 401 `status`）。
- [x] 加密：DB `probe_target_groups.request` 以 `enc:v1:` 开头且不含明文 token；API 回读为明文请求头；历史明文行读取与保存后加密由真实 Postgres 集成测试覆盖。
- [x] DNS：自定义 `dns_server`（127.0.0.1:15353）+ A 记录 `probe.e2e.local` 成功（答案 127.0.0.1、记录数 1、DNS_MS）；TXT 记录成功（答案 healthy）；NXDOMAIN 返回 `dns`；期望值不匹配返回 `assert`；IP 字面量目标被 400 拒绝。
- [x] TLS：`127.0.0.1:18443` + SNI `probe.e2e.local` + 跳过校验成功，返回 `cert_days=29`、`TLSv1.2`、套件、Subject/Issuer/SAN；不跳过校验对自签证书返回 `tls`；无端口目标被 400 拒绝；SNI 带端口被 400 拒绝。
- [x] ICMP：`icmp_count=3`、payload 64 成功并返回 `packet_loss=0`、jitter/min/max；对 TEST-NET 全丢包返回 `timeout`、`packet_loss=100`；`icmp_count=11` 被 400 拒绝。
- [x] 新指标判定：四条规则分别用 `ANSWER_COUNT EQ 1`、`TTFB_MS GTE 0`、`PACKET_LOSS EQ 0`、`CERT_DAYS GT 0`；15 秒槽位结果中 `value` 分别为 1 / 0 / 0 / 29，均产生 FIRING 事件，事件 `probe_kind` 为 `DNS` / `HTTP` / `ICMP` / `TLS`，`expr`、`datasource` 与 `sample_value` 正确。
- [x] 立即探测（probe-now）对 DNS / TLS 组返回并落库新增字段（answer_count/dns_answers、cert_days/tls_version/tls_cipher）。
- [x] 能力开关：从租户能力快照移除 `probe_dns`/`probe_tls` 后创建对应目标组返回 403 `feature not enabled`，列表为空；恢复后正常。
- [x] 删除 DNS/TLS/ICMP/HTTP 目标组后 4 条事件全部从主库消失并归档到租户历史库：`status=RECOVERED`、`ends_at` 有值、`probe_kind` 标签保留。
- [x] `make vet` / `make lint` / `make test`（含 `app/alioth` 全部包）与 Web `pnpm typecheck` / `lint` / `build` 通过；`buf generate` 幂等。

端到端发现并已修复的问题：

- **探测指标枚举未同步**：`alert_rules.probe_metric` 的 ent 枚举仍只有 M18 的四个指标，数据层 `probeMetricValue` 对未知指标静默回落 `UP`。表现是 DNS `ANSWER_COUNT`、ICMP `PACKET_LOSS`、HTTP `TTFB_MS` 能保存成功但实际按 `UP` 判定。已把九个新指标加入 ent 枚举并重新生成，同时扩展 `probeMetricValue` 白名单；修复后四条规则按各自指标正确落值与判定。
- **驱动枚举未同步**：`alert_rule.driver` / `probe_target_group.driver` 的 ent 枚举漏了 `PROBE_DNS` / `PROBE_TLS`，由 alioth 实现方发现，已补值并重新生成。
- DNS NXDOMAIN 的 `error_message` 在自定义 `dns_server` 下仍显示系统 resolver 地址（Go resolver 行为），`error_kind=dns` 正确；属于提示信息的小瑕疵，不影响判定。

### 未执行

- 浏览器页面验收（登录态、1366×768 / 390×844、交互与滚动）未执行，没有可用的真实登录会话；Web 仅通过 typecheck / lint / build。
- 通知投递（merak → SMTP）未在 M27 重新跑；DNS/TLS/ICMP 事件标签已与 M26 通知链路对齐，M26 已验证 probe 事件的通知闭环。
- 公网 DNS / TLS / ICMP 目标未测试，本机用本地 DNS/TLS 服务与 TEST-NET 覆盖成功与失败路径。

## 文档同步

- `docs/roadmap.md` 增加 M27。
- `docs/design.md` 探测章节补充指标、断言、DNS/TLS 与请求加密。
- `docs/iterations/conventions.md` 能力数量（57）与基础套餐（30 项）同步。
