# Alioth 待做计划书

本文只收「本轮不做、后续可做」的候选，是 [roadmap.md](roadmap.md) 的候补池。本轮已确认的 M28–M32（告警处置、升级与值班轮换、状态页与 SLA、Kubernetes 控制台增强、Grafana 大盘）只记在 roadmap，这里不重复。

维护方式：

- 每轮迭代结束后由文档负责人更新本文：已完成项移到 [roadmap.md](roadmap.md) 并在里程碑验收里体现，不再留在候选池；被打回或范围变化的项改证据与结论。
- 新增候选必须带「文件:行」的现状证据。没有代码证据的猜想放「待验证」，不能直接给优先级。
- 优先级：**P0** 质量与安全底线，下一轮必须先做；**P1** 价值明确、成本可控；**P2** 等业务或使用量触发再启动。
- 规模：**S** 单点改动，**M** 一个里程碑内的子项，**L** 需要独立里程碑。本文不估工期。
- 每项固定字段：现状证据、缺口、要动的层、价值、依赖、规模、优先级、风险。第十项是已决策「不做」的清单，避免重复讨论。

| # | 候选项 | 优先级 | 规模 | 主要依赖 |
| --- | --- | --- | --- | --- |
| 1 | 短信发送（阿里云 / 腾讯云） | P1 | L | M4/M8 通知链路、腾讯云去留的产品决策 |
| 2 | 边缘探测执行点（alkaid） | P2 | L | M18/M26/M27、M5 Agent 流 |
| 3 | 平台自身可观测性 | P1 | M | 无硬依赖，与 M32 大盘联动 |
| 4 | 安全基线 | P0 | L | 无，可与迭代并行 |
| 5 | 自动化 e2e | P0 | M | compose 测试栈、稳定测试账号 |
| 6 | 数据保留与容量治理 | P1 | M | 第 3 项指标、合规保留期 |
| 7 | 商业化（支付 / 发票 / 用量） | P2 | L | M20/M21 订单与套餐 |
| 8 | Jaeger 规则评测 | P2 | M/L | M3、M10–M12 的数值化先例 |
| 9 | 小步速赢 | P1（⑧ P2） | S–M | 彼此独立 |
| 10 | 明确不做 / 保持现状 | — | — | 决策清单，不排期 |

## 1. 短信发送（阿里云 / 腾讯云）

- **现状证据**：能力目录已可售两个短信能力（`app/alioth/pkg/perm/capability.go:43-44`），腾讯云条目自述「只上架能力，暂不支持配置与发送」（`app/alioth/pkg/perm/capability.go:206-210`）；渠道/模板的 `app` 枚举只有 `SMS_ALICLOUD`，没有腾讯云（`api/alioth/v1/notification.proto:297`、`:482`）；渠道字段只有 endpoint、username、password、fromAddress（`api/alioth/v1/notification.proto:266-288`），缺 AccessKey/SecretKey、签名 `SignName`、模板码、SDK AppId 等短信必需字段；投递端直接把短信任务判失败（`app/merak/internal/biz/notify/deliver.go:29-32`），能力侧同样拒绝（`app/alioth/internal/biz/notification/notification.go:238`）。
- **缺口**：各提供商的凭据与模板参数模型、发送器与限频重试、短信模板校验（变量占位、长度）、可选送达回执；腾讯云是否保留属产品决策——若保留则补枚举与表单，若不保留应从能力目录与套餐里移除或标灰，避免「可售不可用」。
- **要动的层**：proto（渠道/模板字段与枚举）→ ent schema 与迁移 → biz（类型矩阵、模板校验）→ data（加密存储）与 `merak` 发送器 → web 表单与 i18n；签名工具可放 `pkg`。
- **价值**：补齐通知渠道最后一块；无企业 IM、收信不稳定的场景里，短信是唯一通道。
- **依赖**：M4/M8 通知链路、M21/M24 渠道能力细化、`ALIOTH_ENCRYPT_KEY`；腾讯云保留决策。
- **规模**：L（只做阿里云约 M）。
- **优先级**：P1，先做阿里云。
- **风险**：凭据与签名泄露、短信计费与限频失控、模板受外部审核约束；已购套餐里含腾讯云能力时，移除要给出迁移说明。

