// Package i18n provides a i18n service.
package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert/yaml"

	"github.com/aide-family/magicbox/strutil/cnst"
)

// New creates a new i18n bundle.
func New(config Config) (*i18n.Bundle, error) {
	lang := config.GetLang()
	newBundle := i18n.NewBundle(lang)
	format := config.GetFormat()

	switch format {
	case FormatJSON:
		newBundle.RegisterUnmarshalFunc(format.String(), json.Unmarshal)
	case FormatYAML:
		newBundle.RegisterUnmarshalFunc(format.String(), yaml.Unmarshal)
	case FormatTOML:
		newBundle.RegisterUnmarshalFunc(format.String(), toml.Unmarshal)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	dir := config.GetDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ext := format.Ext()
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ext) {
			continue
		}
		newBundle.MustLoadMessageFile(path.Join(dir, file.Name()))
	}
	return newBundle, nil
}

var bundle *i18n.Bundle

// RegisterBundle registers a bundle.
func RegisterBundle(b *i18n.Bundle) {
	bundle = b
}

// RegisterBundleWithDefault registers a bundle with a default config.
func RegisterBundleWithDefault(config Config) (err error) {
	bundle, err = New(config)
	return
}

// GetBundle returns the registered bundle.
func GetBundle() (*i18n.Bundle, error) {
	if bundle == nil {
		return nil, fmt.Errorf("bundle is not registered")
	}
	return bundle, nil
}

// Message returns a localized message.
func Message(bundle *i18n.Bundle, lang string, key string, args ...interface{}) (string, error) {
	localize, err := i18n.NewLocalizer(bundle, lang).
		Localize(&i18n.LocalizeConfig{MessageID: key, TemplateData: args})
	if err != nil {
		return "", err
	}
	return localize, nil
}

// MessageX returns a localized message.
func MessageX(bundle *i18n.Bundle, lang string, key string, args ...interface{}) string {
	localize, _ := Message(bundle, lang, key, args...)
	return localize
}

func GetLanguage(ctx context.Context) string {
	if md, ok := metadata.FromServerContext(ctx); ok {
		return md.Get(cnst.HTTHeaderLang)
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return tr.RequestHeader().Get(cnst.HTTHeaderLang)
}
