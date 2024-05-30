package service

import (
	"context"
	"io"
	"time"

	pb "github.com/aide-cloud/moon/api/rabbit/hook"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-cloud/moon/pkg/types"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type HookService struct {
	pb.UnimplementedHookServer

	msgBiz *biz.MsgBiz
}

func NewHookService(msgBiz *biz.MsgBiz) *HookService {
	return &HookService{
		msgBiz: msgBiz,
	}
}

func (s *HookService) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgReply, error) {
	if err := s.msgBiz.SendMsg(ctx, &bo.SendMsgParams{
		Route: req.Route,
		Data:  []byte(req.JsonData),
	}); !types.IsNil(err) {
		return nil, err
	}
	return &pb.SendMsgReply{
		Msg:  "ok",
		Code: 0,
		Time: types.NewTime(time.Now()).String(),
	}, nil
}

func (s *HookService) HookSendMsgHTTPHandler() func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in pb.SendMsgRequest
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
			return s.SendMsg(ctx, req.(*pb.SendMsgRequest))
		})
		out, err := h(ctx, &in)
		if !types.IsNil(err) {
			return err
		}
		reply := out.(*pb.SendMsgReply)
		return ctx.Result(200, reply)
	}
}
