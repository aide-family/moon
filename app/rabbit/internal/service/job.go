package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewJobService(jobBiz *biz.Job) *JobService {
	return &JobService{
		jobBiz: jobBiz,
	}
}

type JobService struct {
	apiv1.UnimplementedJobServer

	jobBiz *biz.Job
}

func (s *JobService) SendMessage(ctx context.Context, req *apiv1.JobSendMessageRequest) (*apiv1.JobSendReply, error) {
	if err := s.jobBiz.AppendMessage(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.JobSendReply{Uid: req.Uid}, nil
}
