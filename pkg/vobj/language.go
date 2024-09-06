package vobj

// Language 字典类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Language -linecomment
type Language int

const (
	// LanguageUnknown 未知
	LanguageUnknown Language = iota // 未知

	// LanguageZHCN 中文
	LanguageZHCN // zh-CN

	// LanguageENUS 英文
	LanguageENUS // en-US
)

// ToLanguage 获取语言
func ToLanguage(s string) Language {
	switch s {
	case "zh-CN":
		return LanguageZHCN
	case "en-US":
		return LanguageENUS
	default:
		return LanguageUnknown
	}
}
