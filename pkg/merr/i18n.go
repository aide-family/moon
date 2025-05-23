package merr

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/validate"
)

type Localizer interface {
	Localize(ctx context.Context, err error) error
}

var (
	globalLocalizer Localizer
	once            sync.Once
	bundle          *i18n.Bundle
	bundleOnce      sync.Once
)

func RegisterGlobalLocalizer(localizer Localizer) {
	once.Do(func() {
		globalLocalizer = localizer
	})
}

func RegisterBundle(b *i18n.Bundle) {
	bundleOnce.Do(func() {
		bundle = b
	})
}

func GetBundle() *i18n.Bundle {
	return bundle
}

func I18n() middleware.Middleware {
	if validate.IsNil(globalLocalizer) {
		if validate.IsNil(bundle) {
			panic("please register bundle first or use RegisterGlobalLocalizer")
		}
		globalLocalizer = NewLocalizer(bundle)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			reply, err = handler(ctx, req)
			if err != nil {
				return nil, globalLocalizer.Localize(ctx, err)
			}
			return reply, nil
		}
	}
}

func NewLocalizer(bundle *i18n.Bundle) Localizer {
	return &localizer{
		bundle: bundle,
	}
}

type localizer struct {
	bundle *i18n.Bundle
}

func (l *localizer) Localize(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	se := errors.FromError(err)
	if se == nil {
		return err
	}

	lang := GetLanguage(ctx)
	localize, localizeErr := i18n.NewLocalizer(l.bundle, lang).
		Localize(&i18n.LocalizeConfig{MessageID: se.GetReason()})
	if localizeErr != nil {
		return se.WithCause(localizeErr)
	}

	md := kv.NewStringMap(se.GetMetadata())
	md.Set("__message__", se.GetMessage())
	return errors.New(int(se.GetCode()), se.GetReason(), localize).WithMetadata(md).WithCause(err)
}

func GetLanguage(ctx context.Context) string {
	if md, ok := metadata.FromServerContext(ctx); ok {
		return md.Get(cnst.HttpHeaderAcceptLang)
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return tr.RequestHeader().Get(cnst.HttpHeaderAcceptLang)
}
