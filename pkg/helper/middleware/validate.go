package middleware

import (
	"context"
	"errors"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/middleware"
	"google.golang.org/protobuf/proto"
)

// Validate 验证请求参数
func Validate(opts ...protovalidate.ValidatorOption) middleware.Middleware {
	validator := ValidateParams(opts...)
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

var errMsgMap = map[string]string{
	"value is required": "参数必须填写",
}

func getMsg(msg string) string {
	if v, ok := errMsgMap[msg]; ok {
		return v
	}
	return msg
}

type ValidateHandler func(ctx context.Context, req interface{}) error

// ValidateParams 验证请求参数
func ValidateParams(opts ...protovalidate.ValidatorOption) ValidateHandler {
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
			return merr.ErrorI18nAlert(ctx).WithCause(err)
		}

		if types.IsNil(validationError) || len(validationError.Violations) == 0 {
			return merr.ErrorI18nAlert(ctx)
		}

		errMap := make(map[string]string)
		for _, v := range validationError.Violations {
			field := v.GetFieldPath()
			if types.TextIsNull(field) {
				continue
			}
			msg := v.GetMessage()
			id := v.GetConstraintId()
			if !types.TextIsNull(id) {
				_msg := merr.GetI18nMessage(ctx, id)
				if !types.TextIsNull(_msg) {
					msg = _msg
				}
			}
			errMap[field] = getMsg(msg)
		}

		return merr.ErrorI18nAlert(ctx).WithMetadata(errMap)
	}
}
