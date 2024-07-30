package middleware

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/aide-family/moon/api/merr"

	"github.com/BurntSushi/toml"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// LocaleFS 加载国际化文件
//
//go:embed locale/active.*.toml
var LocaleFS embed.FS

// I18N 国际化中间件
func I18N() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				accept := tr.RequestHeader().Get("Accept-language")
				languages, _, err := language.ParseAcceptLanguage(strings.Split(accept, "-")[0])
				if err != nil {
					log.Warnw("err", err)
				}
				if len(languages) == 0 {
					languages = append(languages, language.Chinese)
				}
				lastLang := languages[len(languages)-1]
				bundle := i18n.NewBundle(lastLang)
				bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
				_, err = bundle.LoadMessageFileFS(LocaleFS, fmt.Sprintf("locale/active.%s.toml", lastLang.String()))
				if err != nil {
					log.Warnw("err", err)
				}

				ctx = merr.WithLocalize(ctx, i18n.NewLocalizer(bundle, lastLang.String()))
			}
			return handler(ctx, req)
		}
	}
}
