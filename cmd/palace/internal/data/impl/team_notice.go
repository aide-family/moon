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
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamNotice(data *data.Data, logger log.Logger) repository.TeamNotice {
	return &teamNoticeRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team_notice")),
	}
}

type teamNoticeRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamNoticeRepoImpl) List(ctx context.Context, req *bo.ListNoticeGroupReq) (*bo.ListNoticeGroupReply, error) {
	if validate.IsNil(req) {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	noticeGroupQuery := bizQuery.NoticeGroup
	wrapper := noticeGroupQuery.WithContext(ctx).Where(noticeGroupQuery.TeamID.Eq(teamId))
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(noticeGroupQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(noticeGroupQuery.Name.Like(req.Keyword))
	}
	if len(req.MemberIds) > 0 {
		noticeMemberQuery := bizQuery.NoticeMember
		var noticeGroupIds []uint32
		if err := noticeMemberQuery.WithContext(ctx).Select(noticeMemberQuery.NoticeGroupID).Scan(&noticeGroupIds); err != nil {
			return nil, err
		}
		if len(noticeGroupIds) == 0 {
			return req.ToListReply(nil), nil
		}
		wrapper = wrapper.Where(noticeGroupQuery.ID.In(noticeGroupIds...))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	noticeGroups, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(noticeGroups, func(noticeGroup *team.NoticeGroup) do.NoticeGroup { return noticeGroup })
	return req.ToListReply(rows), nil
}

func (t *teamNoticeRepoImpl) Create(ctx context.Context, group bo.SaveNoticeGroup) error {
	members := slices.MapFilter(group.GetNoticeMembers(), func(member *bo.SaveNoticeMemberItem) (*team.NoticeMember, bool) {
		if validate.IsNil(member) || member.MemberID <= 0 {
			return nil, false
		}
		item := &team.NoticeMember{
			UserID:        member.UserID,
			NoticeType:    member.NoticeType,
			TeamModel:     do.TeamModel{},
			NoticeGroupID: 0,
			DutyCycle:     []*team.TimeEngine{},
		}
		if dutyCycles := member.GetDutyCycles(); len(dutyCycles) > 0 {
			item.DutyCycle = slices.Map(dutyCycles, func(dutyCycle do.TimeEngine) *team.TimeEngine {
				item := build.ToTimeEngine(ctx, dutyCycle)
				item.WithContext(ctx)
				return item
			})
		}
		item.WithContext(ctx)
		return item, true
	})
	noticeGroupDo := &team.NoticeGroup{
		Name:    group.GetName(),
		Remark:  group.GetRemark(),
		Status:  group.GetStatus(),
		Members: members,
	}
	if validate.IsNotNil(group.GetEmailConfig()) {
		noticeGroupDo.EmailConfigID = group.GetEmailConfig().GetID()
	}
	if validate.IsNotNil(group.GetSMSConfig()) {
		noticeGroupDo.SMSConfigID = group.GetSMSConfig().GetID()
	}
	noticeGroupDo.WithContext(ctx)
	bizMutation := getTeamBizQuery(ctx, t)
	if err := bizMutation.NoticeGroup.WithContext(ctx).Create(noticeGroupDo); err != nil {
		return err
	}

	hooks := slices.MapFilter(group.GetHooks(), func(hook do.NoticeHook) (*team.NoticeHook, bool) {
		if validate.IsNil(hook) || hook.GetID() <= 0 {
			return nil, false
		}
		hookItem := &team.NoticeHook{
			TeamModel: do.TeamModel{
				CreatorModel: do.CreatorModel{
					BaseModel: do.BaseModel{ID: hook.GetID()},
				},
			},
		}
		hookItem.WithContext(ctx)
		return hookItem, true
	})
	if len(hooks) > 0 {
		if err := bizMutation.NoticeGroup.Hooks.Model(noticeGroupDo).Append(hooks...); err != nil {
			return err
		}
	}

	return nil
}

func (t *teamNoticeRepoImpl) Update(ctx context.Context, group bo.SaveNoticeGroup) error {
	if validate.IsNil(group) {
		return nil
	}
	noticeGroupMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		noticeGroupMutation.NoticeGroup.TeamID.Eq(teamId),
		noticeGroupMutation.NoticeGroup.ID.Eq(group.GetID()),
	}
	mutations := []field.AssignExpr{
		noticeGroupMutation.NoticeGroup.Name.Value(group.GetName()),
		noticeGroupMutation.NoticeGroup.Remark.Value(group.GetRemark()),
		noticeGroupMutation.NoticeGroup.Status.Value(group.GetStatus().GetValue()),
	}
	if validate.IsNotNil(group.GetEmailConfig()) {
		mutations = append(mutations, noticeGroupMutation.NoticeGroup.EmailConfigID.Value(group.GetEmailConfig().GetID()))
	}
	if validate.IsNotNil(group.GetSMSConfig()) {
		mutations = append(mutations, noticeGroupMutation.NoticeGroup.SMSConfigID.Value(group.GetSMSConfig().GetID()))
	}
	_, err := noticeGroupMutation.NoticeGroup.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutations...)
	if err != nil {
		return err
	}
	groupDo := &team.NoticeGroup{
		TeamModel: do.TeamModel{
			CreatorModel: do.CreatorModel{
				BaseModel: do.BaseModel{ID: group.GetID()},
			},
		},
	}
	hooks := slices.MapFilter(group.GetHooks(), func(hook do.NoticeHook) (*team.NoticeHook, bool) {
		if validate.IsNil(hook) || hook.GetID() <= 0 {
			return nil, false
		}
		hookItem := &team.NoticeHook{
			TeamModel: do.TeamModel{
				CreatorModel: do.CreatorModel{
					BaseModel: do.BaseModel{ID: hook.GetID()},
				},
			},
		}
		hookItem.WithContext(ctx)
		return hookItem, true
	})
	members := slices.MapFilter(group.GetNoticeMembers(), func(member *bo.SaveNoticeMemberItem) (*team.NoticeMember, bool) {
		if validate.IsNil(member) || member.MemberID <= 0 {
			return nil, false
		}
		item := &team.NoticeMember{
			UserID:     member.UserID,
			NoticeType: member.NoticeType,
		}
		if dutyCycles := member.GetDutyCycles(); len(dutyCycles) > 0 {
			item.DutyCycle = slices.Map(dutyCycles, func(dutyCycle do.TimeEngine) *team.TimeEngine {
				item := build.ToTimeEngine(ctx, dutyCycle)
				item.WithContext(ctx)
				return item
			})
		}
		item.WithContext(ctx)
		return item, true
	})
	hookAssociation := noticeGroupMutation.NoticeGroup.Hooks.WithContext(ctx).Model(groupDo)
	if len(group.GetHooks()) == 0 {
		if err := hookAssociation.Clear(); err != nil {
			return err
		}
	} else {
		if err := hookAssociation.Replace(hooks...); err != nil {
			return err
		}
	}
	if len(group.GetNoticeMembers()) == 0 {
		if err := noticeGroupMutation.NoticeGroup.Members.WithContext(ctx).Model(groupDo).Clear(); err != nil {
			return err
		}
	} else {
		if err := noticeGroupMutation.NoticeGroup.Members.WithContext(ctx).Model(groupDo).Replace(members...); err != nil {
			return err
		}
	}
	return nil
}

