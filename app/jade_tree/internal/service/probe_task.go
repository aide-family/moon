package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

func NewProbeTaskService(b *biz.ProbeTask) *ProbeTaskService {
	return &ProbeTaskService{bizProbeTask: b}
}

type ProbeTaskService struct {
	apiv1.UnimplementedProbeTaskServer
	bizProbeTask *biz.ProbeTask
}

func (s *ProbeTaskService) CreateProbeTask(ctx context.Context, req *apiv1.CreateProbeTaskRequest) (*apiv1.ProbeTaskItem, error) {
	in, err := bo.NewCreateProbeTaskBo(req, contextx.GetUserUID(ctx))
	if err != nil {
		return nil, err
	}
	item, err := s.bizProbeTask.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ProbeTaskItem(item), nil
}

func (s *ProbeTaskService) UpdateProbeTask(ctx context.Context, req *apiv1.UpdateProbeTaskRequest) (*apiv1.ProbeTaskItem, error) {
	in, err := bo.NewUpdateProbeTaskBo(req)
	if err != nil {
		return nil, err
	}
	item, err := s.bizProbeTask.Update(ctx, in)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ProbeTaskItem(item), nil
}

func (s *ProbeTaskService) DeleteProbeTask(ctx context.Context, req *apiv1.DeleteProbeTaskRequest) (*apiv1.DeleteProbeTaskReply, error) {
	if err := s.bizProbeTask.Delete(ctx, snowflake.ID(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteProbeTaskReply{}, nil
}

func (s *ProbeTaskService) GetProbeTask(ctx context.Context, req *apiv1.GetProbeTaskRequest) (*apiv1.ProbeTaskItem, error) {
	item, err := s.bizProbeTask.Get(ctx, snowflake.ID(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ProbeTaskItem(item), nil
}

func (s *ProbeTaskService) ListProbeTasks(ctx context.Context, req *apiv1.ListProbeTasksRequest) (*apiv1.ListProbeTasksReply, error) {
	listBo := bo.NewListProbeTasksBo(req)
	page, err := s.bizProbeTask.List(ctx, listBo)
	if err != nil {
		return nil, err
	}
	items := make([]*apiv1.ProbeTaskItem, 0, len(page.GetItems()))
	for _, item := range page.GetItems() {
		items = append(items, bo.ToAPIV1ProbeTaskItem(item))
	}
	return &apiv1.ListProbeTasksReply{
		Items:    items,
		Total:    page.GetTotal(),
		Page:     page.GetPage(),
		PageSize: page.GetPageSize(),
	}, nil
}

func (s *ProbeTaskService) UpdateProbeTaskStatus(ctx context.Context, req *apiv1.UpdateProbeTaskStatusRequest) (*apiv1.ProbeTaskItem, error) {
	in, err := bo.NewUpdateProbeTaskStatusBo(req)
	if err != nil {
		return nil, err
	}
	item, err := s.bizProbeTask.UpdateStatus(ctx, in)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ProbeTaskItem(item), nil
}

func (s *ProbeTaskService) CreatePingProbeTasks(ctx context.Context, req *apiv1.CreatePingProbeTasksRequest) (*apiv1.DispatchCreateProbeTasksReply, error) {
	reply, err := s.bizProbeTask.CreatePingProbeTasks(ctx, bo.NewCreatePingProbeTasksInput(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1DispatchCreateProbeTasksReply(reply), nil
}
