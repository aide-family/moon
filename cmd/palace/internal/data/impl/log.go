package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewOperateLogRepo(d *data.Data, logger log.Logger) repository.OperateLog {
	return &operateLogImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.operate_log")),
	}
}

type operateLogImpl struct {
	*data.Data
	helper *log.Helper
}

func (o *operateLogImpl) OperateLog(ctx context.Context, log *bo.AddOperateLog) error {
	operateLog := &system.OperateLog{
		OperateType:     log.OperateType,
		OperateModule:   log.OperateModule,
		OperateDataID:   log.OperateDataID,
		OperateDataName: log.OperateDataName,
		Title:           log.Title,
		Before:          log.Before,
		After:           log.After,
		IP:              log.IP,
	}
	operateLog.WithContext(ctx)
	operateLogMutation := getMainQuery(ctx, o).OperateLog
	return operateLogMutation.WithContext(ctx).Create(operateLog)
}

func (o *operateLogImpl) List(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	if validate.IsNil(req) {
		return nil, nil
	}
	operateLogQuery := getMainQuery(ctx, o).OperateLog
	wrapper := operateLogQuery.WithContext(ctx)

	if len(req.OperateTypes) > 0 {
		wrapper = wrapper.Where(operateLogQuery.OperateType.In(slices.Map(req.OperateTypes, func(operateType vobj.OperateType) int8 { return operateType.GetValue() })...))
	}
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
	return req.ToOperateLogListReply(operateLogs), nil
}

func (o *operateLogImpl) TeamOperateLog(ctx context.Context, log *bo.AddOperateLog) error {
	operateLog := &team.OperateLog{
		OperateType:     log.OperateType,
		OperateModule:   log.OperateModule,
		OperateDataID:   log.OperateDataID,
		OperateDataName: log.OperateDataName,
		Title:           log.Title,
		Before:          log.Before,
		After:           log.After,
		IP:              log.IP,
	}
	operateLog.WithContext(ctx)
	bizMutation := getTeamBizQuery(ctx, o)
	operateLogMutation := bizMutation.OperateLog
	return operateLogMutation.WithContext(ctx).Create(operateLog)
}

func (o *operateLogImpl) TeamList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	if validate.IsNil(req) {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, o)
	operateLogQuery := bizQuery.OperateLog
	wrapper := operateLogQuery.WithContext(ctx).Where(operateLogQuery.TeamID.Eq(teamId))

	if len(req.OperateTypes) > 0 {
		wrapper = wrapper.Where(operateLogQuery.OperateType.In(slices.Map(req.OperateTypes, func(operateType vobj.OperateType) int8 { return operateType.GetValue() })...))
	}
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
	return req.ToTeamOperateLogListReply(operateLogs), nil
}
