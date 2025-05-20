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
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamHook(data *data.Data, logger log.Logger) repository.TeamHook {
	return &teamHookRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team_hook")),
	}
}

type teamHookRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamHookRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.NoticeHook, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)

	hookQuery := bizQuery.NoticeHook
	wrapper := []gen.Condition{
		hookQuery.TeamID.Eq(teamId),
		hookQuery.ID.In(ids...),
	}
	hooks, err := hookQuery.WithContext(ctx).Where(wrapper...).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(hooks, func(hook *team.NoticeHook) do.NoticeHook { return hook }), nil
}

func (t *teamHookRepoImpl) Create(ctx context.Context, hook bo.NoticeHook) error {
	noticeHook := &team.NoticeHook{
		Name:    hook.GetName(),
		Remark:  hook.GetRemark(),
		Status:  vobj.GlobalStatusEnable,
		URL:     hook.GetURL(),
		Method:  hook.GetMethod(),
		Secret:  crypto.String(hook.GetSecret()),
		Headers: crypto.NewObject(hook.GetHeaders()),
		APP:     hook.GetApp(),
	}
	noticeHook.WithContext(ctx)

	query := getTeamBizQuery(ctx, t)

	return query.NoticeHook.WithContext(ctx).Create(noticeHook)
}

func (t *teamHookRepoImpl) Update(ctx context.Context, hook bo.NoticeHook) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		query.NoticeHook.ID.Eq(hook.GetID()),
		query.NoticeHook.TeamID.Eq(teamID),
	}

	hookQuery := query.NoticeHook
	mutations := []field.AssignExpr{
		hookQuery.Name.Value(hook.GetName()),
		hookQuery.Remark.Value(hook.GetRemark()),
		hookQuery.Method.Value(hook.GetMethod().GetValue()),
		hookQuery.Headers.Value(crypto.NewObject(hook.GetHeaders())),
		hookQuery.APP.Value(hook.GetApp().GetValue()),
		hookQuery.Secret.Value(crypto.String(hook.GetSecret())),
		hookQuery.URL.Value(hook.GetURL()),
	}

	_, err := hookQuery.WithContext(ctx).Where(wrapper...).UpdateSimple(mutations...)
	return err
}

func (t *teamHookRepoImpl) UpdateStatus(ctx context.Context, req *bo.UpdateTeamNoticeHookStatusRequest) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	wrapper := []gen.Condition{
		query.NoticeHook.ID.Eq(req.HookID),
		query.NoticeHook.TeamID.Eq(teamID),
	}

	hookQuery := query.NoticeHook
	_, err := hookQuery.WithContext(ctx).
		Where(wrapper...).
		UpdateSimple(hookQuery.Status.Value(req.Status.GetValue()))
	return err
}

func (t *teamHookRepoImpl) Delete(ctx context.Context, hookID uint32) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	wrapper := []gen.Condition{
		query.NoticeHook.ID.Eq(hookID),
		query.NoticeHook.TeamID.Eq(teamID),
	}

	hookQuery := query.NoticeHook
	_, err := hookQuery.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (t *teamHookRepoImpl) Get(ctx context.Context, hookID uint32) (do.NoticeHook, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	hookQuery := query.NoticeHook
	wrapper := []gen.Condition{
		hookQuery.ID.Eq(hookID),
		hookQuery.TeamID.Eq(teamID),
	}

	hook, err := hookQuery.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, hookNotFound(err)
	}
	return hook, nil
}

func (t *teamHookRepoImpl) List(ctx context.Context, req *bo.ListTeamNoticeHookRequest) (*bo.ListTeamNoticeHookReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	hookQuery := query.NoticeHook
	wrapper := hookQuery.WithContext(ctx).Where(hookQuery.TeamID.Eq(teamID))
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(hookQuery.Status.Eq(req.Status.GetValue()))
	}
	if len(req.Apps) > 0 {
		wrapper = wrapper.Where(hookQuery.APP.In(slices.Map(req.Apps, func(app vobj.HookApp) int8 { return app.GetValue() })...))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(hookQuery.Name.Like(req.Keyword))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}

	noticeHooks, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(noticeHooks, func(noticeHook *team.NoticeHook) do.NoticeHook { return noticeHook })
	return req.ToListReply(rows), nil
}

func (t *teamHookRepoImpl) Select(ctx context.Context, req *bo.TeamNoticeHookSelectRequest) (*bo.TeamNoticeHookSelectReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)

	hookQuery := query.NoticeHook
	wrapper := hookQuery.WithContext(ctx).Where(hookQuery.TeamID.Eq(teamID))
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(hookQuery.Status.Eq(req.Status.GetValue()))
	}
	if len(req.Apps) > 0 {
		wrapper = wrapper.Where(hookQuery.APP.In(slices.Map(req.Apps, func(app vobj.HookApp) int8 { return app.GetValue() })...))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(hookQuery.Name.Like(req.Keyword))
	}
	if !validate.TextIsNull(req.URL) {
		wrapper = wrapper.Where(hookQuery.URL.Eq(req.URL))
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
		hookQuery.ID,
		hookQuery.Name,
		hookQuery.Remark,
		hookQuery.Status,
		hookQuery.Method,
		hookQuery.APP,
		hookQuery.DeletedAt,
	}
	noticeHooks, err := wrapper.Select(selectColumns...).Order(hookQuery.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(noticeHooks, func(noticeHook *team.NoticeHook) do.NoticeHook { return noticeHook })
	return req.ToSelectReply(rows), nil
}
