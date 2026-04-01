---
name: proto-backend-module
description: Implements backend modules from proto definitions for goddess, marksman, and rabbit apps. Keeps style consistent with existing project structure, reuses magicbox and in-app code, follows Go and project conventions, and requires syncing README (API overview, features, usage) when adding, modifying, or removing modules or APIs. Use when the user says "帮我完成某某功能" or manually @ this skill to implement a module based on proto.
---

# Proto 后端模块实现

根据 proto 定义在 goddess、marksman、rabbit 中完成后端功能时，严格遵循本技能中的分层、命名和复用约定。

## 何时使用

- 用户说「帮我完成某某模块」或「完成 XX 模块」时
- 用户手动 @ 本 skill 并要求基于 proto 实现后端时

## 核心原则

1. **风格与项目一致**：目录、包名、接口命名、错误处理、日志风格必须与现有模块一致。
2. **能复用则复用**：优先使用 magicbox 公共包和当前 app 内已有类型（bo、enum、merr、contextx、safety 等），不重复造轮子。
3. **避免重复逻辑、提高复用率**：同一模块或同一 biz 内若出现两段及以上相同或高度相似的校验/逻辑（如「先查 A 再判类型再往下」），应抽取为私有方法或共享函数（如 `validateXxx`、`ensureYyy`），由多处调用该函数，避免复制粘贴。这样后续修改校验或文案时只需改一处。
4. **不创造突兀结构**：除非用户明确要求，不得新建与现有 app 结构不符的目录或包；新代码必须落在已有分层内。
5. **命名与导入**：遵循 Go 官方规范；import 顺序见下文。

## 项目结构速查

- **Proto**：`proto/<app>/api/v1/*.proto`，生成到 `app/<app>/pkg/api/v1/*.pb.go`（由 buf/kratos 生成，勿手改）。
- **配置**：`internal/conf/conf.proto` 定义 Bootstrap；**运行时配置**（如 `app/<app>/config/server.yaml`）必须与 conf.proto 的字段同步，否则新增字段无法生效。修改 conf.proto 后必须同步修改实际使用的 YAML/配置文件。
- **App 分层**（以 rabbit 为例，goddess/marksman 类似）：
  - `cmd/run/{http,grpc,job,all}/wire.go`：wire 注入
  - `internal/conf`：配置定义与生成
  - `internal/server`：HTTP/gRPC 注册（RegisterHTTPService / RegisterGRPCService / RegisterService）
  - `internal/service`：对接 proto 生成的 Server 接口，调 biz
  - `internal/biz`：业务逻辑；`biz/repository` 接口；`biz/bo` 请求/响应 BO
  - `internal/data`：`data/impl` 实现 repository；`impl/do` GORM 模型；`impl/convert` do↔bo 转换；`impl/query` 为 gorm gen 生成
- **magicbox**：公共能力（enum、merr、contextx、safety、strutil、encoding、hello 等），在 `magicbox/` 下按包使用。

实现新模块时，只在这些既有分层中新增或修改文件，不要新增顶层包或与现有命名风格不符的目录。

## 实现流程（按顺序）

1. **确认 proto 与生成代码**  
   - Proto 在 `proto/<app>/api/v1/`，`go_package` 指向 `github.com/aide-family/<app>/pkg/api/v1`。  
   - 确保已生成 `pkg/api/v1/*.pb.go`（如未生成，提醒用户执行项目既定 make/buf 命令）。  
   - **HTTP 路由勿与已有路径冲突**：若已有 `GET /v1/resource/{uid}` 这类带路径参数的 route，不要新增 `GET /v1/resource/xxx`（会被匹配成 `uid=xxx`，导致解析错误）。应使用不同前缀或层级，例如 `GET /v1/resources/xxx`（复数前缀）、`GET /v1/resource/action/xxx`（多段），或单独资源名如 `GET /v1/namespaces/simple`。新增/修改 proto 的 `google.api.http` 后需执行 `make api` 并确认生成的路由表无冲突。
   - **DELETE 不得声明 body**：HTTP DELETE 方法不应携带 request body（符合 RFC 与 OpenAPI 惯例，否则 `make api`/OpenAPI 生成会告警）。若需传递「要解除绑定的 ID 列表」等参数，**不要**使用 `body: "*"`，应去掉 body，让 Request 中除路径变量外的字段通过 **query 重复参数** 传递（例如 `DELETE /v1/notification-groups/{uid}/webhooks?webhook_uids=1&webhook_uids=2`）。路径变量（如 `{uid}`）照常写在 path 中。

