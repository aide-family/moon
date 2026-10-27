// Package cluster 管理租户接入的 Kubernetes 集群，并通过保存的凭据访问集群资源。
//
// 集群接入、替换凭据和删除由 Registry 负责；资源的只读查询由 Explorer 负责，
// 改调度、改副本数、滚动重启和删 Pod 由 Operator 负责。所有入口都在这里
// 校验租户权限，并确认集群属于当前租户。
package cluster
