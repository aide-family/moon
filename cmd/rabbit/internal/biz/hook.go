package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/go-kratos/kratos/v2/log"
)

func NewHook(sendRepo repository.Send, logger log.Logger) *Hook {
	return &Hook{
		helper:   log.NewHelper(log.With(logger, "module", "biz.hook")),
		sendRepo: sendRepo,
	}
}

type Hook struct {
	helper *log.Helper

	sendRepo repository.Send
}

func (h *Hook) Send(ctx context.Context, params bo.SendHookParams) error {
	if len(params.GetConfigs()) == 0 {
		return merr.ErrorParamsError("No hook configuration is available")
	}
	return h.sendRepo.Hook(ctx, params)
}
