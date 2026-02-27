package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewJob(
	messageRepo repository.Message,
	helper *klog.Helper,
) *Job {
	return &Job{
		messageRepo: messageRepo,
		helper:      klog.NewHelper(klog.With(helper.Logger(), "biz", "job")),
	}
}

type Job struct {
	helper      *klog.Helper
	messageRepo repository.Message
}

func (e *Job) AppendMessage(ctx context.Context, messageUID snowflake.ID) error {
	return e.messageRepo.AppendMessage(ctx, messageUID)
}
