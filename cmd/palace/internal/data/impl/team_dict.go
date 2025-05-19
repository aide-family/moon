package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamDictRepo(d *data.Data, logger log.Logger) repository.TeamDict {
	return &teamDictImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team_dict")),
	}
}

type teamDictImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamDictImpl) Get(ctx context.Context, dictID uint32) (do.TeamDict, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	bizDictQuery := query.Dict
	wrappers := []gen.Condition{
		bizDictQuery.TeamID.Eq(teamID),
		bizDictQuery.ID.Eq(dictID),
	}
	dict, err := bizDictQuery.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		return nil, teamDictNotFound(err)
	}
	return dict, nil
}

func (t *teamDictImpl) Delete(ctx context.Context, dictID uint32) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizDictQuery := query.Dict
	wrappers := []gen.Condition{
		bizDictQuery.TeamID.Eq(teamID),
		bizDictQuery.ID.Eq(dictID),
	}
	_, err := bizDictQuery.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

func (t *teamDictImpl) Create(ctx context.Context, dict bo.Dict) error {
	query := getTeamBizQuery(ctx, t)
	dictDo := &team.Dict{
		Key:      dict.GetKey(),
		Value:    dict.GetValue(),
		Lang:     dict.GetLang(),
		Color:    dict.GetColor(),
		DictType: dict.GetType(),
		Status:   dict.GetStatus(),
	}
	dictDo.WithContext(ctx)
	bizDictQuery := query.Dict
	return bizDictQuery.WithContext(ctx).Create(dictDo)
}

func (t *teamDictImpl) Update(ctx context.Context, dict bo.Dict) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizDictQuery := query.Dict
	mutations := []field.AssignExpr{
		bizDictQuery.Key.Value(dict.GetKey()),
		bizDictQuery.Value.Value(dict.GetValue()),
		bizDictQuery.Lang.Value(dict.GetLang()),
		bizDictQuery.Color.Value(dict.GetColor()),
		bizDictQuery.DictType.Value(dict.GetType().GetValue()),
		bizDictQuery.Status.Value(dict.GetStatus().GetValue()),
	}
	wrappers := []gen.Condition{
		bizDictQuery.TeamID.Eq(teamID),
		bizDictQuery.ID.Eq(dict.GetID()),
	}
	_, err := bizDictQuery.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

func (t *teamDictImpl) UpdateStatus(ctx context.Context, req *bo.UpdateDictStatusReq) error {
	if len(req.DictIds) == 0 {
		return nil
	}
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizDictQuery := query.Dict
	wrappers := []gen.Condition{
		bizDictQuery.TeamID.Eq(teamID),
		bizDictQuery.ID.In(req.DictIds...),
	}
	_, err := bizDictQuery.WithContext(ctx).Where(wrappers...).
		UpdateColumnSimple(bizDictQuery.Status.Value(req.Status.GetValue()))
	return err
}

func (t *teamDictImpl) List(ctx context.Context, req *bo.ListDictReq) (*bo.ListDictReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizDictQuery := query.Dict
	wrapper := bizDictQuery.WithContext(ctx).Where(bizDictQuery.TeamID.Eq(teamID))
	if len(req.Langs) > 0 {
		wrapper = wrapper.Where(bizDictQuery.Lang.In(req.Langs...))
	}
	if len(req.DictTypes) > 0 {
		dictTypes := slices.Map(req.DictTypes, func(item vobj.DictType) int8 { return item.GetValue() })
		wrapper = wrapper.Where(bizDictQuery.DictType.In(dictTypes...))
	}
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizDictQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(bizDictQuery.Key.Like(req.Keyword))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	dictItems, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(dictItems, func(dictItem *team.Dict) do.TeamDict { return dictItem })
	return req.ToListReply(rows), nil
}

func (t *teamDictImpl) Select(ctx context.Context, req *bo.SelectDictReq) (*bo.SelectDictReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizDictQuery := query.Dict
	wrapper := bizDictQuery.WithContext(ctx).Where(bizDictQuery.TeamID.Eq(teamID))
	if len(req.Langs) > 0 {
		wrapper = wrapper.Where(bizDictQuery.Lang.In(req.Langs...))
	}
	if len(req.DictTypes) > 0 {
		dictTypes := slices.Map(req.DictTypes, func(item vobj.DictType) int8 { return item.GetValue() })
		wrapper = wrapper.Where(bizDictQuery.DictType.In(dictTypes...))
	}
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizDictQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(bizDictQuery.Key.Like(req.Keyword))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	selectColumns := []field.Expr{
		bizDictQuery.ID,
		bizDictQuery.Key,
		bizDictQuery.Value,
		bizDictQuery.Lang,
		bizDictQuery.Color,
		bizDictQuery.DictType,
		bizDictQuery.Status,
		bizDictQuery.DeletedAt,
	}
	dictItems, err := wrapper.Select(selectColumns...).Order(bizDictQuery.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return req.ToSelectReply(slices.Map(dictItems, func(dictItem *team.Dict) do.TeamDict { return dictItem })), nil
}

func (t *teamDictImpl) FindByIds(ctx context.Context, dictIds []uint32) ([]do.TeamDict, error) {
	if len(dictIds) == 0 {
		return nil, nil
	}
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.Dict
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID), mutation.ID.In(dictIds...))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.Dict) do.TeamDict { return row }), nil
}