func (t *teamNoticeRepoImpl) UpdateStatus(ctx context.Context, req *bo.UpdateTeamNoticeGroupStatusRequest) error {
	groupIds := slices.MapFilter(req.GroupIds, func(groupId uint32) (uint32, bool) {
		if groupId <= 0 {
			return 0, false
		}
		return groupId, true
	})
	if len(groupIds) == 0 {
		return nil
	}

	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		bizMutation.NoticeGroup.TeamID.Eq(teamId),
		bizMutation.NoticeGroup.ID.In(groupIds...),
	}
	_, err := bizMutation.NoticeGroup.WithContext(ctx).Where(wrapper...).
		UpdateColumnSimple(bizMutation.NoticeGroup.Status.Value(req.Status.GetValue()))
	return err
}

func (t *teamNoticeRepoImpl) Delete(ctx context.Context, groupID uint32) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		bizMutation.NoticeGroup.TeamID.Eq(teamId),
		bizMutation.NoticeGroup.ID.Eq(groupID),
	}
	_, err := bizMutation.NoticeGroup.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (t *teamNoticeRepoImpl) Get(ctx context.Context, groupID uint32) (do.NoticeGroup, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	wrapper := []gen.Condition{
		bizQuery.NoticeGroup.TeamID.Eq(teamId),
		bizQuery.NoticeGroup.ID.Eq(groupID),
	}
	noticeGroup, err := bizQuery.NoticeGroup.WithContext(ctx).Where(wrapper...).Preload(field.Associations).First()
	if err != nil {
		return nil, noticeGroupNotFound(err)
	}
	return noticeGroup, nil
}

func (t *teamNoticeRepoImpl) FindByIds(ctx context.Context, groupIds []uint32) ([]do.NoticeGroup, error) {
	if len(groupIds) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.NoticeGroup
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId), mutation.ID.In(groupIds...))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.NoticeGroup) do.NoticeGroup { return row }), nil
}

func (t *teamNoticeRepoImpl) FindLabelNotices(ctx context.Context, labelNoticeIds []uint32) ([]do.StrategyMetricRuleLabelNotice, error) {
	if len(labelNoticeIds) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.StrategyMetricRuleLabelNotice
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId), mutation.ID.In(labelNoticeIds...))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.StrategyMetricRuleLabelNotice) do.StrategyMetricRuleLabelNotice { return row }), nil
}

func (t *teamNoticeRepoImpl) Select(ctx context.Context, req *bo.TeamNoticeGroupSelectRequest) (*bo.TeamNoticeGroupSelectReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, t)
	noticeGroupQuery := query.NoticeGroup
	wrapper := noticeGroupQuery.WithContext(ctx).Where(noticeGroupQuery.TeamID.Eq(teamID))
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(noticeGroupQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(noticeGroupQuery.Name.Like(req.Keyword))
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
		noticeGroupQuery.ID,
		noticeGroupQuery.Name,
		noticeGroupQuery.Remark,
		noticeGroupQuery.Status,
		noticeGroupQuery.DeletedAt,
	}
	rows, err := wrapper.Select(selectColumns...).Order(noticeGroupQuery.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return req.ToSelectReply(slices.Map(rows, func(row *team.NoticeGroup) do.NoticeGroup { return row })), nil
}
