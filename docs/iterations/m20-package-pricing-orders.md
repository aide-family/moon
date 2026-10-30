# M20 套餐定价、公开套餐与租户订单

M20 在 M19 的「功能套餐」上补商业属性：套餐带**人数上限**、**定价**和**公开 / 内测版本**。用户在创建租户时必须先选择公开套餐，经过一次**虚拟支付**才真正创建租户并绑定套餐；订单作为独立记录保留，供用户和平台查询。当前没有对接支付渠道，支付成功由接口调用直接判定，但「创建租户 + 绑定套餐 + 订单置为已支付」必须在同一个事务里完成。

依赖：M1（身份与租户）、M19（功能套餐、租户授权与能力校验）。不改变 M19 的能力模型和租户授权语义，只在套餐上叠加可售属性与购买流程。

## 套餐属性

| 属性 | 字段 | 语义 |
| --- | --- | --- |
| 人数上限 | `feature_packages.max_members` | 租户可容纳的 JOINED 成员数（含负责人）；`0` 表示不限制 |
| 定价 | `feature_packages.price_cents` | 人民币分，非负整数；`0` 表示免费 |
| 版本 | `feature_packages.visibility` | `PUBLIC` 可被用户创建租户时选择；`BETA` 只由平台在租户功能授权里分配 |
| 授权 / 计费周期 | `feature_packages.duration_days` | 沿用 M19：`0` 表示永久；同时作为订单的授权时长快照 |

约束：

- 版本默认 `BETA`，避免新建套餐被误上架；基础套餐由 seed 保证为 `PUBLIC`（平台默认免费套餐）。
- `max_members`、`price_cents` 在 proto 和用例层都做非负校验；不合法的值返回 `invalid max members` / `invalid price`。
- 套餐改价、改人数上限、改版本只影响之后的下单；已下订单保存快照，支付时按快照的授权时长计算到期时间。
- 套餐被租户使用时不能删除（M19 既有规则）；改版本不会影响正在使用它的租户。

## 数据模型

| 实体 | 关键字段 | 说明 |
| --- | --- | --- |
| `FeaturePackage` | 增加 `visibility`、`max_members`、`price_cents` | 见上表 |
| `NamespaceOrder` | `uid`、`user_uid`、`namespace_uid`、`namespace_name`、`namespace_remark`、`package_uid`、`package_name`、`price_cents`、`duration_days`、`max_members`、`status`、`paid_at`、`created_at` | 一笔「先下单、后支付」的租户创建记录 |

`NamespaceOrder` 的语义：

- `status` 取 `PENDING` / `PAID` / `CANCELLED`；`namespace_uid` 只在支付成功后写入，`paid_at` 为空表示未支付。
- 订单没有软删除：取消后仍然保留，便于订单查询解释结果。
- 价格、套餐名、授权时长、人数上限都是下单时的快照，不随套餐后续修改而变化。
- 订单只属于下单人：支付、取消、列表都按 `user_uid` 限定，别人的订单一律按「不存在」处理。

## 成员容量

- 租户容量 = 当前套餐的 `max_members`，按 `namespaces.package_uid` 实时读取，不在租户表冗余；换套餐后容量立即变化。
- 计数口径与 M19 一致：只统计未删除、`status = JOINED` 的成员；待确认邀请和已移除成员不占名额。负责人本身是一名 JOINED 成员，因此上限至少要容纳负责人。
- 邀请成员时先在用例层预检容量（`member limit reached`）；真正防并发越界在确认邀请的事务里：
  1. 事务内先用 `UPDATE namespaces SET updated_at = ...` 锁住租户行（`updated_at` 只用于取行锁）；
  2. 再统计 JOINED 成员数；
  3. 达到上限则回滚并返回 `member limit reached`。
  已加入的成员重复确认邀请不重复占名额。
