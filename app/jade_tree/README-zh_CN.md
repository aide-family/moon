# Jade Tree（玉树）

Jade Tree 是 Moon 的 Agent 运行时服务。

## 职能定位

- 作为 agent 部署在目标服务器中。
- 提供稳定的采集与通信端能力。
- 面向生产采用 RPM + `systemctl` 的运维方式。
- 管理预置 SSH 命令模板：新增与变更需走审核；审核通过后写入正式命令表；可按模板在远端执行。
- 支持一键命令下发，支持下发前 count 预估，以及按机器/系统类型/采集器版本做正向与排除筛选。
- 提供部署机器基础信息采集能力（CPU、内存、磁盘与挂载点、网络、主机名、系统基础信息）。
- 支持按配置定时向指定 HTTP 端点主动上报本机机器基础信息（包含采集器 `agent` 连接与版本信息）。
- 提供机器信息 CLI 命令：`machine get`、`machine list`、`machine push`、`machine pull`，支持 `table/json/yaml` 输出。
- 以 Prometheus metrics 方式暴露探测能力（`probe_tcp_*`、`probe_http_*`、`probe_port_*`、`probe_tls_cert_*`）。

## 架构

Jade Tree 保持与现有 Moon 服务一致的分层：

- `cmd` -> 命令入口与运行模式
- `internal/server` -> HTTP/gRPC 服务启动与中间件
- `internal/service` -> 服务处理层
- `internal/biz` -> 业务层
- `internal/data` -> 数据与仓储层

API 定义位于 `proto/jade_tree/api/v1/`；生成代码在 `pkg/api/v1/`。修改 proto 后执行 `make api`（已包含在 `make all` 中）。审核单的 **kind** / **status** 使用 `proto/magicbox/enum/enum.proto` 中的 `SSHCommandAuditKind`、`SSHCommandAuditStatus`（变更后需在 `magicbox/` 下执行 `make proto`）。

## 配置说明

- **数据库**（`config/server.yaml` 中的 `bootstrap.database`）为**必填**。默认示例为 SQLite，本地开发可直接使用；生产可改为 MySQL / PostgreSQL（`dialector` 与 `options` 结构与其他 Moon 应用一致）。
- 表 `ssh_commands`（已审核通过的命令）与 `ssh_command_audits`（审核单）在进程启动时通过 GORM `AutoMigrate` 创建/迁移。
- 机器信息主动上报由 `bootstrap.machineInfoReport` 控制（`enabled`、`collectSelfData`、`interval`、`timeout`、`endpoints` 以及可选 `headers`）。

## 接口概览

| 服务 | 方法 / HTTP | 说明 |
|------|-------------|------|
| `magicbox.api.v1.Health` | `GET /health` | 健康检查（存活/就绪） |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/submissions` | 提交新命令审核（写入审核表，待审批） |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/{command_uid}/submissions` | 提交对已有命令的变更审核 |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-commands` | 分页列出已生效命令 |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-commands/{uid}` | 获取单条已生效命令 |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-command-audits` | 分页列出审核记录（可选 query：`statusFilter`） |
| `jade_tree.api.v1.SSHCommand` | `GET /v1/ssh-command-audits/{uid}` | 获取单条审核记录 |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-audits/{uid}/approve` | 审批通过：更新正式命令表并更新审核状态 |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-audits/{uid}/reject` | 驳回审核并填写原因 |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-commands/{command_uid}/execute` | 选择已生效命令，携带主机与凭证在远端执行 |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/batch-execute` | 批量接收并执行多条命令请求 |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/dispatch/count` | 一键下发前执行目标机器数量预估（默认排除自身） |
| `jade_tree.api.v1.SSHCommand` | `POST /v1/ssh-command-actions/dispatch` | 根据正向/排除筛选条件将命令一键下发到匹配的上报 agent |
| `jade_tree.api.v1.MachineInfo` | `GET /v1/machine-info` | 获取部署机器详情（CPU、内存、磁盘+挂载点容量、网络、主机名、机器唯一标识、架构/系统/版本/内核，以及 `agent` 连接/版本） |
| `jade_tree.api.v1.MachineInfo` | `POST /v1/machine-info/report` | 上报机器信息（按机器 UUID + 主机名 + 本机 IP 组合去重合并并落库）；返回空响应 |
| `jade_tree.api.v1.MachineInfo` | `GET /v1/machine-infos` | 获取接收方/主节点已知的机器信息（同一组合自然键，按 `page`/`pageSize` 分页） |
| `jade_tree.api.v1.ProbeTask` | `POST /v1/probe-tasks` | 新增数据库探测任务 |
| `jade_tree.api.v1.ProbeTask` | `PUT /v1/probe-tasks/{uid}` | 更新探测任务字段（`type/host/port/url/name/timeoutSeconds`） |
| `jade_tree.api.v1.ProbeTask` | `PATCH /v1/probe-tasks/{uid}/status` | 状态管理（`ENABLED` / `DISABLED`） |
| `jade_tree.api.v1.ProbeTask` | `DELETE /v1/probe-tasks/{uid}` | 删除探测任务并动态移除 |
| `jade_tree.api.v1.ProbeTask` | `GET /v1/probe-tasks/{uid}` | 获取单个探测任务 |
| `jade_tree.api.v1.ProbeTask` | `GET /v1/probe-tasks` | 分页查询探测任务 |

