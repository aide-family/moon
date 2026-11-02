// Package alertescalation 管理告警升级策略：命中的 FIRING 事件在指定分钟数后按步骤
// 向通知组生成升级通知任务；事件被认领或恢复后，未发送的步骤取消。
package alertescalation