- 把租户换到人数上限更小的套餐不会移除现有成员；之后的新邀请和确认会被拒绝，直到成员数降到上限以内。
- 套餐已删除或租户未绑定套餐时按 `0`（不限制）处理，避免历史数据把租户锁死。

## 购买流程与虚拟支付

1. **选择公开套餐**：`GET /v1/packages` 返回 `visibility = PUBLIC` 且 `status = ACTIVE` 的套餐，按价格升序。登录即可调用，不要求 `X-Namespace`，刚注册、还没有租户的用户也能走完流程。
2. **下单**：`POST /v1/orders` 校验套餐仍公开可售，把套餐名、价格、授权时长、人数上限快照进 `namespace_orders`，状态 `PENDING`。下单**不创建租户**，避免出现未付款的空租户。
3. **虚拟支付**：`POST /v1/orders/{uid}/pay` 由仓储在一个事务里完成：
   - 校验订单属于当前用户；
   - 支付时重新确认套餐未删除且启用，否则返回 `package is no longer available`；
   - 用订单快照的授权时长创建完整租户：绑定套餐、按套餐能力创建预置角色与权限、写入负责人成员和 owner 角色；
   - 条件更新订单 `PENDING -> PAID` 并写入 `namespace_uid`、`paid_at`；
   - 提交。任一步失败整体回滚。
   - 已支付订单重复支付时原样返回当时创建的租户，不创建第二个租户；并发支付由条件更新保证只有一个成功。
4. **取消**：`POST /v1/orders/{uid}/cancel` 只能取消自己的待支付订单；已支付不能取消。
5. 关闭创建弹窗不会删除待支付订单，用户可以在订单页继续支付或取消。

`POST /v1/namespace` 仍然保留，作为平台内部直接开通的入口：它要求传 `package_uid`，且同样只接受公开且启用的套餐，按套餐当前时长计算到期时间。面向用户的标准路径是「下单 → 支付」。

## 订单查询

| 入口 | 权限 | 说明 |
| --- | --- | --- |
| `GET /v1/orders?page&pageSize&status` | 登录用户 | 只列出自己的订单，按创建时间倒序；`status` 留空表示全部 |
| `GET /v1/system/namespace/{uid}/orders?page&pageSize` | `system:namespace:read` | 平台按租户查询订单，租户管理页的「订单」弹窗使用 |

## 接口

全部为 HTTP JSON，字段名 camelCase（proto 的 `json_name`）。错误文案是英文原文，由前端本地化。

| 方法与路径 | 说明 |
| --- | --- |
| `GET /v1/packages` | 公开且启用的套餐列表 |
| `POST /v1/orders` | 创建待支付订单：`{name, remark, packageUid}` |
| `POST /v1/orders/{uid}/pay` | 虚拟支付并创建租户、绑定套餐 |
| `POST /v1/orders/{uid}/cancel` | 取消自己的待支付订单 |
| `GET /v1/orders` | 当前用户订单分页 |
| `GET /v1/system/namespace/{uid}/orders` | 平台按租户订单分页 |
| `POST /v1/namespace` | 直接创建租户（要求 `packageUid`，只接受公开套餐） |
| `GET/POST/PUT /v1/system/package[/{uid}]` | 套餐 CRUD，增加 `visibility`、`maxMembers`、`priceCents` |

订单视图字段：`uid`、`namespaceUid`、`namespaceName`、`namespaceRemark`、`packageUid`、`packageName`、`priceCents`、`durationDays`、`maxMembers`、`status`、`createdAt`、`paidAt`。时间均为 Unix 秒，`0` 表示无。

## 执行点

