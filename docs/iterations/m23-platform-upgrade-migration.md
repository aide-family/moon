# M23 升级与迁移

M23 补上平台升级与搬迁服务器时缺的一环：超级管理员在管理台完成「查看状态 → 导出迁移包 → 新环境导入 → 补跑数据库迁移」，不再手工拼 SQL、拷库或临时写脚本。

依赖：M22（租户权益快照与菜单化授权）。

## 1. 页面与入口

新增路由 `/system/upgrade`，侧边栏「升级与迁移」。整页只对平台角色开放：

- `system:migration:read`：查看环境状态与迁移步骤说明，`SYSTEM_READONLY` 和 `SYSTEM_ADMIN` 也能看；
- `system:migration:write`：执行数据库迁移、导出迁移包、导入迁移包，只授予超级管理员。

写操作都要在服务端再校验一次权限，页面隐藏按钮不能代替服务端拒绝。页面分四块：

1. **环境状态**：控制面版本、迁移包格式版本、PostgreSQL 版本、配置表数量与行数、加密密钥指纹。数据全部实时读取，不落库。
2. **导出迁移包**：点一次生成并下载 `.alioth-backup`（gzip 后的 tar）。导出在服务端流式完成，不写临时文件。
3. **导入迁移包**：上传文件后先校验格式、版本、密钥指纹和表集合，再覆盖配置数据并补跑 seed；成功后当前会话失效，必须用迁移包里原有的账号重新登录。
4. **执行数据库迁移**：等价于 `alioth migrate`，可重复执行；用于新环境初始化、升级后补跑和迁移 Job 失败重试。

页面底部列出升级步骤：旧环境导出 → 部署新版本并执行数据库迁移 → 导入迁移包 → 用旧账号登录校验 → 切流量 → 旧环境至少保留一个回滚窗口。

## 2. 迁移包格式

文件是 gzip 压缩的 NDJSON 流，扩展名 `.alioth-backup`，每行一个 JSON 对象，共三种 `kind`：

```
{"kind":"manifest","manifest":{...}}        第一行，固定为清单
{"kind":"row","table":"users","row":{...}}  每行一条记录，按清单里的表顺序排列
{"kind":"summary","summary":{"tables":[...]}} 最后一行，记录每张表实际导出的行数
```

`manifest` 字段：

| 字段 | 说明 |
| --- | --- |
| `formatVersion` | 迁移包格式版本，本期为 `1`；不认识的值直接拒绝导入 |
| `appVersion` | 导出时的控制面版本，仅用于展示与排查，不参与兼容判断 |
| `createdAt` | 导出时间，RFC3339 |
| `encryptKeyCheck` | `ENCRYPT_KEY` 的 SHA-256 前 16 位十六进制，用于判断目标环境能否解密历史密文 |
| `tables` | 表清单，按导出与导入顺序排列，每项含 `name` 与 `columns` |

`row` 的 `row` 字段是 `row_to_json` 的输出，列顺序与清单里的 `columns` 一致。选择 NDJSON 而不是 tar/zip：gzip 能流式写、流式读，压缩包不需要预先知道每张表的字节数，也不会为导出在服务端落临时文件；`summary` 在最后一行，上传被截断时导入能直接发现。

### 包含的表

身份与租户：`users`、`user_oauths`、`namespaces`、`members`、`member_invites`、`member_roles`、`roles`、`role_permissions`、`permissions`、`system_messages`、`user_alert_pages`。

套餐与订单：`feature_packages`、`feature_package_items`、`namespace_features`、`namespace_orders`。

告警配置：`datasources`、`alert_dictionaries`、`alert_rules`、`alert_rule_groups`、`alert_rule_levels`、`alert_pages`、`alert_event_columns`、`probe_target_groups`、`probe_targets`。

通知配置：`notification_channels`、`notification_templates`、`notification_presets`、`notification_groups`、`notification_on_calls`、`notification_on_call_groups`、`notification_subscriptions`、`notification_subscription_groups`。

边缘：`clusters`、`agents`、`ssh_commands`、`ssh_command_audits`。

共 36 张表。表清单在数据层集中维护，并有单测对照 ent 的 `migrate.Tables`：之后新增 ent 表时，要么把它加进迁移包，要么显式加进运行数据排除名单，不允许漏掉。

### 不包含的表

运行数据 11 张：`alert_events`、`alert_event_firings`、`alert_eval_batches`、`alert_history_events`、`notification_tasks`、`message_logs`、`message_log_shards`、`probe_results`、`captchas`、`email_login_codes`、`worker_instances`。

理由：这些是队列、实时事件、验证码和进程心跳，跨环境搬运没有业务意义；历史告警和消息日志还落在按租户拆分的独立库（周表）里，体量远大于配置，本期不迁。页面和文档都要写明「迁移包只含配置与身份数据」，不能让人误以为实时告警历史也一起搬走了。

## 3. 导出

- 逐个表执行 `SELECT row_to_json(t)::text FROM (SELECT * FROM <表>) t`，把结果逐行写进 gzip 流，全程流式输出，不落临时文件，也不把整表读进内存。
- 导出前先按 `information_schema.columns` 读出每张表的列清单写进 manifest；导入时用同一份列清单做校验和插入，所以升级后新增的列走目标库默认值。
- 导出结束写 `summary`，记录每张表实际导出的行数；`gzip` 自身的校验和加上 `summary` 的行数核对，保证截断或损坏的包不会被悄悄导入。
- 导出不修改任何数据；并发写期间导出的快照不保证跨表一致，页面提示在维护窗口内导出。导出期间持有迁移锁，同一个部署同一时间只能有一个迁移包操作。

