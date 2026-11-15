// Package registry 管理镜像仓库连接、远端镜像浏览与镜像同步任务。
//
// Registry 负责连接的增删改查与连通性测试，Explorer 负责只读的仓库、Tag 与平台，
// Synchronizer 负责同步任务的创建、调度认领与执行。远端元数据不落库，
// 同步任务落 PostgreSQL 并由 alioth 调度器按租约执行。
package registry