1. `namespace.Tenancy.CreateNamespace`：校验套餐公开且启用（`package is not available`），再交给仓储创建。
2. `order.Purchasing`：公开套餐列表、下单快照、支付 / 取消 / 本人订单查询；平台订单查询走 `RequireSystem(system:namespace:read)`。
3. `data.namespaceOrderRepo.PayOrder` / `createNamespaceTx`：事务内创建租户并条件更新订单，租户创建逻辑与直接开通共用同一份实现。
4. `member.Membership.InviteMember` 预检容量；`data.inviteRepo.AcceptInvite` 在事务内锁租户后复核容量。
5. 认证中间件把这些新接口加入 `namespaceOptional`：公开套餐、下单、支付、取消、本人订单和平台订单查询都不读 `X-Namespace`。

## 迁移与 seed

- ent 自动迁移新增 `feature_packages.visibility/max_members/price_cents` 与 `namespace_orders` 表；存量套餐的 `visibility` 取列默认值 `BETA`，人数上限和价格取 `0`（不限制 / 免费）。
- seed 保证基础套餐存在且为 `PUBLIC`；已存在的基础套餐只把版本改为 `PUBLIC`，不覆盖管理员配置的价格、人数上限和时长。
- 基础套餐新建时固定 `PUBLIC`、价格 `0`、人数上限 `0`，保持升级前「新租户默认套餐、人数不限制」的行为。
- 不提供历史订单补录：功能上线前创建的租户没有订单记录。

## 回滚

- 回滚代码不会删除新表和新列；旧版本忽略 `visibility`、`max_members`、`price_cents` 和订单表，租户创建回到「默认基础套餐」的旧行为。
- 若在回滚前已经用非基础套餐创建租户，旧代码仍按 `package_uid` 读取能力，功能不受影响；但用户购买入口会消失。

## Web

- 套餐管理弹窗增加「版本」「人数上限」「价格」；列表和详情展示这三项，基础套餐名称仍不可改。
- 创建租户改为两步：填写名称 / 备注并选择公开套餐（卡片展示价格、时长、人数上限、能力数）→ 创建订单 → 确认支付。支付成功后才设置租户并刷新会话。
- 新增「订单」页：本人订单分页，可按状态筛选；待支付订单可以支付或取消；支付成功刷新会话和列表。
- 平台租户管理新增「订单」行操作，只读弹窗按租户查询订单。
- 成员页标题旁展示「成员已用 / 上限」，容量来自会话里的租户快照。
- 后端新增错误原文补齐中英文映射：`member limit reached`、`package is not available`、`package is no longer available`、订单状态相关错误。

## 验收

1. `FeaturePackage` 新增三个字段可写可读；`maxMembers`、`priceCents` 为负时 400；`visibility` 只接受 `PUBLIC` / `BETA`。
2. 公开套餐列表只返回 `PUBLIC + ACTIVE` 套餐；内测套餐不出现在列表中，但平台功能授权仍然可用它。
3. 未登录调用 `/v1/packages`、`/v1/orders` 返回 401；不带 `X-Namespace` 也能正常下单和支付。
4. 下单后订单为 `PENDING` 且没有创建租户；`GET /v1/orders` 能看到它。
5. 支付后订单为 `PAID`、写入 `namespaceUid` 与 `paidAt`，租户已创建并绑定所选套餐、预置角色权限与套餐能力一致；重复支付同一订单不会创建第二个租户。
6. 取消待支付订单后状态为 `CANCELLED`，租户不存在；已支付订单取消返回 400。
7. 他人订单支付 / 取消返回 404；平台按租户查订单需要 `system:namespace:read`。
8. 套餐 `maxMembers = 2` 时：负责人占 1 个名额，第 2 位成员确认后成员数到 2，再邀请或确认返回 400 `member limit reached`；并发确认不会超过上限；把套餐换成上限更小的套餐不会移除成员。
9. 存量库迁移后基础套餐为 `PUBLIC`，其他套餐为 `BETA`；`max_members = 0` 的存量套餐不影响邀请。
10. `cd app/alioth && go build ./... && go test ./...`、`cd api && go test ./...`、`gofmt -l`、`scripts/check-naming.sh`、`buf lint`、`make wire`、`web` 的 typecheck / lint / build 全部通过。
