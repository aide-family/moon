---
name: alioth-platform
description: >-
  Applies Alioth service boundaries, ports, and out-of-scope limits before
  changing alioth, megrez, merak, alkaid, or web. Use when working in the
  Alioth repo, choosing where code belongs, or when the user mentions 玉衡,
  天权, 天璇, 摇光, PostgreSQL, Redis locks, or multi-replica deployment.
---

# Alioth 平台边界

改任何服务前先读 [docs/design.md](../../../docs/design.md) 和 [docs/roadmap.md](../../../docs/roadmap.md)。实现细节以当前里程碑的 [docs/iterations/](../../../docs/iterations/) 为准，不要把后续里程碑提前做进来。后端框架为 Kratos v3（详见 [alioth-backend](../alioth-backend/SKILL.md)）。

## 进程

| 编码名 | 职责 |
| --- | --- |
| `alioth` | 玉衡。唯一写入 PostgreSQL 的控制面 |
| `megrez` | 天权。只评测告警规则 |
| `merak` | 天璇。只投递通知 |
| `alkaid` | 摇光。出站连接 `alioth:9000`，执行 SSH 与 Web Shell |
| `web` | 集群外唯一入口 |

## 端口

- `web` HTTP 8080，`/v1` 只反代到 `alioth:8000`
- `alioth` HTTP 8000，gRPC 9000
- `megrez` HTTP 8001，仅 `/health`
- `merak` HTTP 8002，仅 `/health`
- `alkaid` HTTP 8022，gRPC 9022。这两个端口不对集群外发布

浏览器只访问 8080。

## 数据

业务数据在玉衡的 ent schema。Redis 只保存分布式锁，不当缓存、会话或消息队列。`megrez`、`merak`、`alkaid` 不建业务表。

## 本期不做

探测任务、微前端、每个服务一块业务库、短信发送实现、日志与链路的规则评测。
