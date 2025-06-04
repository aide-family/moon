package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

// NewTimeEngine creates a time engine repository implementation
func NewTimeEngine(data *data.Data) repository.TimeEngine {
	return &timeEngineRepoImpl{
		Data: data,
	}
}

type timeEngineRepoImpl struct {
	*data.Data
}

// CreateTimeEngine creates a time engine
func (r *timeEngineRepoImpl) CreateTimeEngine(ctx context.Context, req *bo.SaveTimeEngineRequest) error {
	timeEngine := &team.TimeEngine{
		Name:   req.Name,
		Remark: req.Remark,
		Status: vobj.GlobalStatusEnable,
	}
	timeEngine.WithContext(ctx)
	bizQuery := getTeamBizQuery(ctx, r)
	if err := bizQuery.TimeEngine.WithContext(ctx).Create(timeEngine); err != nil {
		return err
	}

	if len(req.GetRules()) == 0 {
		return nil
	}
	rules := build.ToTimeEngineRules(ctx, req.GetRules())
	return bizQuery.TimeEngine.Rules.Model(timeEngine).Append(rules...)
}

// UpdateTimeEngine updates a time engine
func (r *timeEngineRepoImpl) UpdateTimeEngine(ctx context.Context, timeEngineId uint32, req *bo.SaveTimeEngineRequest) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngine := build.ToTimeEngine(ctx, req.GetTimeEngine())
	timeEngineMutation := bizMutation.TimeEngine
	wrappers := []gen.Condition{
		timeEngineMutation.ID.Eq(timeEngineId),
		timeEngineMutation.TeamID.Eq(teamId),
	}
	columns := []field.AssignExpr{
		timeEngineMutation.Name.Value(req.Name),
		timeEngineMutation.Remark.Value(req.Remark),
	}

	_, err := timeEngineMutation.WithContext(ctx).Where(wrappers...).UpdateSimple(columns...)
	if err != nil {
		return err
	}

	if len(req.GetRules()) > 0 {
		rules := build.ToTimeEngineRules(ctx, req.GetRules())
		return bizMutation.TimeEngine.Rules.Model(timeEngine).Replace(rules...)
	}

	return nil
}

// UpdateTimeEngineStatus updates the status of a time engine
func (r *timeEngineRepoImpl) UpdateTimeEngineStatus(ctx context.Context, req *bo.UpdateTimeEngineStatusRequest) error {
	if len(req.TimeEngineIds) == 0 {
		return nil
	}
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineMutation := bizMutation.TimeEngine
	wrappers := []gen.Condition{
		timeEngineMutation.ID.In(req.TimeEngineIds...),
		timeEngineMutation.TeamID.Eq(teamId),
	}
	columns := []field.AssignExpr{
		timeEngineMutation.Status.Value(req.Status.GetValue()),
	}
	_, err := timeEngineMutation.WithContext(ctx).Where(wrappers...).UpdateSimple(columns...)
	return err
}

// DeleteTimeEngine deletes a time engine
func (r *timeEngineRepoImpl) DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineMutation := bizMutation.TimeEngine
	wrappers := []gen.Condition{
		timeEngineMutation.ID.Eq(req.TimeEngineId),
		timeEngineMutation.TeamID.Eq(teamId),
	}
	_, err := timeEngineMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// GetTimeEngine gets the details of a time engine
func (r *timeEngineRepoImpl) GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (do.TimeEngine, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineQuery := bizQuery.TimeEngine
	timeEngine, err := timeEngineQuery.WithContext(ctx).
		Where(timeEngineQuery.ID.Eq(req.TimeEngineId), timeEngineQuery.TeamID.Eq(teamId)).
		Preload(field.Associations).
		First()
	if err != nil {
		return nil, teamTimeEngineNotFound(err)
	}

	return timeEngine, nil
}

// ListTimeEngine gets the list of time engines
func (r *timeEngineRepoImpl) ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) (*bo.ListTimeEngineReply, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineQuery := bizQuery.TimeEngine
	timeEngineWrapper := timeEngineQuery.Where(timeEngineQuery.TeamID.Eq(teamId))

	if !req.Status.IsUnknown() {
		timeEngineWrapper = timeEngineWrapper.Where(timeEngineQuery.Status.Eq(req.Status.GetValue()))
	}

	if validate.TextIsNotNull(req.Keyword) {
		or := []gen.Condition{
			timeEngineQuery.Name.Like(req.Keyword),
			timeEngineQuery.Remark.Like(req.Keyword),
		}
		timeEngineWrapper = timeEngineWrapper.Where(timeEngineQuery.Or(or...))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := timeEngineWrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		timeEngineWrapper = timeEngineWrapper.Limit(int(req.Limit)).Offset(int(req.Offset()))
	}
	timeEngineWrapper = timeEngineWrapper.Order(timeEngineQuery.CreatedAt.Desc())
	timeEngines, err := timeEngineWrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}

	dos := slices.Map(timeEngines, func(v *team.TimeEngine) do.TimeEngine { return v })
	return req.ToListReply(dos), nil
}

// SelectTimeEngine gets the list of time engines
func (r *timeEngineRepoImpl) SelectTimeEngine(ctx context.Context, req *bo.SelectTimeEngineRequest) (*bo.SelectTimeEngineReply, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineQuery := bizQuery.TimeEngine
	timeEngineWrapper := timeEngineQuery.Where(timeEngineQuery.TeamID.Eq(teamId))

	if !req.Status.IsUnknown() {
		timeEngineWrapper = timeEngineWrapper.Where(timeEngineQuery.Status.Eq(req.Status.GetValue()))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := timeEngineWrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
	}
	selectColumns := []field.Expr{
		timeEngineQuery.ID,
		timeEngineQuery.Name,
		timeEngineQuery.Remark,
		timeEngineQuery.Status,
		timeEngineQuery.DeletedAt,
	}
	timeEngineWrapper = timeEngineWrapper.Order(timeEngineQuery.CreatedAt.Desc()).Select(selectColumns...)
	timeEngines, err := timeEngineWrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(timeEngines, func(v *team.TimeEngine) do.TimeEngine { return v })
	return req.ToSelectReply(rows), nil
}

func (r *timeEngineRepoImpl) Find(ctx context.Context, ids ...uint32) ([]do.TimeEngine, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, r)
	timeEngineQuery := bizQuery.TimeEngine
	timeEngineWrapper := timeEngineQuery.Where(timeEngineQuery.TeamID.Eq(teamId))
	timeEngineWrapper = timeEngineWrapper.Where(timeEngineQuery.ID.In(ids...))
	timeEngines, err := timeEngineWrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(timeEngines, func(v *team.TimeEngine) do.TimeEngine { return v }), nil
}
