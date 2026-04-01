package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
)

type ProbeTask struct {
	repo   repository.ProbeTask
	helper *klog.Helper
}

func NewProbeTask(repo repository.ProbeTask, helper *klog.Helper) *ProbeTask {
	return &ProbeTask{repo: repo, helper: helper}
}

func (p *ProbeTask) Create(ctx context.Context, in *bo.CreateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	return p.repo.Create(ctx, in)
}

func (p *ProbeTask) Update(ctx context.Context, in *bo.UpdateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	return p.repo.Update(ctx, in)
}

func (p *ProbeTask) Delete(ctx context.Context, uid snowflake.ID) error {
	return p.repo.Delete(ctx, uid)
}

func (p *ProbeTask) Get(ctx context.Context, uid snowflake.ID) (*bo.ProbeTaskItemBo, error) {
	return p.repo.Get(ctx, uid)
}

func (p *ProbeTask) List(ctx context.Context, req *bo.ListProbeTasksBo) (*bo.PageResponseBo[*bo.ProbeTaskItemBo], error) {
	return p.repo.List(ctx, req)
}
