package i18n

import (
	"encoding/json"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/aide-family/moon/pkg/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func New(c Config) *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	suffix := ".toml"
	switch c.GetFormat() {
	case config.I18NFormat_JSON:
		suffix = ".json"
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	case config.I18NFormat_TOML:
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	default:
		panic("unknown i18n format")
	}
	dir := c.GetDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if path.Ext(file.Name()) != suffix {
			continue
		}
		bundle.MustLoadMessageFile(path.Join(dir, file.Name()))
	}
	return bundle
}
