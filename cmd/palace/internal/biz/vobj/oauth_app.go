package vobj

// OAuthAPP oauth app
//
//go:generate stringer -type=OAuthAPP -linecomment -output=oauth_app.string.go
type OAuthAPP int8

const (
	OAuthAPPUnknown OAuthAPP = iota // unknown
	OAuthAPPGithub                  // github
	OAuthAPPGitee                   // gitee
)
