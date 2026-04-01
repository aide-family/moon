# Jade Tree（玉树）

Jade Tree 是 Moon 的 Agent 运行时服务。

## 职能定位

- 作为 agent 部署在目标服务器中。
- 提供稳定的采集与通信端能力。
- 面向生产采用 RPM + `systemctl` 的运维方式。
- 管理预置 SSH 命令模板：新增与变更需走审核；审核通过后写入正式命令表；可按模板在远端执行。

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

上述 SSH 命令相关接口均需要已登录 JWT 用户（中间件写入的 user UID）。`SSHCommand` 的 OpenAPI 输出在 `internal/server/swagger/openapi.yaml`。

## 部署

Jade Tree 以 RPM + systemd 为推荐部署方式。

- systemd 单元模板：`deploy/systemd/jade-tree.service`
- RPM 说明：`deploy/rpm/README.md`

## 启动

```bash
make all
make dev
```

修改 `internal/data/impl/do/` 下模型后，重新生成 gorm gen 查询代码：

```bash
make gen
```
