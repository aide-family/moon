# M21 套餐类型细化与租户套餐运维

M21 把 M19/M20 的「一个模块一个能力」细化到**资源类型级**：数据源按驱动、探测按驱动、通知渠道按渠道类型拆分能力。未授权的类型不出现在列表、选择器和规则里；通知渠道类型未授权时，对应模板也不再出现。同时给租户补上**套餐升级 / 续期**，并把平台按租户查订单从弹窗改成独立路由页。

依赖：M19（能力模型）、M20（套餐定价与租户订单）。不改变权限名与能力检查的执行顺序，只在能力目录里增加类型维度，并让资源读写按类型做二次检查。

## 能力目录变化

能力总数由 30 增加到 37；基础套餐由 15 项增加到 22 项。新增类型能力都复用所在模块的权限：

| 模块 | 能力编码 | 名称 | 对应资源类型 |
| --- | --- | --- | --- |
| datasource | `datasource_prometheus` | Prometheus 指标 | `METRICS_PROMETHEUS` |
| datasource | `datasource_victoria_metrics` | VictoriaMetrics 指标 | `METRICS_VICTORIA_METRICS` |
| datasource | `datasource_elasticsearch` | Elasticsearch 日志 | `LOGS_ELASTICSEARCH` |
| datasource | `datasource_loki` | Loki 日志 | `LOGS_LOKI` |
| datasource | `datasource_vlogs` | VictoriaLogs 日志 | `LOGS_VLOGS` |
| datasource | `datasource_jaeger` | Jaeger 链路 | `TRACE_JAEGER` |
| probe | `probe_http` | HTTP 探测 | `PROBE_HTTP` |
| notification | `notify_channel_smtp` | 邮件渠道 | `SMTP` |
| notification | `notify_channel_webhook` | Webhook 渠道 | `WEBHOOK` |
| notification | `notify_channel_sms` | 短信渠道 | `SMS` |

旧的 `datasource`、`probe_target`、`notify_channel` 能力码不再出现在目录里。模板仍是 `notify_template` 能力，但模板类型由渠道类型授权决定：

| 渠道类型 | 对应模板类型 |
| --- | --- |
| `SMTP` | `EMAIL` |
| `WEBHOOK` | `WEBHOOK` |
| `SMS` | `SMS_ALICLOUD` |

配套查询：`DatasourceCapabilityOf/DatasourceDriversOf`、`ProbeCapabilityOf/ProbeDriversOf`、`ChannelCapabilityOf/ChannelTypesOf`、`ChannelTypeOfTemplate`、`TemplateTypesOfChannelTypes`。前端在 `web/src/lib/feature-catalog.ts` 维护同一份映射，并新增 `authorizedDatasourceDrivers`、`authorizedProbeDrivers`、`authorizedChannelTypes`、`authorizedTemplateTypes`。

## 类型级鉴权

`access.Checker` 增加两个方法：

- `Capabilities(ctx)`：当前租户当前生效的能力，供列表按类型过滤；
- `RequireCapability(ctx, capability)`：模块权限通过后的类型级检查，未开通返回 403 `feature not enabled`。

执行点：

1. **数据源**：创建 / 修改校验新驱动和原驱动的能力；详情、状态、删除、连通检查、指标 / 日志查询都按 `Driver` 二次校验；列表和 `Select` 只返回已开通驱动。未授权数据源不能通过已知 UID 访问。
2. **告警规则**：规则驱动映射到数据源或探测能力，新建 / 修改校验新驱动和原驱动；详情、启停、删除同样校验；列表按已开通驱动过滤（数据源驱动 + `PROBE_HTTP`）。未授权类型的规则不出现在规则页。
3. **探测目标**：目标组按 `PROBE_HTTP` 能力校验创建、修改、启停、删除、立即探测和探测记录读取；列表按已开通探测驱动过滤。一期只有 HTTP，TCP / ICMP 实现执行器后再追加能力。
4. **通知渠道**：创建 / 修改 / 详情 / 启停 / 删除 / 测试都校验渠道类型；列表与选择器按已开通渠道类型过滤。
5. **消息模板**：类型映射到渠道类型后校验创建 / 修改 / 详情 / 启停 / 删除 / 预览 / 测试；列表、选择器和系统预设按可用的模板类型过滤。
6. **通知组**：保存时拒绝绑定未授权的渠道或模板；读取时按「渠道类型 + 模板类型」成对过滤 `channel_uids/template_uids`，未授权渠道连同它绑定的模板一起不返回。历史配置保留在库里，重新授权后恢复展示。
7. **菜单入口**：数据源、探测和通知渠道的菜单键改为模块码，`session.canFeatureEntry` 在模块下任一能力开通且持有对应 `menu:<能力>` 时显示入口。

**控制面与数据面边界**：类型级能力只收紧管理面（列表、创建、修改、删除、选择器）。套餐降级后，已经创建的数据源 / 规则 / 通知组继续按原配置参与评测和投递，不会因为页面隐藏而中断；重新授权后重新可见。与 M19「能力只影响接口访问，不影响已落库数据」的取舍一致。

