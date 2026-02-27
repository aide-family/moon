package i18n

import "golang.org/x/text/language"

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
)

type Config interface {
	GetFormat() Format
	GetDir() string
	GetLang() language.Tag
}

func (f Format) String() string {
	return string(f)
}

func (f Format) Ext() string {
	return "." + f.String()
}
