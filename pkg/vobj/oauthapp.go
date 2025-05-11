package vobj

// OAuthAPP 系统全局角色
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=OAuthAPP -linecomment
type OAuthAPP int

const (
	// OAuthAPPAll 未知
	OAuthAPPAll OAuthAPP = iota // 未知

	// OAuthAPPGithub Github
	OAuthAPPGithub // Github

	// OAuthAPPGitee Gitee
	OAuthAPPGitee // Gitee
	// OAuthFeiShu FeiShu
	OAuthFeiShu
)
