package server

import (
	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v3/middleware"
	kvalidate "github.com/go-kratos/kratos/v3/middleware/validate"
	"google.golang.org/protobuf/proto"
)

// protoValidator 把 proto 上的 buf.validate 约束接到 Kratos 中间件。
// 生成代码本身不调用校验器，缺少这一步时所有 min_len、min_items、lte 等注解都不会生效。
func protoValidator() middleware.Middleware {
	return kvalidate.Validator(func(req any) error {
		msg, ok := req.(proto.Message)
		if !ok {
			return nil
		}
		return protovalidate.Validate(msg)
	})
}
