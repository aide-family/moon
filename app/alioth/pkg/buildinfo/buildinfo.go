// Package buildinfo 保存由构建注入的版本号。
// 入口程序在启动时把 ldflags 注入的 main.version 写进来，
// 这样 biz 层不用为了一个字符串改动 Wire 的构造参数。
package buildinfo

// Version 是控制面构建版本，未注入时为 dev。
var Version = "dev"
