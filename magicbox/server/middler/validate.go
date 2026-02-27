package middler

import (
	"context"
	"errors"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"google.golang.org/protobuf/proto"

	mi18n "github.com/aide-family/magicbox/i18n"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
)

// Validate validate request parameters
func Validate(opts ...protovalidate.ValidatorOption) middleware.Middleware {
	validator := validateParams(opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			message, isOk := req.(proto.Message)
			if !isOk {
				return handler(ctx, req)
			}

			if err := validator(ctx, message); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}

func getMsg(ctx context.Context, constraintId string, msg string) string {
	if strutil.IsEmpty(constraintId) {
		return msg
	}
	if strings.EqualFold(constraintId, "required") {
		constraintId = "REQUIRED"
	}

	lang := mi18n.GetLanguage(ctx)
	bundle, err := mi18n.GetBundle()
	if err != nil {
		return msg
	}
	localize, localizeErr := i18n.NewLocalizer(bundle, lang).
		Localize(&i18n.LocalizeConfig{MessageID: constraintId})
	if pointer.IsNotNil(localizeErr) {
		log.Warnf("%s => validate error: %v", constraintId, localizeErr)
		return msg
	}

	return localize
}

// ValidateHandler validate handler
type ValidateHandler func(ctx context.Context, req interface{}) error

// validateParams validate params
func validateParams(opts ...protovalidate.ValidatorOption) ValidateHandler {
	validator, err := protovalidate.New(opts...)
	if err != nil {
		panic(err)
	}
	protovalidate.WithMessages()
	return func(ctx context.Context, req interface{}) error {
		message, isOk := req.(proto.Message)
		if !isOk {
			return nil
		}

		err = validator.Validate(message)
		if err == nil {
			return nil
		}
		var validationError *protovalidate.ValidationError
		if !errors.As(err, &validationError) {
			return merr.ErrorInternalServer("system error").WithCause(err)
		}

		if pointer.IsNil(validationError) || len(validationError.Violations) == 0 {
			return merr.ErrorInternalServer("system error")
		}

		errMap := make(map[string][]string)
		for _, v := range validationError.Violations {
			elements := v.Proto.Field.GetElements()
			fields := make([]string, 0, len(elements))
			for _, element := range elements {
				fields = append(fields, element.GetFieldName())
			}
			constraintId := v.Proto.GetRuleId()
			msg := v.Proto.GetMessage()
			if len(fields) == 0 {
				return merr.ErrorParams("%s", getMsg(ctx, constraintId, msg))
			}
			field := strings.Join(fields, ".")
			errMap[field] = append(errMap[field], getMsg(ctx, constraintId, msg))
		}

		msgMap := make(map[string]string)
		for k, v := range errMap {
			msgMap[k] = strings.Join(v, ",")
		}
		return merr.ErrorParams("params error").WithMetadata(msgMap)
	}
}
