package repository

import (
	"context"
	"io"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SSHSession 是一条从控制面到目标机器的 SSH 连接。
// 它只提供安装摇光所需的命令执行和文件上传，不承担业务判断。
type SSHSession interface {
	// Run 执行一条 shell 命令，返回合并后的 stdout 与 stderr、退出码。
	// stdin 会原样写到命令的标准输入，用于 sudo 密码和配置文件内容。
	Run(ctx context.Context, command string, stdin []byte) (output string, exitCode int, err error)
	// Upload 把 reader 的内容写到目标机器的 path，覆盖已有文件。
	Upload(ctx context.Context, path string, content io.Reader) error
	Close() error
}

// SSHDialer 用机器上保存的 SSH 凭据建立连接。
type SSHDialer interface {
	Dial(ctx context.Context, target bo.SSHTarget) (SSHSession, error)
}

// AgentPackageSource 取目标机器架构对应的摇光安装包。
// 本地安装包和远程下载都通过它暴露给用例层，用例层不关心包从哪里来。
type AgentPackageSource interface {
	// Fetch 返回安装包文件名和内容，调用方负责关闭。
	Fetch(ctx context.Context, arch string) (name string, content io.ReadCloser, err error)
}