2. **biz/bo**  
   - 在 `internal/biz/bo/` 下为当前模块增加类型（如 `CreateXxxBo`、`UpdateXxxBo`、`XxxItemBo`、`ListXxxBo` 等），与 proto 的 Request/Reply/Item 对应。  
   - 提供从 `*apiv1.*Request` 构造 BO 的构造函数（如 `NewCreateXxxBo(req *apiv1.CreateXxxRequest)`）。  
   - 提供 BO → proto 响应的转换（如 `ToAPIV1XxxItem`、`ToAPIV1ListXxxReply`）。  
   - 分页复用现有 `PageRequestBo` / `PageResponseBo[T]`（见 `biz/bo/page.go`），不要新建分页类型。

3. **biz/repository**  
   - 在 `internal/biz/repository/` 下新增接口（如 `XxxConfig`），方法签名使用 `context.Context`、`*bo.*Bo`、`snowflake.ID` 等，返回业务所需类型（含 `*bo.PageResponseBo[*bo.XxxItemBo]` 等）。  
   - 与 proto 的 RPC 一一对应，但接口命名保持「领域+动作」（如 `CreateXxxConfig`、`ListXxxConfig`）。  
   - **单一职责**：每个 repository 方法只做「单表 / 单次查询 / 单一语义」的一件事；不在一个方法内同时查多张表或组合多种判断。若业务需要「A 表或 B 表存在则如何」，应在 biz 层多次调用 repo（如 `HasXxxData`、`HasYyyData`），由 biz 组合逻辑，repo 只提供原子能力。

4. **data/impl/do**  
   - 在 `internal/data/impl/do/` 下新增 GORM 模型（嵌入 `BaseModel`，使用 `snowflake.ID`、`gorm` 标签、`TableName()`）。  
   - 敏感字段使用 magicbox 类型（如 `strutil.EncryptString`、`safety.Map`）。  
   - 将新 model 加入 `do.Models()` 的返回切片，以便迁移与 gen。

5. **data/impl/convert**  
   - 在 `internal/data/impl/convert/` 下实现 do ↔ bo 的转换（如 `ToXxxConfigDO`、`ToXxxConfigItemBo`），使用 `contextx.GetNamespace(ctx)`、`contextx.GetUserUID(ctx)` 等填充创建人/命名空间。

6. **data/impl/query（gorm gen）**  
   - 项目使用 gorm.io/gen 生成 query。新增 do 后需重新生成 query（见项目 Makefile/cmd），并在 impl 层用 `query.Xxx.WithContext(ctx).Where(...)` 等调用。

7. **data/impl**  
   - 在 `internal/data/impl/` 下新增 `xxx.go`，实现 `repository.XxxConfig`：  
     - 构造函数 `NewXxxConfigRepository(d *data.Data) repository.XxxConfig`，内部调用 `query.SetDefault(d.DB())`（若该 app 使用 SetDefault）。  
     - 各方法用 `query.Xxx`、`convert.*`、`contextx.GetNamespace(ctx)` 等实现，错误用 `merr.ErrorNotFound` / `merr.ErrorInvalidArgument` 等（与现有 impl 一致）。  
     - **单一职责**：每个 impl 方法只操作一张表或只做一次查询/写入；不在同一方法内混用多张表或多种语义。多表组合逻辑放在 biz 层通过多次调用 repo 完成。
   - 在 `impl/provider_set.go` 的 `ProviderSetImpl` 中注册 `NewXxxConfigRepository`。
   - **必须使用 gen 模式**：所有 DB 条件与排序必须通过 gen 生成的 field 表达式（如 `u.UserUID.Eq(x)`、`u.SortOrder.Desc()`），**禁止**手写字符串条件（如 `Where("user_uid = ?", ...)`、`Order("sort_order ASC")`）。
   - **批量写入禁止循环写库**：需要插入多条时使用 `query.Xxx.WithContext(ctx).CreateInBatches(rows, len(rows))` 或一次 `Create(rows...)`，**禁止**在 for 循环内逐条 `Create`。

