# Jade Tree（玉树）

Jade Tree 是 Moon 的 Agent 运行时服务。

## 职能定位

- 作为 agent 部署在目标服务器中。
- 提供稳定的采集与通信端能力。
- 面向生产采用 RPM + `systemctl` 的运维方式。

## 架构

Jade Tree 保持与现有 Moon 服务一致的分层：

- `cmd` -> 命令入口与运行模式
- `internal/server` -> HTTP/gRPC 服务启动与中间件
- `internal/service` -> 服务处理层
- `internal/biz` -> 业务层
- `internal/data` -> 数据与仓储层

## 接口概览

| 服务 | 方法 / HTTP | 说明 |
|------|-------------|------|
| `magicbox.api.v1.Health` | `GET /health` | 健康检查（存活/就绪） |

## 部署

Jade Tree 以 RPM + systemd 为推荐部署方式。

- systemd 单元模板：`deploy/systemd/jade-tree.service`
- RPM 说明：`deploy/rpm/README.md`

## 启动

```bash
make all
make dev
```
