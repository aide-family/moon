---
name: alioth-backend
description: >-
  Implements Alioth Go services with Kratos, ent, and Redis locks. Use when
  editing app/alioth, app/megrez, app/merak, app/alkaid, api protobuf, pkg, or
  when the user mentions ent, PostgreSQL schema, service key, JWT, or
  distributed locks.
---

# Alioth 后端

先读 [alioth-platform](../alioth-platform/SKILL.md) 和 [docs/iterations/conventions.md](../../../docs/iterations/conventions.md)。

## 分层

每个应用使用 `cmd`、`internal/server`、`internal/service`、`internal/biz`、`internal/data`、`internal/conf`。依赖方向是 server → service → biz → data。`biz` 定义仓储接口。用 Wire 组装。应用之间不 import 对方的 `internal`。

业务表和迁移只放在 `app/alioth` 的 ent schema。`megrez`、`merak`、`alkaid` 不连接业务库。

## 接口

对外 HTTP 与 gRPC 只由 `alioth` 提供，前缀 `/v1`。路径在 `api/alioth/v1` 定稿，不新造设计文档没有的资源。

内部调用不进 Web nginx：

- `api/megrez/v1`：`alioth` 实现，`megrez` 调用
- `api/merak/v1`：`alioth` 实现，`merak` 调用
- Agent 双向流在 `api/alioth/v1`，调用者是 `alkaid`

## 凭证

用户 JWT、`Bearer sk-...` service key、Agent 令牌三者不能互相代替。业务请求带 `X-Namespace`，查询必须带 `namespace_uid`。

## Redis 锁

只使用 `pkg` 里的锁客户端。

| 键 | 持有者 | TTL |
| --- | --- | --- |
| `alioth:lock:schedule:{name}` | `alioth` | 15 秒，不续期 |
| `alioth:lock:eval:{rule_uid}` | `megrez` | 30 秒，每 10 秒续期 |
| `alioth:lock:notify:{task_uid}` | `merak` | 60 秒，每 20 秒续期 |

加锁用 `SET key token NX PX`。续期和释放用 Lua，value 必须等于 token。续期失败或 Redis 不可用时，不把本次执行报成成功。任务状态留在 PostgreSQL。
