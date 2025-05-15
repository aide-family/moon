package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
)

func NewDict(
	teamDictRepo repository.TeamDict,
	logger log.Logger,
) *Dict {
	return &Dict{
		helper:       log.NewHelper(log.With(logger, "module", "biz.dict")),
		teamDictRepo: teamDictRepo,
	}
}

type Dict struct {
	helper *log.Helper

	teamDictRepo repository.TeamDict
}

func (d *Dict) SaveDict(ctx context.Context, req *bo.SaveDictReq) error {
	if req.DictID == 0 {
		return d.teamDictRepo.Create(ctx, req)
	}
	dictItem, err := d.teamDictRepo.Get(ctx, req.DictID)
	if err != nil {
		return err
	}
	return d.teamDictRepo.Update(ctx, req.WithUpdateParams(dictItem))
}

func (d *Dict) GetDict(ctx context.Context, req *bo.OperateOneDictReq) (do.TeamDict, error) {
	return d.teamDictRepo.Get(ctx, req.DictID)
}

func (d *Dict) UpdateDictStatus(ctx context.Context, req *bo.UpdateDictStatusReq) error {
	return d.teamDictRepo.UpdateStatus(ctx, req)
}

func (d *Dict) DeleteDict(ctx context.Context, req *bo.OperateOneDictReq) error {
	return d.teamDictRepo.Delete(ctx, req.DictID)
}

func (d *Dict) ListDict(ctx context.Context, req *bo.ListDictReq) (*bo.ListDictReply, error) {
	return d.teamDictRepo.List(ctx, req)
}