## 2. 边缘探测执行点（alkaid 作为探测执行器）

- **现状证据**：目标组的 `executor_labels` 只存不调度（`docs/design.md:387`），领取侧只返回空标签的组（`app/alioth/internal/data/probe_repo.go:282-284`）；`megrez` 领取时写死 `CLUSTER`（`app/megrez/internal/biz/probe/worker.go:210`），服务端也强制该值并拒绝其他执行器（`app/alioth/internal/service/eval.go:133-135`）；`alkaid` 只有健康检查与 Agent 流，没有探测执行代码（`app/alkaid/internal/biz/biz.go:12-30`，该应用内无 probe 业务）。
- **缺口**：执行器身份（Namespace + labels）、领取协议携带执行器、调度按 `executor_labels` 过滤、`alkaid` 侧执行器与结果回写、边缘时延/丢包的可比性、边缘任务的超时与配额、带执行器标签的结果指标。
- **要动的层**：`api/probe` proto（领取带执行器身份与标签）、`alioth` 数据层过滤、`megrez` 多执行器身份、`alkaid` biz 与出站流协议、部署配置（ICMP 需要 `NET_RAW` 提示）。
- **价值**：从公网与多地域看真实可用性，补上「只在集群内探测」的盲区；目标组、事件与通知链路可全部复用。
- **依赖**：M18/M26/M27 探测链路、M5 Agent 出站流与令牌、M19 能力。
- **规模**：L。
- **优先级**：P2。
- **风险**：协议与身份跨三个进程改动；边缘时钟与网络差异会改变同一阈值的语义；`alkaid` 部署面扩大后的版本兼容。

## 3. 平台自身可观测性

- **现状证据**：`alioth` 只注册健康检查与业务路由（`app/alioth/internal/server/http.go:62-131`），中间件只有 recovery、鉴权、校验（`app/alioth/internal/server/http.go:62`），没有访问指标或请求 ID；`megrez`、`merak` 的 HTTP 只有 `healthserver`（`app/megrez/internal/server/http.go:26`、`app/merak/internal/server/http.go:26`）；全仓没有 `/metrics` 注册；prometheus/otel 在 `alioth`、`megrez`、`merak` 里都是 indirect 依赖（`app/alioth/go.mod:81-94`），`pkg` 里的直接依赖只用于模板标签解析（`pkg/alerttmpl/template.go:18-19`）。
- **缺口**：`/metrics`（HTTP 延迟与错误、调度扫描、任务领取、投递成功率、Redis 锁失败）、贯穿日志与响应头的请求 ID、统一结构化日志字段、可选 OTel 导出、抓取配置与示例面板。
- **要动的层**：`pkg`（指标与中间件、请求 ID）、各应用的 server 中间件、biz 埋点（调度/评测/投递）、deploy（抓取配置）、docs。
- **价值**：排障从翻日志变成看指标；也是状态页与大盘（M30/M32）的数据基础。
- **依赖**：无硬依赖；与 M32 联动。
- **规模**：M。
- **优先级**：P1。
- **风险**：高基数标签（规则/任务）会拖垮 Prometheus；请求 ID 不与 nginx 日志打通就断链。

## 4. 安全基线

