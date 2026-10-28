---
name: alioth-milestone
description: >-
  Implements or accepts one Alioth milestone from M0 through M15. Use when the
  user asks to build, continue, or verify a milestone, or mentions any M0-M15
  milestone, SSH, Web Shell, Loki, VictoriaLogs, Elasticsearch, OnCall, or
  Kubernetes.
---

# 按里程碑实现

先读 [docs/roadmap.md](../../../docs/roadmap.md) 和 [docs/iterations/conventions.md](../../../docs/iterations/conventions.md)。然后只打开当前这一份细则，并按它的范围、开发细则和勾选验收执行。

| 里程碑 | 细则 |
| --- | --- |
| M0 骨架 | [m0.md](../../../docs/iterations/m0.md) |
| M1 身份与租户 | [m1.md](../../../docs/iterations/m1.md) |
| M2 告警配置 | [m2.md](../../../docs/iterations/m2.md) |
| M3 天权评测 | [m3.md](../../../docs/iterations/m3.md) |
| M4 天璇通知 | [m4.md](../../../docs/iterations/m4.md) |
| M5 摇光 SSH | [m5.md](../../../docs/iterations/m5.md) |
| M7 多副本部署 | [m7.md](../../../docs/iterations/m7.md) |
| M8 通知增强 | [m8.md](../../../docs/iterations/m8.md) |
| M9 平台用户与邀请确认 | [m9.md](../../../docs/iterations/m9.md) |
| M10 Loki LogQL 告警闭环 | [m10-loki-alerts.md](../../../docs/iterations/m10-loki-alerts.md) |
| M11 VictoriaLogs LogsQL 告警闭环 | [m11-vlogs-alerts.md](../../../docs/iterations/m11-vlogs-alerts.md) |
| M12 Elasticsearch 时间桶告警闭环 | [m12-es-alerts.md](../../../docs/iterations/m12-es-alerts.md) |
| M13 OnCall 值班路由 | [m13-oncall.md](../../../docs/iterations/m13-oncall.md) |
| M14 Kubernetes 多集群管理 | [m14-k8s.md](../../../docs/iterations/m14-k8s.md) |
| M15 告警规则自定义变量 | [m15-alert-rule-variables.md](../../../docs/iterations/m15-alert-rule-variables.md) |

## 顺序

M0 → M1。M1 之后，告警线是 M2 → M3 → M4，边缘线是 M5，平台线是 M9。日志告警在 M10（Loki）之后由 M11（VictoriaLogs）、M12（Elasticsearch）扩展，通知路由在 M8 之后由 M13（OnCall）扩展。M14（Kubernetes）只依赖 M1，M15（规则变量）依赖 M3、M4。M7 等 M3、M4、M5 都通过后再做整平台多副本验收。

不把「本期不做」的能力写进当前步骤：探测任务、微前端、分库、Redis 队列、短信发送、Loki / VictoriaLogs / Elasticsearch 数值型以外的日志与链路评测。

## 摇光

M5 只做令牌、出站流、心跳和审核后的 SSH。`alkaid` 连接 `alioth:9000`，并监听 8022 与 9022。控制面不回拨 Agent。

## 验收

勾选该里程碑文档里的验收项。后端核对端口、库表和 Redis 锁。页面在 `http://<web>:8080` 走完主路径。未通过的项先修再继续下一里程碑。
