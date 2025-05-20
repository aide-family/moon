package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewOperateLogRepo(d *data.Data, logger log.Logger) repository.OperateLog {
	return &operateLogRepoImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.operate_log")),
	}
}

type operateLogRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func (o *operateLogRepoImpl) OperateLog(ctx context.Context, log *bo.AddOperateLog) error {
	operateLog := &system.OperateLog{}
	operateLog.WithContext(ctx)
	operateLogMutation := getMainQuery(ctx, o).OperateLog
	return operateLogMutation.WithContext(ctx).Create(operateLog)
}

func (o *operateLogRepoImpl) List(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	if validate.IsNil(req) {
		return nil, nil
	}
	operateLogQuery := getMainQuery(ctx, o).OperateLog
	wrapper := operateLogQuery.WithContext(ctx)

	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			operateLogQuery.Before.Like(req.Keyword),
			operateLogQuery.After.Like(req.Keyword),
			operateLogQuery.Title.Like(req.Keyword),
			operateLogQuery.OperateDataName.Like(req.Keyword),
			operateLogQuery.IP.Like(req.Keyword),
		}
		wrapper = wrapper.Where(operateLogQuery.Or(ors...))
	}
	if req.UserID > 0 {
		wrapper = wrapper.Where(operateLogQuery.CreatorID.Eq(req.UserID))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	operateLogs, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(operateLogs, func(operateLog *system.OperateLog) do.OperateLog { return operateLog })
	return req.ToListReply(rows), nil
}

func (o *operateLogRepoImpl) TeamOperateLog(ctx context.Context, log *bo.AddOperateLog) error {
	operateLog := &team.OperateLog{}
	operateLog.WithContext(ctx)
	bizMutation := getTeamBizQuery(ctx, o)
	operateLogMutation := bizMutation.OperateLog
	return operateLogMutation.WithContext(ctx).Create(operateLog)
}

func (o *operateLogRepoImpl) TeamList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	if validate.IsNil(req) {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, o)
	operateLogQuery := bizQuery.OperateLog
	wrapper := operateLogQuery.WithContext(ctx).Where(operateLogQuery.TeamID.Eq(teamId))

	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			operateLogQuery.Before.Like(req.Keyword),
			operateLogQuery.After.Like(req.Keyword),
			operateLogQuery.Title.Like(req.Keyword),
			operateLogQuery.OperateDataName.Like(req.Keyword),
			operateLogQuery.IP.Like(req.Keyword),
		}
		wrapper = wrapper.Where(operateLogQuery.Or(ors...))
	}
	if req.UserID > 0 {
		wrapper = wrapper.Where(operateLogQuery.CreatorID.Eq(req.UserID))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	operateLogs, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(operateLogs, func(operateLog *team.OperateLog) do.OperateLog { return operateLog })
	return req.ToListReply(rows), nil
}