- **现状证据**：登录只有邮箱验证码 60 秒节流与邮箱级失败锁定（`app/alioth/internal/biz/auth/auth.go:47`、`:254`、`:278-285`、`:309-310`），没有按 IP/账号的通用失败限流或接口限流（`app/alioth/internal/server/http.go:62` 中间件里没有 limiter）。刷新令牌是自签 JWT（`app/alioth/internal/biz/auth/auth.go:619-623`），刷新请求体为空（`api/alioth/v1/identity.proto:828`），全仓没有 jti、黑名单或会话版本，因此没有「登出全部设备 / 吊销刷新令牌」。全仓无 MFA/TOTP/WebAuthn。只有 SSH 审计表（`app/alioth/internal/data/ent/schema/ssh_command_audit.go:1`），没有通用操作审计。nginx 只有三个响应头，没有 CSP 与 HSTS（`web/nginx.conf:39-41`）。CI 六个 job（`.github/workflows/ci.yml:18,42,89,110,136,153`）没有 govulncheck、npm audit 或镜像扫描。
- **缺口**：六件套——登录/接口限流、登出全部设备与刷新令牌吊销、MFA、通用操作审计、CSP/HSTS、CI 依赖与镜像扫描。
- **要动的层**：`pkg` 中间件与 `auth` biz、ent（会话版本或会话表、通用审计表）、`web/nginx.conf`、CI、部署文档。
- **价值**：生产化门槛；M23 迁移包与 M25 审计都还没有覆盖平台级安全审计。
- **依赖**：无；通用审计表要与 M28 的事件时间线区分（操作审计 vs 告警时间线）。
- **规模**：L（限流 S、吊销 S/M、MFA M、审计 M、CSP/HSTS S、CI S）。
- **优先级**：P0。
- **风险**：CSP 会打断内联主题启动脚本，需要 nonce/hash；MFA 会收窄第三方登录路径；审计表增长依赖第 6 项的保留策略。

## 5. 自动化 e2e

- **现状证据**：`web/package.json:6-15` 只有 dev/build/lint/format/typecheck/preview，没有 test；全仓无 Playwright/vitest 配置；里程碑的页面验收一直写在人肉清单里（`docs/iterations/conventions.md:225`），M27 明确记录浏览器验收未执行（`docs/iterations/m27-probe-metrics-dns-tls.md:204`）。
- **缺口**：Playwright 骨架与种子数据、登录态复用、关键路径用例（登录、租户切换、告警处置、通知试发、K8s 只读页、探测页）、CI job、失败截图与 trace。
- **要动的层**：`web` 测试目录与脚本、CI、compose 测试栈、迭代验收模板。
- **价值**：把人工浏览器验收变成可重复门禁；M28–M32 页面更多，回归成本只会继续上升。
- **依赖**：现有 `make dev-up` 的 compose 栈；需要稳定的测试账号与租户。
- **规模**：M。
- **优先级**：P0。
- **风险**：CI 起全栈耗时长；断言易脆（用 role/label 定位，少用快照）；第三方 OAuth 无法自动化，走邮箱验证码通道。

## 6. 数据保留与容量治理

- **现状证据**：分区只预建不回收——`alert_events`、`message_logs` 两张表（`app/alioth/internal/data/partition.go:21`），维护器提前四周创建（`app/alioth/internal/biz/partition/partition.go:17`），全仓没有 DROP 历史分区的调度；历史告警与消息日志按租户库 + 周表创建后从不删除（`app/alioth/internal/data/history_store.go:18`、`:49`）；`probe_results` 有 7 天清理，但传入的「每目标 200 条」上限被忽略，只按时间删（`app/alioth/internal/biz/alertevent/probe.go:388-393`、`app/alioth/internal/data/probe_repo.go:624-629`）；路由索引 `alert_history_events`、`message_log_shards` 无清理（`app/alioth/internal/data/maintenance_repo.go:77-93` 标为运行数据）。
- **缺口**：保留策略（按表/按租户/按天可配）、分区 DETACH 与 DROP 的调度和审计、历史周表与租户库清理、目标条数上限、容量与增长率指标、清理前导出。
- **要动的层**：data（淘汰 SQL、分区维护）、biz（策略与锁）、conf（默认值与上限）、proto/web（管理员可见策略与容量）、docs。
- **价值**：长期运行不靠人工删表；库膨胀会先拖垮告警写入。
- **依赖**：第 3 项指标用于观察清理效果；合规保留期需产品确认。
- **规模**：M。
- **优先级**：P1。
- **风险**：误删不可恢复（先 DETACH 保留若干天再 DROP）；分区 DDL 的锁；保留期与历史查询 30 天上限要对齐。

## 7. 商业化（真实支付、发票、用量与配额）

