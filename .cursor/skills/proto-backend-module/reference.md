# Proto 后端模块 — 参考与示例

本文档补充 SKILL.md 中的分层与命名，给出各层文件位置和典型写法，便于与 goddess、marksman、rabbit 保持一致。

## 1. 应用与 Proto 对应关系

| App      | Proto 根                 | 生成 Go 包（pkg）           |
|----------|--------------------------|-----------------------------|
| goddess  | proto/goddess/api/v1/    | goddess/pkg/api/v1          |
| marksman | proto/marksman/api/v1/   | marksman/pkg/api/v1         |
| rabbit   | proto/rabbit/api/v1/     | rabbit/pkg/api/v1           |

公共枚举与错误：proto 中常用 `magicbox/enum`、`magicbox/merr`，生成代码在 magicbox 仓库或本仓库 magicbox 目录。

## 2. 各层文件位置（以 rabbit 为例）

- **Proto 与生成**
  - 定义：`proto/rabbit/api/v1/<service>.proto`
  - 生成：`app/rabbit/pkg/api/v1/<service>.pb.go`、`*_grpc.pb.go`、`*_http.pb.go`（由 buf/kratos 生成）
  - **HTTP 路由冲突**：同一 service 内若存在 `get: "/v1/foo/{id}"`，则不要再设 `get: "/v1/foo/simple"`（请求 `/v1/foo/simple` 会被前者匹配为 `id=simple` 并导致解析失败）。应使用不同 path 前缀，如 `get: "/v1/foos/simple"` 或 `get: "/v1/foo/by-secret"`（多段）等，确保与所有带 `{var}` 的路径不重叠。

- **BO**
  - 文件：`app/rabbit/internal/biz/bo/<module>.go`（如 `webhook.go`）
  - 内容：Create/Update/List/Select 等 BO 结构体，`NewCreateXxxBo(req)`，`ToAPIV1XxxItem()`、`ToAPIV1ListXxxReply()` 等

- **Repository 接口**
  - 文件：`app/rabbit/internal/biz/repository/<module>.go`（如 `webhook.go`）
  - 内容：`type XxxConfig interface { CreateXxxConfig(...) (snowflake.ID, error); ... }`

- **DO（GORM 模型）**
  - 文件：`app/rabbit/internal/data/impl/do/<table>.go`（如 `webhook_config.go`）
  - 内容：嵌入 `do.BaseModel`，字段用 gorm tag，敏感用 `strutil.EncryptString`、`safety.Map`；在 `do/base.go` 的 `Models()` 中注册

- **Convert**
  - 文件：`app/rabbit/internal/data/impl/convert/<module>.go`（如 `webhook_config.go`）
  - 内容：`ToXxxConfigDO(ctx, req *bo.CreateXxxBo)`、`ToXxxConfigItemBo(do *do.XxxConfig)` 等

- **Query**
  - 由 gorm gen 生成：`app/rabbit/internal/data/impl/query/gen.go`、`query/<table>.gen.go`
  - 新增 do 后需执行项目约定的 gen 命令并重新生成

- **Impl（Repository 实现）**
  - 文件：`app/rabbit/internal/data/impl/<module>.go`（如 `webhook_config.go`）
  - 内容：`NewXxxConfigRepository(d *data.Data) repository.XxxConfig`，实现接口各方法；在 `impl/provider_set.go` 的 `ProviderSetImpl` 中注册

- **Biz**
  - 文件：`app/rabbit/internal/biz/<module>.go` 或 `<module>_config.go`
  - 内容：`NewXxxConfig(repo repository.XxxConfig, helper *klog.Helper) *XxxConfig`，方法内调 repo、用 merr 与 helper.Errorw；在 `biz/provider_set.go` 的 `ProviderSetBiz` 中注册

- **Service**
  - 文件：`app/rabbit/internal/service/<module>.go`（如 `webhook.go`）
  - 内容：实现 `apiv1.XxxServer`，嵌入 `apiv1.UnimplementedXxxServer`，各 RPC 转 bo → 调 biz → 转 reply；在 `service/provider_set.go` 的 `ProviderSetService` 中注册