探测指标通过 `config/server.yaml` 中 `bootstrap.probe` 配置，复用现有 `GET /metrics` 暴露。

上述 SSH 命令相关接口均需要已登录 JWT 用户（中间件写入的 user UID）。`SSHCommand` 的 OpenAPI 输出在 `internal/server/swagger/openapi.yaml`。

## 部署

Jade Tree 以 RPM + systemd 为推荐部署方式。

- systemd 单元模板：`deploy/systemd/jade-tree.service`
- RPM 说明：`deploy/rpm/README.md`
- RPM 打包文件：`packaging/rpm/`

## RPM 打包

Linux 本机构建（需要 `rpmbuild` 与 `rsync`）：

```bash
make rpm
```

本机构建源码 RPM：

```bash
make srpm
```

macOS / 无本地 rpmbuild（推荐）：

```bash
make rpm-docker
```

产物位置：

- 二进制 RPM：`rpmbuild/RPMS/<arch>/jade-tree-<version>-<release>.<arch>.rpm`
- 源码 RPM：`rpmbuild/SRPMS/jade-tree-<version>-<release>.src.rpm`

## 启动

```bash
make all
make dev
```

## 机器信息命令行

- `jade_tree machine get [endpoint...]`：对每个目标调用 `GET /v1/machine-info`，每个端点本机一行摘要（位置参数、配置 `endpoints`，否则 `--endpoint` 或配置 `endpoint`）。子命令 `get cpu|memory|network|disk|sys [endpoint...]` 对每个端点输出同类详情（`sys`：架构、操作系统、版本、内核）。
- `jade_tree machine list [endpoint...]`：对每个目标分页调用 `GET /v1/machine-infos`（`--page-size`），列出该节点存储中已知的全部机器。子命令 `list cpu|memory|network|disk|sys [endpoint...]` 对上述每台机器输出同类详情。
- `jade_tree machine push [endpoint...]`：从 `--from`（或配置 `endpoint`）将本机及已存储机器信息推送到目标端点。
- `jade_tree machine pull [endpoint...]`：从 `--from`（源端点）拉取机器信息并存储到给定目标端点。
- 所有命令都支持 `--output table|json|yaml`（默认 `table`）。
- 所有机器命令都支持通过 `--jwt` 或配置文件进行 JWT 鉴权。

默认客户端配置文件：`~/.jade_tree/client.yaml`

```yaml
endpoint: http://127.0.0.1:8000
endpoints:
  - http://10.0.0.11:8000
  - http://10.0.0.12:8000
jwt: eyJhbGciOi...
```

命令行参数优先级高于配置文件。

修改 `internal/data/impl/do/` 下模型后，重新生成 gorm gen 查询代码：

```bash
make gen
```
