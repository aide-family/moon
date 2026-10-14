---
name: alioth-milestone
description: >-
  Implements or accepts one Alioth milestone from M0 through M7. Use when the
  user asks to build, continue, or verify a milestone, or mentions M0, M1, M2,
  M3, M4, M5, M6, M7, SSH, SFTP, or Kubernetes API proxy.
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
| M6 SFTP 与 Kubernetes 代理 | [m6.md](../../../docs/iterations/m6.md) |
| M7 多副本部署 | [m7.md](../../../docs/iterations/m7.md) |

## 顺序

M0 → M1。M1 之后，告警线是 M2 → M3 → M4，边缘线是 M5 → M6。M7 等 M3、M4、M6 都通过后再做。

不把「本期不做」的能力写进当前步骤：探测任务、微前端、分库、Redis 队列、短信发送、日志与链路评测。

## 摇光

M5 只做令牌、出站流、心跳和审核后的 SSH。`alkaid` 连接 `alioth:9000`，并监听 8022 与 9022。控制面不回拨 Agent。

M6 才增加 SFTP 和 Kubernetes API 代理。文件字节不落在玉衡磁盘。`sftp:read` 与 `sftp:write`、`k8s:proxy` 分开鉴权。

## 验收

勾选该里程碑文档里的验收项。后端核对端口、库表和 Redis 锁。页面在 `http://<web>:8080` 走完主路径。未通过的项先修再继续下一里程碑。