- **Server 注册**
  - 文件：`app/rabbit/internal/server/provider_set.go`
  - 在 `RegisterHTTPService` / `RegisterGRPCService` 的形参中增加 `xxxService *service.XxxService`，并在函数体内调用 `apiv1.RegisterXxxHTTPServer` / `apiv1.RegisterXxxServer`；若使用 `RegisterService`，同样增加参数并传入

## 3. 命名与类型约定

- **表名**：DO 的 `TableName()` 返回复数或项目约定表名（如 `webhooks`），与数据库一致。
- **主键**：统一使用 `snowflake.ID`，DO 嵌入 `BaseModel` 即包含 `ID`。
- **命名空间/多租户**：DO 中 `NamespaceUID`，在 impl 层用 `contextx.GetNamespace(ctx)` 过滤；创建时用 `contextx.GetUserUID(ctx)` 设 Creator。
- **软删**：使用 `gorm.DeletedAt`，查询时由 GORM 自动过滤。
- **BO 与 Proto 字段**：Proto 用 `snake_case`（如 `created_at`），Go 用 `CreatedAt`；BO 与 do 的字段名与 proto 语义一致，转换函数中显式映射。

## 4. Import 示例（手写 Go，非生成）

```go
import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)
```

顺序：标准库 → 第三方（magicbox 优先）→ 当前项目。有空白导入时，空白导入紧跟标准库后单独一组。

## 5. 测试

- 公共包（magicbox）和 app 内对外复用的函数/方法应有 `*_test.go`。
- 测试风格与现有项目一致（表驱动、mock repository 等），不新建与项目风格不符的测试框架或目录。

## 6. Goddess 的 domain 层

- 领域实现：`app/goddess/domain/<domain>/v1/`（如 `namespace`、`user`）下有该领域的实现与注册（如 `register.go`、`goddess.go`）。
- 新增「领域」时在此目录下新增包并注册；仅扩展现有 API 时在 `internal/service` 与 `internal/biz` 按上述流程即可，无需动 domain。

完成模块后，用 SKILL.md 末尾的检查清单自检，并确保未引入新目录或与现有 app 不一致的包结构。

## 7. 配置与 conf.proto 同步

- **conf.proto**：定义在 `app/<app>/internal/conf/conf.proto`，描述 Bootstrap 等配置结构；生成 `conf.pb.go` 需执行该应用的 `make conf`（或项目约定的命令）。
- **运行时配置**：应用启动时加载的是 YAML 等文件（如 `app/<app>/config/server.yaml`），**不是** conf.proto。因此：
  - **新增** Bootstrap 字段时：除执行 `make conf` 外，必须在实际使用的配置文件中**新增对应配置项**（键名与 proto 字段 camelCase 一致，如 `selfConfig`、`userConfig`），结构参考同类型已有项（如 `namespaceConfig` 的 driver、version、options）。
  - **删除或重命名**字段时：同步从配置文件中删除或重命名对应项。
- 未同步时，新字段在运行时为 nil/零值，依赖该配置的代码可能报错（如 "xxxConfig is required"）。详见 SKILL.md「配置与 conf.proto 同步」与「何时需要更新运行时配置」。

## 8. README 维护（与 SKILL 同步）

- **位置**：每个应用有 `README.md`、`README-zh_CN.md`；根目录与 magicbox 同理。见 SKILL.md「README 与文档同步」。
- **API 来源**：README 中的「API Overview / 接口概览」必须与 `proto/<app>/api/v1/*.proto` 中的 `service`、`rpc` 及 `option (google.api.http)` 一致；修改/增删 RPC 或 HTTP 路径后，必须同步改两张表（中英文）。
- **表格格式**：列为「Service / 服务」「Method / HTTP」「Description / 说明」；新增一行即按现有格式补全三列，删除则整行去掉。
- **禁止臆造**：路径、方法名、服务名以 proto 与生成代码为准，不确定时直接查 proto 文件。
