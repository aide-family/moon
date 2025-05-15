package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamDict interface {
	Get(ctx context.Context, dictID uint32) (do.TeamDict, error)
	FindByIds(ctx context.Context, dictIds []uint32) ([]do.TeamDict, error)
	Delete(ctx context.Context, dictID uint32) error
	Create(ctx context.Context, dict bo.Dict) error
	Update(ctx context.Context, dict bo.Dict) error
	UpdateStatus(ctx context.Context, req *bo.UpdateDictStatusReq) error
	List(ctx context.Context, req *bo.ListDictReq) (*bo.ListDictReply, error)
}
