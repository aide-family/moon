package service

import (
	"context"
	"io"
	"time"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// HookService hook服务
type HookService struct {
	hookapi.UnimplementedHookServer

	msgBiz *biz.MsgBiz
}

// NewHookService 创建hook服务
func NewHookService(msgBiz *biz.MsgBiz) *HookService {
	return &HookService{
		msgBiz: msgBiz,
	}
}

// SendMsg 发送消息
func (s *HookService) SendMsg(ctx context.Context, req *hookapi.SendMsgRequest) (*hookapi.SendMsgReply, error) {
	if err := s.msgBiz.SendMsg(ctx, &bo.SendMsgParams{
		Route: req.Route,
		Data:  []byte(req.JsonData),
	}); !types.IsNil(err) {
		return nil, err
	}
	return &hookapi.SendMsgReply{
		Msg:  "ok",
		Code: 0,
		Time: types.NewTime(time.Now()).String(),
	}, nil
}

// HookSendMsgHTTPHandler hook发送消息http handler
func (s *HookService) HookSendMsgHTTPHandler() func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in hookapi.SendMsgRequest
		if err := ctx.BindVars(&in); !types.IsNil(err) {
			return err
		}

		body := ctx.Request().Body
		all, err := io.ReadAll(body)
		if !types.IsNil(err) {
			return err
		}
		in.JsonData = string(all)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.SendMsg(ctx, req.(*hookapi.SendMsgRequest))
		})
		out, err := h(ctx, &in)
		if !types.IsNil(err) {
			return err
		}
		reply := out.(*hookapi.SendMsgReply)
		return ctx.Result(200, reply)
	}
}