8. **internal/biz**  
   - 新增 `xxx_config.go`（或与模块同名的 biz 文件），实现 `NewXxxConfig(repo repository.XxxConfig, helper *klog.Helper) *XxxConfig`，方法内调 repo、用 `merr` 和 `helper.Errorw` 处理错误与日志。  
   - 在 `biz/provider_set.go` 的 `ProviderSetBiz` 中注册 `NewXxxConfig`。

9. **internal/service**  
   - 新增 `xxx.go`，实现 proto 生成的 `XxxServer` 接口（嵌入 `apiv1.UnimplementedXxxServer`），各 RPC 方法：将 `*apiv1.*Request` 转为 bo、调 biz、将 bo 转为 `*apiv1.*Reply`。  
   - 在 `service/provider_set.go` 的 `ProviderSetService` 中注册 `NewXxxService`。

10. **internal/server**  
    - 在 `RegisterHTTPService` / `RegisterGRPCService`（以及如需全量的 `RegisterService`）中增加对新 service 的依赖参数，并调用 `apiv1.RegisterXxxHTTPServer` / `apiv1.RegisterXxxServer`。  
    - 若该服务有「允许未登录/未命名空间」的接口，在 `namespaceAllowList` 或 `authAllowList` 中加上对应 Operation 常量。  
    - **跨领域 allowlist**：若本应用**引用并注册了其他领域**（如 marksman 注册 goddess 的 namespace/auth、rabbit 的 webhook/template/sender），必须审查这些被引用领域暴露的 Operation；凡需「免登录」或「免选命名空间」才能访问的接口，须将对应 Operation 常量加入**使用方应用**的 `authAllowList` 或 `namespaceAllowList`，否则中间件会拦截导致接口不可用。

11. **wire**  
    - 若 wire 为自动生成，运行 `wire ./cmd/run/...` 等以更新注入；否则在对应 `wire.go` 中确保 server、service、biz、impl、data 的 ProviderSet 已包含新模块。

12. **配置与 conf.proto 同步**  
    - 若修改了 `internal/conf/conf.proto`（如在 Bootstrap 中新增、删除或重命名字段），必须**同步修改**该应用实际使用的**运行时配置文件**（如 `app/<app>/config/server.yaml` 或 `config/*.yaml`），为新增字段补充对应配置项，否则启动时可能缺省或报错。  
    - 配置项命名与 conf.proto 中字段的 camelCase 一致（如 `selfDomain`、`userDomain`）；领域模块配置统一使用「领域名称+Domain」（如 `namespaceDomain`、`webhookDomain`）；**认证/登录领域**使用 `authDomain`（不用 loginDomain，以保持命名专业性）；结构需与 magicbox 或现有同类型配置一致（如 DomainConfig 的 driver、version、options）。