- **现状证据**：下单到支付是虚拟的（`app/alioth/internal/biz/order/order.go:21`、`:214-216`；`app/alioth/internal/data/namespace_order_repo.go:91-94`；`api/alioth/v1/identity.proto:171-176` 注释「当前没有对接支付渠道」）；只有人数上限快照（`app/alioth/internal/data/ent/schema/feature_package.go:35-37`、`app/alioth/internal/data/ent/schema/namespace_order.go:39-40`），没有用量计量与配额扣减；全仓没有发票或账单模型。
- **缺口**：支付渠道与回调验签、订单对账/退款/优惠、发票（抬头、税号、状态）、用量计量（成员数、探测目标数、告警量）与超限策略、账单与用量页。
- **要动的层**：proto 与 ent（支付流水、发票、用量）、biz（订单状态机扩展、配额检查）、data（渠道适配器）、web 页面、部署回调地址。
- **价值**：从内部演示套餐走到真实收费；需要运营与财务流程一起落地。
- **依赖**：M20/M21 订单与套餐；第 6 项的容量口径。
- **规模**：L。
- **优先级**：P2（业务触发再启动）。
- **风险**：金额与币种精度、幂等与对账、发票合规；真实支付失败会阻塞租户创建，需要补偿与人工兜底。

## 8. Jaeger（TRACE_JAEGER）规则评测

- **现状证据**：Jaeger 数据源可创建、查询可代理（`app/alioth/internal/biz/datasource/datasource.go:516`），但评测只认数值型驱动，Jaeger 不在其中而被跳过（`app/megrez/internal/biz/eval/judge.go:284-285`、`app/megrez/internal/biz/eval/worker.go:261`）；测试把它列为未实现驱动（`app/megrez/internal/biz/eval/eval_test.go:203`）；设计文档写明本期不做（`docs/design.md:35`）。
- **缺口**：trace 到数值序列的口径（按服务/操作/错误率/延迟分位）、查询接口与窗口语义、采样偏差处理、事件标签映射、Web 查询编辑器与图表。
- **要动的层**：`megrez` judge/worker 与 data 客户端、`alioth` 回写与事件标签、proto/web 编辑器与图表、docs。
- **价值**：补齐第三类数据源；链路故障可告警。
- **依赖**：M3 评测链路、M10/M11/M12 的数值化先例。
- **规模**：M/L。
- **优先级**：P2。
- **风险**：trace 没有天然数值序列，口径要自定并长期维护；采样率会让阈值不稳，误报成本高于日志规则。

## 9. 小步速赢

- **现状证据**：审计页写死 `kind=EXECUTE`（`web/src/pages/ssh-command/audits.tsx:88`、`:146`），而 proto 已支持全部 kind（`api/alioth/v1/ssh.proto:283-285`）；成员页角色选择器固定 `page=1,pageSize=100`，代码注释自己承认上限（`web/src/pages/member/index.tsx:110-111`）；外观偏好只存 localStorage（`web/src/lib/appearance.ts:27-36`），ent 里没有用户偏好表；路由里没有值班大屏（`web/src/App.tsx:49-119`），概览页现在承担值班入口（`README.md:37`）；OAuth 的 state 与跳转现场由前端生成并存 sessionStorage（`web/src/auth/oauth-redirect.ts:75-91`，proto 注释也写前端生成 `api/alioth/v1/identity.proto:724-725`）；`make api` 只跑 buf generate，没有 OpenAPI 产物（`Makefile:47-48`）；README 只有中文，无 `README.en.md`；Kubernetes 只有 PV 的 `storageClassName` 字段（`api/alioth/v1/kubernetes.proto:670`、`:716`），没有 HPA/NetworkPolicy 资源与能力码。
- **缺口与规模**：① 审计页放开全部 kind —— S；② 角色选择器改分页搜索（`web/src/components/searchable-select.tsx` 已有）—— S；③ 外观偏好服务端持久化（新表 + `/self/preferences` + Web 回填）—— S/M；④ 值班大屏 wallboard（只读、免登录需单独鉴权设计）—— S/M；⑤ OAuth 绑定 state 服务端化（防 CSRF、支持跨设备回调）—— S；⑥ OpenAPI 产物随 buf 生成并发布 —— S；⑦ 英文 README —— S；⑧ K8s HPA / StorageClass / NetworkPolicy（能力码 + 权限 + 只读页 + 操作）—— M/项。
- **要动的层**：以 web 为主；③④⑤ 涉及 proto + biz，⑧ 涉及 proto + 能力目录 + 权限 + web 页面。
- **价值**：低成本补可见缺陷，见效快；⑧ 其实是能力扩展，不算速赢。
- **依赖**：彼此独立，适合打包成一个批次；②③⑤ 一起改可以减少一次回归。
- **规模**：①–⑦ 合计 M；⑧ 另计 M。
- **优先级**：①–⑦ P1，⑧ P2。
- **风险**：大屏免登录与租户隔离冲突；偏好服务端化要兼容旧 localStorage；OpenAPI 会带来「产物与 proto 不同步」的新维护点。

