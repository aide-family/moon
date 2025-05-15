package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamNotice interface {
	Create(ctx context.Context, group bo.SaveNoticeGroup) error
	Update(ctx context.Context, group bo.SaveNoticeGroup) error
	UpdateStatus(ctx context.Context, req *bo.UpdateTeamNoticeGroupStatusRequest) error
	Delete(ctx context.Context, groupID uint32) error
	Get(ctx context.Context, groupID uint32) (do.NoticeGroup, error)
	List(ctx context.Context, req *bo.ListNoticeGroupReq) (*bo.ListNoticeGroupReply, error)
	FindByIds(ctx context.Context, groupIds []uint32) ([]do.NoticeGroup, error)

	FindLabelNotices(ctx context.Context, labelNoticeIds []uint32) ([]do.StrategyMetricRuleLabelNotice, error)
}
