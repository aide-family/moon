# M34 巡检与定时巡检

M34 把数据源查询、探测目标与告警汇总成周期报告，按 cron 执行并按通知组投递，支持无登录的只读分享。设计与原型见 `docs/inspection-design.md`（`docs/inspection-plan-prototype` 分支）。

## 交付范围

| 层 | 内容 |
| --- | --- |
| 存储 | `inspection_plans` / `inspection_items` / `inspection_templates` / `inspection_runs` / `inspection_reports`，报告含分享哈希、有效期、密码哈希与访问统计 |
| 调度 | 调度器每 15s 认领到期计划，`next_run_at` 乐观推进；`(plan_uid, scheduled_at)` 唯一索引幂等 |
| 执行 | 每秒 `SKIP LOCKED` 领取一条 run，35 分钟租约；五类检查项（指标/日志/探测/告警/说明）；超时、部分失败、结果 JSON 快照 |
| 渲染 | HTML / Markdown 双格式；白名单函数、10s 超时、2MB 上限、模板缺失回退内置 |
| 投递 | 按通知组展开渠道；`phase=REPORT`、`pre_rendered=true`；摘要 + 报告链接；`ALWAYS` / `ON_ABNORMAL` |
| 分享 | token 只存 SHA-256、7/30 天或永久、bcrypt 密码、停用；公开接口免登录并加 `no-store` / `noindex` 与按 token 限流 |
| Web | 巡检任务四步向导、报告模板、执行记录、报告预览/下载/分享、公开报告页 |

## 接口

- `POST /v1/inspection/plans`、`PUT /v1/inspection/plans/{uid}`、`PATCH/PUT .../status`、`DELETE ...`
- `GET /v1/inspection/plans`、`GET /v1/inspection/plans/{uid}`
- `POST /v1/inspection/plans/{uid}/run`
- `GET /v1/inspection/runs`、`GET /v1/inspection/runs/{uid}`
- `GET /v1/inspection/runs/{uid}/reports`、`GET /v1/inspection/reports/{uid}`
- `POST /v1/inspection/reports/{uid}/share`、`DELETE .../share`、`POST .../resend`
- `GET /v1/public/inspection-reports/{token}`

## 验证

- `gofmt`、`check-naming`、`buf lint`、`go vet`、`go test`（alioth + merak）
- `npm --prefix web run check`（typecheck / eslint / i18n / UI conventions）与 `build`
- 真实 PostgreSQL + alioth + Vite：创建计划 → 手动执行 → HTML/Markdown 报告 → 分享 → 公开访问 → 停用 → 404
- `web/e2e/inspection.spec.ts`：真实后端浏览器路径
- 通知任务落库验证：`phase=REPORT`、`pre_rendered=true`
- 公开页响应头与限流验证：30 次/秒/token，超额 429

## 检查项编辑器 v2（分类型）

检查项不再共用一个 query 文本域，按五类分别配置，详细方案见 `docs/inspection-item-editor-design.md`。

- 指标：PromQL 编辑器、窗口/步长、单位、展示方式、阈值（WARN/FAILED）、空窗口告警
- 日志：按驱动切换 LogQL / LogsQL / ES DSL，级别过滤、条数/排序、命中数阈值
- 探测：目标组、可用率/平均延迟/P95/证书天数、阈值、down 目标告警
- 告警：未恢复/全部、级别、标签匹配、条数、告警级别条件
- 说明：Markdown 编辑与预览
- 统一测试接口 `POST /v1/inspection/test-item`，与正式执行共用 `executeSection`，不落库、不产生告警、不投递
- `params` 按类型严格校验，未知键拒绝；探测/日志/告警的新字段已接入执行器
- 指标/日志查询结果补充 ES、VictoriaLogs 解析；探测表增加 P95 与证书剩余列

**验证**：真实 PostgreSQL + Prometheus + VictoriaLogs 调用 test-item；指标返回序列/指标卡，VLogs 解析出 10 条样本，探测与告警返回表格；Web 编辑器完成指标测试（1810ms）、类型切换、保存检查项与保存整份计划。

## 报告模板 v2（单格式 + 可视化）

- 模板 `format` 只能是 HTML 或 MARKDOWN，一份模板只保留一段正文；系统内置模板已移除，全部由租户自定义。
- 模板正文用 Monaco 代码编辑器维护，保存时按格式 Parse 校验；格式创建后锁定，复制后可切换。
- 计划绑定模板后 `output_formats` 自动收敛为该格式，报告只渲染一种格式。
- 新增 HTML 报告 helper：`sectionHTML`、`metricCards`、`seriesChart`、`tableHTML`、`logsHTML`、`alertsHTML`；报告不再只展示原始数据。
- aaa 租户验收数据：4 个计划、5 次执行、2 个自定义模板、5 份分享、2 条 REPORT 通知任务。

## 已知限制

- 公开限流是进程内近似值，多副本按副本独立计算。
- 定时触发只验证代码路径，完整等待 cron 到点未纳入自动化。
- 天璇实际发送未在本轮验收（只验证任务入队）。
- 报告只存库，90 天清理；对象存储归档后续再做。