## 10. 明确不做 / 保持现状的决策项

| 决策项 | 现状依据 | 结论 | 复核触发条件 |
| --- | --- | --- | --- |
| 入站 Webhook（Prometheus/Alertmanager 推送告警） | `docs/design.md:574` 明确不迁入 `/v1/alerts/prometheus/webhook`，评测结果走内部回写 | 保持不做 | 需要接外部告警源时，另设适配层而不是改评测链路 |
| 独立探测任务模型（玉树式 `probe_task`） | `docs/design.md:30`、`docs/iterations/conventions.md:242-244` | 保持不做，探测只作为告警规则的一种类型 | 需要独立探测报表或非告警探测时重评，与第 2 项一起考虑 |
| 微前端 | `docs/design.md:31`、`:490` | 保持单个 Vite 应用 | 单应用体积或团队边界成为瓶颈时重评 |
| 每个服务一块业务库 | `docs/design.md:32`、`:56`，业务表只在 `alioth` 的 ent schema | 保持单业务库 + 按租户拆历史库 | 单库写入成为瓶颈时，先做分区与归档再谈拆库 |
| 把 Redis 当缓存、会话或消息队列 | `docs/design.md:33`、`docs/iterations/conventions.md:190` | 保持只做分布式锁 | 只有锁语义不够时先评估，不顺手改成队列 |
| 密码登录 | `docs/design.md:244-245` 已移除，`password_hash` 为历史列 | 保持只走邮箱验证码与 OAuth2 | 除非有明确合规需求，不恢复 |
| 短信发送 | `docs/roadmap.md:47`、`docs/design.md:34` 原为「不做」 | **决策已变更**：转入本文第 1 项候选 | 腾讯云去留先做产品决策 |

注意第 1 项与这里的关系：短信从「本期不做」变成「待做候选」，后续引用时以本文第 1 项为准，不要再引 roadmap 的旧结论。

## 建议顺序

1. **第 0 批（并行，无依赖）**：第 4 项里的小件（CSP/HSTS、CI 依赖扫描）、第 9 项的 ①②⑥⑦；同时落第 3 项的 `/metrics` 与请求 ID 骨架。这一批不改业务语义，风险最低。
2. **第 1 批（P0）**：第 5 项 e2e 骨架，先覆盖登录、租户切换、告警处置、通知试发四条主路径；第 4 项的登录限流与会话/刷新令牌吊销。没有 e2e 先落地，后面的页面改动都只能靠人肉回归。
3. **第 2 批（P1）**：第 6 项数据保留与容量治理，先做分区与历史周表，再补 `probe_results` 的每目标上限；随后做第 1 项阿里云短信（腾讯云同时给产品决策结论）。
4. **第 3 批（P1/P2）**：第 9 项剩余（③④⑤⑧）；第 3 项的抓取配置与示例面板随 M32 大盘一起交付。
5. **第 4 批（P2）**：第 7 项商业化、第 8 项 Jaeger、第 2 项边缘探测，按业务触发排序；三项都需要独立里程碑和验收口径，不塞进现有迭代。
