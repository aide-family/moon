package i18n

import "github.com/aide-family/moon/pkg/config"

type Config interface {
	GetFormat() config.I18NFormat
	GetDir() string
}