## 4. 导入

导入在一个事务里完成，顺序固定：

1. 校验 `formatVersion`、`encryptKeyCheck`（不一致直接拒绝，历史密文换密钥就无法解密）、表集合与列集合。
2. 校验请求体大小，超过 256 MiB 拒绝。
3. 对迁移包覆盖的 36 张配置表执行 `TRUNCATE ... CASCADE`，并清空 11 张运行数据表（它们不进包，保留下来只会引用已不存在的配置）。
4. 按 manifest 顺序逐表插入（每批 500 行，`INSERT INTO <表> (<列>) SELECT <列> FROM json_populate_recordset(NULL::<表>, $1::json)`）。
5. 读到 `summary` 后核对每张表实际行数与摘要一致；缺 `summary` 视为文件不完整，整笔回滚。
6. 重置每张表 `id` 列的序列，避免导入后新建记录主键冲突。
7. 提交事务后执行 `seed`，补齐新版本新增的权限、预置模板、基础套餐和租户权益快照。

导入会清掉目标库的运行数据（第 3 步显式 `TRUNCATE`），因此只允许在新环境或维护窗口执行；接口返回 `reloginRequired`，前端清会话并跳登录页。历史告警与消息日志的租户分库不在迁移包里，也不会被导入清掉。

密钥不一致、表集合不认识、存档截断或列缺失时整笔回滚，目标库保持导入前的样子。

## 5. 执行数据库迁移

执行 `ent Schema.Create`、seed 和周分区，与 `alioth migrate` 同一套代码；幂等，可重复点。三个写操作（迁移、导出、导入）都用同一个 Redis 锁 `alioth:lock:system:migration` 串行，多副本同时点只有一个执行，其余返回「正在执行」。锁 TTL 取 30 分钟，操作结束主动释放。

## 6. 接口

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/v1/system/maintenance` | `system:migration:read` | 环境状态 |
| POST | `/v1/system/maintenance/migrate` | `system:migration:write` | 执行数据库迁移 |
| GET | `/v1/system/maintenance/archive` | `system:migration:write` | 下载迁移包（流式，`Authorization: Bearer`） |
| POST | `/v1/system/maintenance/archive` | `system:migration:write` | 上传迁移包（流式，请求体就是文件内容） |

两个迁移包接口是裸 HTTP 处理器：它们不进 proto 的普通 JSON 约定，因为要流式收发二进制。鉴权复用同一个 JWT，错误仍按 `code / reason / message` 的 JSON 返回，前端按现有 `http.ts` 的规则提示。

## 7. 权限与 seed

`app/alioth/pkg/perm/perm.go` 的系统权限目录新增：

- `system:migration:read`
- `system:migration:write`

两个都只出现在平台权限目录，不写进租户 `permissions` 表，也不参与角色授权。超级管理员由 `SystemNames()` 自动获得全部；`SYSTEM_ADMIN` 显式加 `system:migration:read`；`SYSTEM_READONLY` 由 `systemReadPermissions()` 自动获得 read。

前端 `web/src/lib/permission-catalog.ts` 同步两份权限文案，侧边栏按 `system:migration:read` 显示入口。

## 8. 升级与回滚说明

`deploy/k8s/README.md` 补一节「版本升级与迁移」，写明：

1. 升级前在旧环境导出迁移包，并记录 `ENCRYPT_KEY`（新环境必须使用同一份，否则数据源密码、Webhook 密钥、SSH 凭据无法解密）。
2. 新环境部署完成后先执行数据库迁移（管理台按钮或 `alioth-migrate` Job 二选一）。
3. 导入迁移包，用旧环境账号登录校验租户、数据源、规则和通知配置。
4. 切换流量后旧环境的 PostgreSQL 至少保留一个发布周期，用于回滚。
5. 迁移包不含历史告警与消息日志的分库数据，需要保留历史的继续用旧库查询。

## 9. 验收

1. 超管在 `/system/upgrade` 能看到控制面版本、PostgreSQL 版本、配置表数量和加密密钥指纹；`SYSTEM_READONLY` 能看状态但不能导出、导入或执行迁移。
2. 导出得到的 `.alioth-backup` 用 `gzip -dc` 能逐行解析出 `manifest`、`row` 与 `summary`；36 张表的行数与库内一致。
3. 在空库上导入迁移包成功：36 张表行数与源库一致，`id` 序列重置后再插入新记录不冲突，seed 补齐新版本权限与套餐。
4. `ENCRYPT_KEY` 与迁移包不一致时导入被拒绝，库内数据不变；伪造 `formatVersion`、截断文件同样被拒绝且回滚。
5. 重复点击「执行数据库迁移」成功且幂等；并发点击时只有一个执行，另一个收到「正在执行」。
6. 导入成功后当前会话失效，重新用迁移包内账号登录可以看到全部租户、数据源、规则和通知配置；实时告警历史为空（本期不迁运行数据）。
7. `go build ./...`、`go test ./...`（alioth）、`gofmt -l`、`scripts/check-naming.sh`、`buf lint`、`make wire`、Web 的 `npm run typecheck`、`npm run lint`、`npm run build` 全部通过。