13. **README 与文档同步**  
    - 对模块或 API 做**新增、修改、删除**时，必须同步更新对应应用的 README，保证「接口说明、功能说明、常用用法」与代码一致。  
    - **必须同时更新**：① **功能特性（Features）**：在功能列表中补充/修改/删除该模块或能力的一句话描述；② **接口概览（API Overview）表**：在表格中补充/修改/删除对应服务与 HTTP 路径、说明；③ **中英文双版本**：同一变更需同时改 `README.md` 与 `README-zh_CN.md`，结构和表格一一对应。  
    - 详见下方 [README 与文档同步](#readme-与文档同步) 小节。

## 数据层与业务层约定（避免常见纰漏）

以下约定来自实践中的中途修正，实现时请直接遵守，避免返工。

### 1. 排序字段语义（越大越靠前）

若业务约定「SortOrder 数值越大越靠前」（列表第一项展示时排最前）：

- **存库**：列表第一项应赋予**最大** SortOrder。推荐用标准库 `slices.Clone` + `slices.Reverse` 反转后再按索引赋 SortOrder（反转后遍历，索引 0 对应原列表最后一项、会得到最小 SortOrder，原列表第一项会得到最大 SortOrder）；或显式 `SortOrder: int32(n-1-i)`。
- **查询**：使用 `Order(field.SortOrder.Desc())`，保证大的先返回。

避免手写「n-1-i」时搞反顺序；用 `slices.Reverse` 时语义更直观。

### 2. 校验：禁止循环单条查询

当需要校验「一批 ID 是否都存在 / 是否都在当前命名空间」时：

- 在 **repository** 增加**批量**方法（如 `CountXxxByUIDs(ctx, uids []snowflake.ID) (int64, error)`），在 impl 内用 `Where(..., ID.In(uidInt64s...)).Count()` **一次**查询。
- 在 **biz** 内比较 `count == len(uids)` 即可，**禁止**在 for 循环内逐条调用 `GetXxx(ctx, id)` 做校验。

### 3. 列表按 ID 批量拉取：禁止循环单条查询

当需要根据一批 UID 拉取详情列表并**保持与输入一致的顺序**时：

- 在 **repository** 增加**批量**方法（如 `GetXxxByUIDs(ctx, uids []snowflake.ID) ([]*bo.XxxItemBo, error)`），在 impl 内用 `Where(..., ID.In(...)).Find()` **一次**查询。
- 在 **biz** 内将返回列表转为 `map[uid]*bo`，再按原始 `uids` 顺序遍历，从 map 中取并 append，得到有序结果；**禁止**在 for 循环内逐条 `GetXxx(ctx, uid)`。

## README 与文档同步

为保证项目质量，**对 proto/模块/功能做任何变更时，必须同步维护 README**，避免文档与实现脱节。

### 何时需要更新运行时配置

| 变更类型 | 需更新的配置 | 更新内容 |
|----------|--------------|----------|
| **conf.proto 新增** Bootstrap 字段 | 该应用实际使用的配置文件（如 `app/<app>/config/server.yaml`） | 在 YAML 中新增与字段名一致的配置项（如 `selfDomain`、`userDomain`），结构参考同类型已有配置（如 `namespaceDomain`） |
| **conf.proto 删除/重命名** 字段 | 同上 | 从配置文件中删除或重命名对应项，避免遗留无效配置 |

**说明**：仅改 conf.proto 并执行 `make conf` 生成 conf.pb.go 不够，运行时加载的是 YAML 等配置文件；若不同步修改配置文件，新字段在运行时为 nil 或零值，依赖该配置的模块可能报错（如 "selfDomain is required"）。

### 何时需要更新 README

| 变更类型 | 需更新的 README | 更新内容 |
|----------|-----------------|----------|
| **新增** RPC / Service / 模块 | 该应用下的 `README.md`**与** `README-zh_CN.md`（二者都改） | **Features**：补充一条新能力描述；**API Overview 表**：补充新接口/新服务；必要时补充常用用法 |
| **修改** RPC 路径、方法名、请求/响应语义 | 同上 | **API Overview 表**：修正 HTTP 路径、方法描述；若影响功能描述则同步更新 **Features** |
| **删除** RPC / Service / 模块 | 同上 | 从 **Features**、**API Overview 表**中移除对应项；若整模块删除则从接口概览整块移除 |
| 新增/删除/重命名**应用**（如新 app） | 仓库根目录 `README.md`、`README-zh_CN.md` | Project Structure 表、Documentation 表、Quick Start 中应用列表 |
| **magicbox** 新增/删除/重命名包或对外能力 | `magicbox/README.md`、`magicbox/README-zh_CN.md` | Module Overview 表、Features；若新增 proto 则更新 Proto 定义表 |

### README 文件位置

- **根目录**：`README.md`、`README-zh_CN.md`（项目结构、各应用入口、文档链接）
- **Goddess**：`app/goddess/README.md`、`app/goddess/README-zh_CN.md`
- **Rabbit**：`app/rabbit/README.md`、`app/rabbit/README-zh_CN.md`
- **Marksman**：`app/marksman/README.md`、`app/marksman/README-zh_CN.md`
- **Magic Box**：`magicbox/README.md`、`magicbox/README-zh_CN.md`

### 更新规范

1. **中英文同步**：同一变更必须同时改 `README.md` 与 `README-zh_CN.md`，结构和表格一一对应，仅语言不同。**只改中文或只改英文视为未完成。**
2. **Features（功能特性）**：  
   - 新增模块/服务：在 Features 列表中增加一条该能力的一句话描述（与 API Overview 中出现的服务对应）。  
   - 修改/删除模块：同步修改或删除 Features 中对应条目。  
   - 避免只改「接口概览」而漏改「功能特性」。
3. **API Overview 表**：  
   - 行内容为「Service / 服务」「Method / HTTP」「Description / 说明」。  
   - 新增 RPC：按现有表格格式增加一行（HTTP 方法 + 路径 + 简短说明）。  
   - 修改：只改对应行，不改变整体表格风格。  
   - 删除：整行删除，保持表格连贯。
4. **不臆造**：README 中的路径、方法名、服务名必须与 **当前 proto 与生成代码** 一致；不确定时查 `proto/<app>/api/v1/*.proto` 与 `google.api.http` 注解。

完成模块实现或 proto 变更后，在自检时必须勾选「README 已同步」，并确认：**Features 已更新、API Overview 已更新、README.md 与 README-zh_CN.md 均已更新。**

## Go 规范

- **Comments, logs, and user-facing text in English only**：All source code must use **professional English** for: comments (line, block, doc), log messages (e.g. `helper.Errorw("msg", "…")`), error strings (e.g. `merr.ErrorNotFound("…")`), and any other user-facing or developer-facing text. No Chinese or mixed language in code. Follow [Effective Go](https://go.dev/doc/effective_go) and standard Go style (concise, clear, consistent). Multi-language UX is handled by frontend/i18n using error codes or keys; the backend exposes stable English messages only.
- **Import 顺序**（严格）：  
  1）标准库  
  2）空白导入（`_ "..."`）  
  3）第三方包（建议先 magicbox，再 kratos、snowflake、gorm、wire 等）  
  4）当前项目包（`github.com/aide-family/<app>/...`）  
  每组之间空一行。
- **命名**：遵循 [Effective Go](https://go.dev/doc/effective_go) 与 Go 官方命名约定；包名简短小写，接口名单词首字母大写，方法名与 proto RPC 对应但用 Go 风格（如 `CreateXxx`）。
- **公共函数/方法**：对外暴露的包级函数或可复用方法需有单元测试（`*_test.go`），与现有 magicbox 和 app 内测试风格一致。

## 复用清单（优先使用，勿重复实现）

- 分页：`biz/bo/page.go` 的 `PageRequestBo`、`PageResponseBo[T]`
- 切片反转/克隆：标准库 `slices.Clone`、`slices.Reverse`（用于按「越大越靠前」赋 SortOrder 等）
- 错误：`magicbox/merr`（如 `ErrorNotFound`、`ErrorInvalidArgument`、`ErrorParams`、`IsNotFound`）
- 上下文：`magicbox/contextx`（如 `GetNamespace`、`GetUserUID`）
- 枚举：`magicbox/enum`（与 proto 一致）
- 安全/敏感：`magicbox/strutil.EncryptString`、`magicbox/safety.Map`
- 日志：`klog.Helper` 的 `Errorw` 等，与现有 biz 一致
- ID：`github.com/bwmarrin/snowflake`，`snowflake.ParseInt64`、`uid.Int64()`

## Goddess 特殊说明

- goddess 存在 `domain/<domain>/v1/` 的领域注册（如 `namespace`、`user`），新领域实现需在对应 `domain/xxx/v1/` 下实现并注册；若只是扩展现有 API，在 `internal/service` 与 `internal/biz` 中按上述流程即可。
- 跨 app 调用：rabbit/marksman 可能依赖 goddess 的 API 客户端（如 `goddessv1 "github.com/aide-family/goddess/pkg/api/v1"`），仅引用已有客户端与接口，不新建 goddess 端未定义的包。

## 检查清单（完成前自检）

- [ ] Proto 已生成到 `pkg/api/v1`，未手改生成代码
- [ ] **conf 与运行时配置已同步**：若修改了 `internal/conf/conf.proto`，已执行 `make conf`（或项目约定的 conf 生成命令），并已**同步修改**该应用实际使用的配置文件（如 `config/server.yaml`），为新增/变更的 Bootstrap 字段补充对应配置项；认证/登录领域配置使用 `authDomain`（勿用 loginDomain）
- [ ] **跨领域 allowlist**：若本应用在 server 中注册了其他领域的服务（如 goddess、rabbit），已将被引用领域中需「免登录」或「免选命名空间」的 Operation 加入本应用的 `authAllowList` 或 `namespaceAllowList`
- [ ] 新增 HTTP 路径与已有路径无冲突（尤其已有 `/resource/{id}` 时，勿用 `/resource/word`，改用 `/resources/word` 等）
- [ ] 未新建与现有 app 结构不符的目录或包
- [ ] bo/repository/do/convert/impl/biz/service 分层与现有模块一致
- [ ] 分页、错误、上下文、枚举、ID 均使用项目与 magicbox 既有类型
- [ ] **Repository/impl 单一职责**：每个 repo 方法只做单表/单次查询；多表或组合判断在 biz 层通过多次调用 repo 完成，不在一个 repo 方法内混多表
- [ ] **data/impl 使用 gen 模式**：条件与排序使用 query 的 field 表达式（如 `u.UserUID.Eq(...)`、`u.SortOrder.Desc()`），未使用手写字符串；批量插入使用 `CreateInBatches` 或一次 `Create(slice)`，未在循环内逐条 Create
- [ ] **DB 操作严格走 gen**：data/impl 中 Create/Update/Delete/Get/List/Count 必须通过 `query.Xxx.WithContext(ctx)` 完成，未直接使用原生 GORM `Model/Where/First/Find/Create/Save/Delete` 链式调用
- [ ] **无循环校验/循环查详情**：批量校验用 CountXxxByUIDs 等一次查询后比较数量；按 UID 列表拉详情用 GetXxxByUIDs 等一次查询后在 biz 用 map 按顺序组装，未在 for 内逐条 Get
- [ ] **排序语义正确**：若「越大越靠前」，存库时第一项赋最大 SortOrder（可用 slices.Reverse），查询用 Order Desc
- [ ] Import 顺序：标准库 → 空白 → 第三方 → 当前项目
- [ ] 新增的 repository、biz、service、impl 已在对应 ProviderSet 与 Register* 中注册
- [ ] 公共可复用函数或方法已添加测试
- [ ] **README 已同步**：若有模块/API 新增、修改或删除，已同时更新对应应用的 `README.md` 与 `README-zh_CN.md`；且 **功能特性（Features）** 与 **接口概览（API Overview）表** 均已补充/修改/删除对应内容，中英文一一对应，与当前 proto 一致
- [ ] **Comments and logs in English**：All comments, log messages, and error strings are in professional English; no Chinese in source
- [ ] **无重复逻辑**：相同或高度相似的校验/逻辑已抽取为私有方法或共享函数，由多处调用，未复制粘贴

更多分层与文件命名细节见 [reference.md](reference.md)。

## 二次修改高频约束（补充）

以下规则来自实际返工点，默认同样强制执行。

### A. 参数设计：除 `ctx` 外超过 2 个参数必须收敛为结构体

- **适用范围**：biz/repository/impl/convert 的函数与方法。
- **规则**：函数签名中除 `context.Context` 外，其余参数若超过 2 个，必须封装到一个 `*bo.XxxInput` / `*bo.XxxRepoBo` 结构体中，不允许长参数列表。
- **建议命名**：
  - 业务输入：`XxxInput`、`SubmitXxxInput`、`ExecuteXxxBo`
  - 仓储输入：`XxxRepoBo`、`UpdateXxxRepoBo`、`CountXxxRepoBo`
- **收益**：避免签名膨胀，便于扩展字段，减少跨层改动成本。

### B. 状态/类型字段必须使用 `magicbox/enum`

- **规则**：凡是状态、类型、动作类别（如审核状态、审核类型）等有限集合值，必须先定义在 `proto/magicbox/enum/enum.proto`，并在业务 proto 中 `import "magicbox/enum/enum.proto"` 后引用。
- **禁止**：
  - 在业务 proto 内重复定义同义 enum。
  - 在 Go 代码中用 `int32` 常量模拟枚举语义。
- **同步要求**：
  1. 更新 `proto/magicbox/enum/enum.proto`
  2. 生成 `magicbox/enum/*.pb.go`（`cd magicbox && make proto`）
  3. 再生成目标 app 的 API 代码（如 `cd app/<app> && make api`）
  4. 同步 `magicbox/README.md` 与 `magicbox/README-zh_CN.md` 的 enum 说明。

### C. data/impl 更新写法：优先 `UpdateColumnSimple`

- **规则**：涉及多列更新时，优先使用 gorm gen 的 `UpdateColumnSimple(columns...)` + `[]field.AssignExpr`。
- **严格要求（新增）**：data/impl 的所有 DB 读写（Create / Update / Delete / Get / List / Count / Order / Pagination）都必须通过 `internal/data/impl/query` 的 gen 对象完成（如 `query.Xxx.WithContext(ctx)`）。
- **禁止**：
  - 直接使用 `r.DB().WithContext(ctx).Model(...).Where(...).First/Find/Create/Save/Delete/Count` 这类原生 GORM 链式调用。
  - `Updates(map[string]interface{}{...})`（字符串列名）
  - 任何手写列名字符串更新方式。
- **推荐写法**：
  - `columns := []field.AssignExpr{ q.FieldA.Value(v), q.FieldB.Value(v2) }`
  - `..., err := q.WithContext(ctx).Where(...).UpdateColumnSimple(columns...)`

### D. 转换职责归位

- **service 层**：不承载 BO↔API 的细节转换实现。
- **规则**：
  - `BO -> APIV1` 转换函数放在 `internal/biz/bo/`（如 `ToAPIV1XxxItem`）。
  - `DO <-> BO` 转换函数放在 `internal/data/impl/convert/`。
  - impl 内临时出现的字段映射函数（如从审核 DO 提取业务字段）应下沉到 `convert`，避免散落在 impl。
  - 类似 `toXxxItemBo` 这类 DO->BO 映射辅助函数，不允许留在 `internal/data/impl/*.go`，必须统一放到 `internal/data/impl/convert/*.go`。

### E. proto 字段命名：message 字段统一小驼峰

- **规则**：`message` 字段名必须是 `camelCase`。
- **禁止**：`snake_case` 字段名（如 `work_dir`, `page_size`, `status_filter`）。
- **注意**：
  - 若 HTTP path 变量引用 request 字段，路径变量名需与字段一致（如 `{commandUid}` 对应 `commandUid`）。
  - 修改字段命名后必须重新生成代码并编译验证。

### F. 时间字符串格式化统一使用 `timex`

- **规则**：BO 转 API 的时间字符串输出，优先使用 `magicbox/timex`（如 `timex.FormatTime`），不要在业务代码里散落 `time.RFC3339` 手动格式化。
- **示例**：
  - `CreatedAt: timex.FormatTime(&x.CreatedAt)`
  - `ReviewedAt: timex.FormatTime(x.ReviewedAt)`

### G. 减少 convert 冗余包装函数

- **规则**：相同语义的 map/safety-map 转换只保留单一入口，避免 `A -> B -> C` 的无意义包裹。
- **建议**：
  - 明确命名一对函数（如 `ToPlainEnv` / `ToSafetyEnv`）
  - 全局替换调用点，删除重复别名函数，保持 API 面最小化。