## 租户套餐升级与续期

订单新增 `kind`：`CREATE` / `UPGRADE` / `RENEW`。

- **升级**：当前租户选择另一个公开且启用的套餐下单。支付成功后切换 `namespaces.package_uid`，按新套餐时长从支付时刻重新计算到期时间；租户额外能力保留，新套餐能力并上额外能力后补齐预置角色权限。若目标套餐与当前套餐相同，拒绝下单。
- **续期**：按当前套餐再买一个周期。支付后从「当前到期时间（未过期时）或支付时刻」累加套餐时长；永久套餐（`duration_days = 0`）不能续期。套餐名称不变。
- 两种订单都要求 `namespace:write`，由下单人本人支付；价格、时长、人数上限仍是下单快照。
- 支付仍在 `namespaceOrderRepo.PayOrder` 的同一个事务里完成：更新租户套餐 / 到期时间、补齐预置角色权限、把订单置为 `PAID`。重复支付返回原结果，不会重复延长。

新增接口：

| 方法与路径 | 权限 | 说明 |
| --- | --- | --- |
| `POST /v1/namespace/orders/upgrade` | `namespace:write` | `{packageUid}`，创建升级订单 |
| `POST /v1/namespace/orders/renewal` | `namespace:write` | 创建续期订单（当前套餐） |
| `GET /v1/namespace/{uid}/package` | `namespace:read` | 当前套餐、价格、人数上限、到期时间 |
| `POST /v1/orders/{uid}/pay` | 下单人 | 按 `kind` 执行创建 / 升级 / 续期 |

前端在租户页展示「当前套餐」卡片（名称、到期时间、人数上限、价格），提供升级 / 续费按钮；升级 / 续费确认后一次完成「下单 + 虚拟支付」，支付失败时订单保留在订单页可稍后支付。

## 平台订单查询改为路由

`/tenants/:uid/orders` 取代原来的订单弹窗：

- 从租户管理页的行操作进入，带上租户名称；直接打开 URL 时退回显示 uid；
- 订单按创建时间倒序分页，展示订单号、套餐、类型（创建 / 升级 / 续期）、金额、状态、创建与支付时间；
- 页面提供「返回上一页」，固定回到 `/tenants`；该路径和租户管理页一样不要求当前租户。

## 迁移与 seed

- 权限目录新增 10 个能力和对应菜单权限；seed 先执行类型能力展开，再跑既有的模块级展开与角色补齐：
  1. `feature_package_items` / `namespace_features` 里的旧 `datasource` 展开成 6 个数据源能力，`probe_target` 展开成 `probe_http`，`notify_channel` 展开成 3 个渠道类型能力，并删除旧行；
  2. 删除目录里不存在的旧菜单权限（`menu:datasource`、`menu:probe_target`、`menu:notify_channel`）及其角色授权；
  3. 预置角色与自定义角色按新能力补菜单权限（既有流程）。
- 基础套餐包含全部 22 个基础能力，保持升级前「基础套餐可用全部数据源与通知渠道类型」的行为。
- `namespace_orders` 新增 `kind`，存量订单取默认值 `CREATE`。

## 验收

1. 套餐勾选里能看到 37 个能力，数据源 6 项、探测 1 项、通知渠道 3 项；基础套餐为 22 项。
2. 只开通 `datasource_prometheus` 时：数据源列表只出现 Prometheus 数据源，驱动下拉没有其他类型；按已知 UID 读取 / 修改 / 删除其他类型数据源返回 403 `feature not enabled`。
3. 只开通 `datasource_loki` 时：告警规则页面的驱动选项只有 Loki，列表不出现 Prometheus / 探测规则；新建其他驱动的规则被拒绝。
4. 未开通 `probe_http` 时：探测目标菜单不显示，目标组列表为空，探测规则不出现在规则页。
5. 只开通 `notify_channel_smtp` 时：渠道页只出现 SMTP，模板页只出现 `EMAIL` 类型，通知组只能绑定邮件渠道与邮件模板；组详情不返回 Webhook / 短信的模板配对。
6. 升级订单支付后租户套餐变化、到期时间按新套餐重算、订单类型为 `UPGRADE`；升级不会创建新租户。
7. 续期订单支付后到期时间在原到期时间上增加一个套餐周期；永久套餐不出现续费入口，服务端拒绝续期。
8. 平台租户管理进入 `/tenants/{uid}/orders` 能看到该租户的创建 / 升级 / 续期订单，支持分页和返回租户列表。
9. `cd app/alioth && go build ./... && go test ./...`、真库集成测试（订单创建 / 支付 / 升级 / 续期 / 人数上限）、`gofmt -l`、`scripts/check-naming.sh`、`buf lint`、`make wire`、web 的 typecheck / lint / build 全部通过。
