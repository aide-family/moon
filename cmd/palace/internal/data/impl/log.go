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
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
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

func (o *operateLogRepoImpl) CreateLog(ctx context.Context, operateLogParams *bo.OperateLogParams) error {
	ctx = permission.WithUserIDContext(ctx, operateLogParams.UserID)
	operateLog := &system.OperateLog{
		CreatorModel: do.CreatorModel{
			CreatorID: operateLogParams.UserID,
		},
		Operation:     operateLogParams.Operation,
		MenuID:        operateLogParams.MenuID,
		MenuName:      operateLogParams.MenuName,
		Request:       operateLogParams.Request,
		Error:         operateLogParams.Error,
		OriginRequest: operateLogParams.OriginRequest,
		Duration:      operateLogParams.Duration,
		RequestTime:   operateLogParams.RequestTime,
		ReplyTime:     operateLogParams.ReplyTime,
		ClientIP:      operateLogParams.ClientIP,
		UserAgent:     operateLogParams.UserAgent,
		UserBaseInfo:  operateLogParams.UserBaseInfo,
	}
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
			operateLogQuery.Operation.Like(req.Keyword),
			operateLogQuery.MenuName.Like(req.Keyword),
			operateLogQuery.Request.Like(req.Keyword),
			operateLogQuery.Error.Like(req.Keyword),
			operateLogQuery.OriginRequest.Like(req.Keyword),
			operateLogQuery.ClientIP.Like(req.Keyword),
			operateLogQuery.UserAgent.Like(req.Keyword),
			operateLogQuery.UserBaseInfo.Like(req.Keyword),
		}
		wrapper = wrapper.Where(operateLogQuery.Or(ors...))
	}
	if req.UserID > 0 {
		wrapper = wrapper.Where(operateLogQuery.CreatorID.Eq(req.UserID))
	}
	if len(req.TimeRange) == 2 {
		wrapper = wrapper.Where(operateLogQuery.RequestTime.Between(req.TimeRange[0], req.TimeRange[1]))
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

func (o *operateLogRepoImpl) TeamCreateLog(ctx context.Context, operateLogParams *bo.OperateLogParams) error {
	ctx = permission.WithUserIDContext(ctx, operateLogParams.UserID)
	ctx = permission.WithTeamIDContext(ctx, operateLogParams.TeamID)
	operateLog := &team.OperateLog{
		TeamModel: do.TeamModel{
			TeamID: operateLogParams.TeamID,
			CreatorModel: do.CreatorModel{
				CreatorID: operateLogParams.UserID,
			},
		},
		Operation:     operateLogParams.Operation,
		MenuID:        operateLogParams.MenuID,
		MenuName:      operateLogParams.MenuName,
		Request:       operateLogParams.Request,
		Error:         operateLogParams.Error,
		OriginRequest: operateLogParams.OriginRequest,
		Duration:      operateLogParams.Duration,
		RequestTime:   operateLogParams.RequestTime,
		ReplyTime:     operateLogParams.ReplyTime,
		ClientIP:      operateLogParams.ClientIP,
		UserAgent:     operateLogParams.UserAgent,
		UserBaseInfo:  operateLogParams.UserBaseInfo,
	}
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
			operateLogQuery.Operation.Like(req.Keyword),
			operateLogQuery.MenuName.Like(req.Keyword),
			operateLogQuery.Request.Like(req.Keyword),
			operateLogQuery.Error.Like(req.Keyword),
			operateLogQuery.OriginRequest.Like(req.Keyword),
			operateLogQuery.ClientIP.Like(req.Keyword),
			operateLogQuery.UserAgent.Like(req.Keyword),
			operateLogQuery.UserBaseInfo.Like(req.Keyword),
		}
		wrapper = wrapper.Where(operateLogQuery.Or(ors...))
	}
	if req.UserID > 0 {
		wrapper = wrapper.Where(operateLogQuery.CreatorID.Eq(req.UserID))
	}
	if len(req.TimeRange) == 2 {
		wrapper = wrapper.Where(operateLogQuery.RequestTime.Between(req.TimeRange[0], req.TimeRange[1]))
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
