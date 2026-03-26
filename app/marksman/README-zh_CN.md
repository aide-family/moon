# Marksman（后羿）

<p align="center">
  <strong>Moon 平台事件与告警服务</strong>
</p>

<p align="center">
  <a href="README-zh_CN.md">中文</a> · <a href="README.md">English</a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="https://github.com/go-kratos/kratos"><img src="https://img.shields.io/badge/Kratos-v2.9.2-00ADD8?style=flat&logo=go" alt="Kratos"></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/Cobra-v1.10-00ADD8?style=flat&logo=go" alt="Cobra"></a>
</p>

---

## 目录

- [项目简介](#项目简介)
- [功能特性](#功能特性)
- [接口概览](#接口概览)
- [环境要求](#环境要求)
- [安装](#安装)
- [快速开始](#快速开始)
- [常用用法](#常用用法)
- [开发说明](#开发说明)
- [许可证](#许可证)
- [致谢](#致谢)

---

## 项目简介

**Marksman**（后羿）是 Moon 平台的事件与告警服务，负责数据源（Prometheus、VictoriaMetrics、Elasticsearch、Jaeger）、策略组与策略、告警级别、以及策略指标（表达式、级别、接收人绑定）的管理。

- **接口定义**：`proto/marksman/api/v1/`
- 基于 Kratos 提供 **HTTP + gRPC**，CLI 基于 Cobra。

---

## 功能特性

- **当前用户（Self，goddess）**：当前用户信息、命名空间列表、修改邮箱/头像/备注、刷新 Token
- **用户（User，goddess）**：用户增删改查、列表、下拉选择、封禁/解封
- **成员（Member，goddess）**：命名空间内成员列表/获取/选择、邀请成员、移除成员、更新状态
- **命名空间（Namespace，goddess）**：命名空间管理（与 goddess 共用能力，需配置 namespaceDomain）
- **验证码（Captcha，goddess）**：图形验证码获取（id、base64 图片），用于登录等无需鉴权场景
- **数据源（Datasource）**：指标/日志/链路数据源增删改查、列表、下拉选择、更新状态（ENABLED/DISABLED）、按数据源的状态时序（主时序库）、指标元数据（label 名与各 label 取值）查询（Prometheus、VictoriaMetrics、Elasticsearch、Jaeger；包含 `levelUid` 绑定到 DATASOURCE 级别）
- **指标查询（MetricQuery）**：指标类型数据源即时查询（Prometheus /api/v1/query）、区间查询（/api/v1/query_range）、以及直接 HTTP 代理
- **策略组（Strategy group）**：增删改查、列表、选择、状态；绑定接收人（收件人组）
- **策略（Strategy）**：增删改查、列表、状态；归属策略组；类型（METRICS/LOGS/TRACE）与驱动
- **级别（Level）**：级别增删改查、列表、选择、状态（区分 `type`：ALERT/DATASOURCE；用于告警严重程度分组，包含用于展示的 `bgColor`）
- **策略指标（Strategy metric）**：保存/查询指标配置（expr、labels、datasourceUIDs、levels）；指标级别的增删改查（mode、condition、values、duration）；按策略绑定接收人（可选 levelUID）
- **告警（实时）**：告警页面增删改查（名称、颜色、排序、按策略组/级别/策略/数据源/数据源级别筛选）；按告警页列表展示实时告警事件；事件操作：介入、抑制（至指定时间）、恢复（手动）；告警统计（当前总告警、按等级、今日已恢复、按告警页）；用户关注的告警页（按用户列出/保存）
- **通知组（Notification group）**：通知组增删改查（名称、备注、元数据、成员、webhook、模板通过创建/更新维护）
- **通知组订阅（Notification group subscription）**：按通知组获取/更新订阅筛选（策略组、策略、策略-级别对、labels、excludeLabels、datasourceUids）；多维度同时设置时告警匹配任意一项即可（OR）

---

## 接口概览

| 服务 | 方法 / HTTP | 说明 |
|------|-------------|------|
| **Self**（goddess） | `GET /v1/self/info` | 当前用户信息 |
| | `GET /v1/self/namespaces` | 当前用户命名空间列表 |
| | `PUT /v1/self/change-email`、`change-avatar`、`change-remark` | 修改邮箱/头像/备注 |
| | `GET /v1/self/refresh-token` | 刷新 Token |
| **User**（goddess） | `GET /v1/user/{uid}` | 获取用户 |
| | `GET /v1/users`、`GET /v1/users/select` | 用户列表、下拉选择 |
| | `PUT /v1/user/ban/{uid}`、`PUT /v1/user/permit/{uid}` | 封禁/解封用户 |
| **Member**（goddess） | `GET /v1/members`、`GET /v1/member/{uid}`、`GET /v1/members/select` | 成员列表、获取、下拉选择 |
| | `POST /v1/member/invite` | 邀请成员（email、role） |
| | `DELETE /v1/member/{uid}`、`PUT /v1/member/{uid}/status` | 移除成员、更新状态 |
| **Captcha**（goddess） | `GET /v1/captcha` | 获取图形验证码（返回 captchaId、captchaB64s） |
| **Datasource** | `POST /v1/datasource` | 创建数据源（name、type、driver、url、metadata、remark、`levelUid` 绑定到 DATASOURCE 级别） |
| | `PUT /v1/datasource/{uid}` | 更新数据源（包含 `levelUid`） |
| | `PUT /v1/datasource/{uid}/status` | 更新状态（ENABLED/DISABLED） |
| | `DELETE /v1/datasource/{uid}` | 删除数据源 |
| | `GET /v1/datasource/{uid}` | 获取数据源 |
| | `GET /v1/datasources` | 列表（keyword、page、pageSize、type、driver、status） |
| | `GET /v1/datasources/select` | 下拉选择 |
| | `GET /v1/datasource/{uid}/status` | 单个数据源状态时序（从主时序库查询；参数 startTime、endTime、stepSeconds；默认最近 1 小时，步长 60s） |
| | `GET /v1/datasource/{uid}/metrics` | 指标列表（仅名称、说明、unit、type；可选 match[]、limit；默认 limit 100） |
| | `GET /v1/datasource/{uid}/metric-label-detail` | 单个 metric 的 label 明细：labels 及各 label 的取值（查询参数 metric=名称） |
| **MetricQuery**（指标查询） | `POST /v1/metric-query/query` | 即时查询（body: uid、query、可选 time）；返回 Prometheus 风格 JSON |
| | `POST /v1/metric-query/query-range` | 区间查询（body: uid、query、start、end、step）；返回 Prometheus 风格 JSON |
| | `POST /v1/metric-query/proxy` | 直接代理到数据源（body: uid、path、method、可选 body）；返回 status_code 与 body |
| **Strategy**（策略组） | `POST /v1/strategy-group` | 创建策略组 |
| | `PUT /v1/strategy-group/{uid}` | 更新策略组 |
| | `PUT /v1/strategy-group/{uid}/status` | 更新状态（ENABLED/DISABLED） |
| | `DELETE /v1/strategy-group/{uid}` | 删除策略组 |
| | `GET /v1/strategy-group/{uid}` | 获取策略组 |
| | `GET /v1/strategy-groups` | 策略组列表 |
| | `GET /v1/strategy-groups/select` | 下拉选择 |
| | `POST /v1/strategy-group/{uid}/receivers` | 绑定接收人（receiverUIDs） |
| **Strategy**（策略） | `POST /v1/strategy` | 创建策略（name、type、driver、strategyGroupUID、status 等） |
| | `PUT /v1/strategy/{uid}` | 更新策略 |
| | `PUT /v1/strategy/{uid}/status` | 更新状态 |
| | `DELETE /v1/strategy/{uid}` | 删除策略 |
| | `GET /v1/strategy/{uid}` | 获取策略 |
| | `GET /v1/strategies` | 列表（keyword、page、pageSize、status、strategyGroupUID、type、driver） |
| | `GET /v1/strategies/select` | 策略下拉选择（keyword、limit、lastUid、status、strategyGroupUids 按策略组筛选，策略组 ID 为列表） |
| **Level** | `POST /v1/level` | 创建级别（name、remark、metadata、bgColor、`type` = ALERT/DATASOURCE） |
| | `PUT /v1/level/{uid}` | 更新级别（bgColor、`type`） |
| | `PUT /v1/level/{uid}/status` | 更新状态 |
| | `DELETE /v1/level/{uid}` | 删除级别 |
| | `GET /v1/level/{uid}` | 获取级别 |
| | `GET /v1/levels` | 列表（page、pageSize、keyword、status） |
| | `GET /v1/levels/select` | 下拉选择 |
| **StrategyMetric** | `POST /v1/metric/strategy/{strategyUID}` | 保存策略指标（expr、labels、datasourceUIDs、summary、description、status） |
| | `GET /v1/metric/strategy/{strategyUID}` | 获取策略指标（含 levels） |
| | `POST /v1/metric/strategy/{strategyUID}/level` | 保存指标级别（levelUID、mode、condition、values、duration、status） |
| | `PUT /v1/metric/strategy/{strategyUID}/level/{uid}/status` | 更新指标级别状态 |
| | `DELETE /v1/metric/strategy/{strategyUID}/level/{uid}` | 删除指标级别 |
| | `GET /v1/metric/strategy/{strategyUID}/level/{uid}` | 获取指标级别 |
| | `POST /v1/metric/strategy/{strategyUID}/receivers` | 绑定接收人（receiverUIDs；可选 levelUID） |
| **NotificationGroup**（通知组） | `POST /v1/notification-groups` | 创建通知组（name、remark、metadata、members、webhooks、templates） |
| | `PUT /v1/notification-groups/{uid}` | 更新通知组（含 members、webhooks、templates） |
| | `PUT /v1/notification-groups/{uid}/status` | 更新状态（ENABLED/DISABLED） |
| | `DELETE /v1/notification-groups/{uid}` | 删除通知组 |
| | `GET /v1/notification-groups/{uid}` | 获取通知组 |
| | `GET /v1/notification-groups` | 通知组列表（page、pageSize、keyword、status） |
| **NotificationGroupSubscription**（通知组订阅） | `GET /v1/notification-groups/{notification_group_uid}/subscription` | 获取某通知组的订阅筛选 |
| | `PUT /v1/notification-groups/{notification_group_uid}/subscription` | 保存订阅筛选（无则创建、有则覆盖；strategy_group_uids、strategy_uids、strategy_levels、labels、exclude_labels、datasource_uids；多维度同时设置时匹配任意一项即可（OR） |
| **Alert**（告警页） | `POST /v1/alert-pages` | 创建告警页（name、color、sortOrder、filter 支持按策略组/级别/策略/数据源/数据源级别筛选） |
| | `PUT /v1/alert-pages/{uid}` | 更新告警页（filter 支持按策略组/级别/策略/数据源/数据源级别筛选） |
| | `DELETE /v1/alert-pages/{uid}` | 删除告警页 |
| | `GET /v1/alert-pages/{uid}` | 获取告警页 |
| | `GET /v1/alert-pages` | 告警页列表（page、pageSize、keyword） |
| **Alert**（实时） | `GET /v1/alert-pages/{alertPageUid}/realtime-alerts` | 按页查询实时告警事件（page、pageSize、status；包含 level 的 `bgColor` 和 datasource levelName） |
| | `GET /v1/alert-statistics` | 告警统计（当前总告警、按等级、今日已恢复、按告警页） |
| | `POST /v1/realtime-alerts/{uid}/intervene` | 介入（值班接管） |
| | `POST /v1/realtime-alerts/{uid}/suppress` | 抑制至指定时间（body: suppressUntil RFC3339） |
| | `POST /v1/realtime-alerts/{uid}/recover` | 手动恢复 |
| **Alert**（用户） | `GET /v1/user/alert-pages` | 当前用户关注的告警页列表（个人配置） |
| | `PUT /v1/user/alert-pages` | 保存当前用户关注的告警页（body: alertPageUids，最多 10 个；覆盖原列表） |

**类型**：`DatasourceType`: METRICS, LOGS, TRACE。`LevelType`: ALERT, DATASOURCE。**驱动**：METRICS_PROMETHEUS, METRICS_VICTORIA_METRICS, LOGS_ELASTICSEARCH, TRACE_JAEGER。

接口定义：Marksman 自有 API 位于 `proto/marksman/api/v1/`（如 `datasource.proto`、`strategy.proto`、`level.proto`、`strategy_metric.proto`、`alert.proto`）；Self、User、Member、Namespace、Captcha 等来自 `goddess` 的 `proto/goddess/api/v1/`。可通过 `make api` 生成 OpenAPI。

---

## 环境要求

- [Go](https://go.dev/) 1.25+
- [Make](https://www.gnu.org/software/make/)

---

## 安装

在 Moon 仓库根目录或本应用目录下执行：

```bash
cd app/marksman   # 若在仓库根目录
make init         # 安装 protoc 插件、wire 等
make build        # 生成 API/conf/wire 并编译 → bin/marksman
```

---

## 快速开始

```bash
# 编译
make init && make build

# 开发模式运行（HTTP + gRPC）
./bin/marksman run all --log-level=DEBUG
# 或
make dev
```

---

## 常用用法

### 命令行

```bash
./bin/marksman -h
./bin/marksman version
./bin/marksman run all -h
./bin/marksman run grpc -h
./bin/marksman run http -h
```

### 运行模式

| 命令 | 说明 |
|------|------|
| `./bin/marksman run all` | 同时启动 HTTP 与 gRPC |
| `./bin/marksman run http` | 仅 HTTP |
| `./bin/marksman run grpc` | 仅 gRPC |

### Make 目标

| 目标 | 说明 |
|------|------|
| `make init` | 安装 protoc 插件、wire、kratos 等 |
| `make conf` | 从 proto 生成配置 |
| `make api` | 从 `proto/marksman` 生成 Go/HTTP/gRPC/OpenAPI |
| `make wire` | 生成 Wire 依赖注入 |
| `make all` | api + conf + wire |
| `make build` | all + 编译到 `bin/marksman` |
| `make dev` | `go run . run all --log-level=DEBUG` |
| `make gen` | 生成 DO/数据层（如带 generate tag 的测试） |
| `make clean` | 删除 `bin/` |
| `make help` | 列出所有目标 |

---

## 开发说明

1. **修改 proto 后重新生成**

   ```bash
   make api
   make wire   # 若依赖有变更
   ```

2. **不编译直接运行**

   ```bash
   go run . run all
   go run . run http
   go run . run grpc
   ```

3. **配置**：见 `internal/conf/`；配置文件路径通过启动参数或环境变量指定。数据源状态时序接口（`GET /v1/datasource/{uid}/status`）需在 `config/server.yaml` 中配置 **mainTsdb**：`driver`（`prometheus` 或 `victoria_metrics`）、`url`；环境变量覆盖：`MARKSMAN_MAIN_TSDB_DRIVER`、`MARKSMAN_MAIN_TSDB_URL`。

---

## 许可证

[MIT](LICENSE)

---

## 致谢

- [Kratos](https://github.com/go-kratos/kratos)
- [Cobra](https://github.com/spf13/cobra)
