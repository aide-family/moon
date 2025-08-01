package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewMemberRepo(data *data.Data, logger log.Logger) repository.Member {
	return &memberRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.member")),
	}
}

type memberRepoImpl struct {
	*data.Data
	helper *log.Helper
}

// Create implements repository.Member.
func (m *memberRepoImpl) Create(ctx context.Context, req *bo.CreateTeamMemberReq) error {
	memberQuery := getMainQuery(ctx, m).TeamMember
	memberDo := &system.TeamMember{
		TeamModel:  do.TeamModel{TeamID: req.Team.GetID()},
		MemberName: req.User.GetUsername(),
		Remark:     req.User.GetRemark(),
		UserID:     req.User.GetID(),
		Position:   req.Position,
		Status:     req.Status,
	}
	memberDo.WithContext(ctx)
	if err := memberQuery.WithContext(ctx).Create(memberDo); err != nil {
		return err
	}
	return nil
}

func (m *memberRepoImpl) List(ctx context.Context, req *bo.TeamMemberListRequest) (*bo.TeamMemberListReply, error) {
	if validate.IsNil(req) {
		return nil, merr.ErrorParams("invalid request")
	}

	memberQuery := getMainQuery(ctx, m).TeamMember
	wrapper := memberQuery.WithContext(ctx)
	if req.TeamID > 0 {
		wrapper = wrapper.Where(memberQuery.TeamID.Eq(req.TeamID))
	}
	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			memberQuery.MemberName.Like(req.Keyword),
			memberQuery.Remark.Like(req.Keyword),
		}
		wrapper = wrapper.Where(memberQuery.Or(ors...))
	}
	if len(req.Status) > 0 {
		status := slices.Map(req.Status, func(statusItem vobj.MemberStatus) int8 { return statusItem.GetValue() })
		wrapper = wrapper.Where(memberQuery.Status.In(status...))
	}
	if len(req.Positions) > 0 {
		positions := slices.Map(req.Positions, func(positionItem vobj.Position) int8 { return positionItem.GetValue() })
		wrapper = wrapper.Where(memberQuery.Position.In(positions...))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	members, err := wrapper.Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(members, func(member *system.TeamMember) do.TeamMember { return member })
	return req.ToListReply(rows), nil
}

func (m *memberRepoImpl) Select(ctx context.Context, req *bo.SelectTeamMembersRequest) (*bo.SelectTeamMembersReply, error) {
	if validate.IsNil(req) {
		return nil, merr.ErrorParams("invalid request")
	}
	mainQuery := getMainQuery(ctx, m)
	memberQuery := mainQuery.TeamMember
	userQuery := mainQuery.User
	wrapper := memberQuery.WithContext(ctx).Where(memberQuery.TeamID.Eq(req.TeamID))
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(memberQuery.MemberName.Like(req.Keyword))
	}
	if len(req.Status) > 0 {
		status := slices.Map(req.Status, func(statusItem vobj.MemberStatus) int8 { return statusItem.GetValue() })
		wrapper = wrapper.Where(memberQuery.Status.In(status...))
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
		memberQuery.ID,
		memberQuery.MemberName,
		memberQuery.Remark,
		memberQuery.Status,
		memberQuery.DeletedAt,
	}
	userSelectColumns := []field.Expr{
		userQuery.ID,
		userQuery.Username,
		userQuery.Nickname,
		userQuery.Avatar,
		userQuery.Gender,
		userQuery.DeletedAt,
		userQuery.Status,
	}
	members, err := wrapper.Select(selectColumns...).Order(memberQuery.ID.Desc()).Preload(memberQuery.User.Select(userSelectColumns...)).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(members, func(member *system.TeamMember) do.TeamMember { return member })
	return req.ToSelectReply(rows), nil
}

func (m *memberRepoImpl) UpdateStatus(ctx context.Context, req bo.UpdateMemberStatus) error {
	if validate.IsNil(req) {
		return merr.ErrorParams("invalid request")
	}
	if len(req.GetMembers()) == 0 {
		return nil
	}
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorPermissionDenied("team id not found")
	}
	memberIds := slices.MapFilter(req.GetMembers(), func(member do.TeamMember) (uint32, bool) {
		if validate.IsNil(member) || member.GetID() <= 0 {
			return 0, false
		}
		return member.GetID(), true
	})
	if len(memberIds) == 0 {
		return nil
	}
	memberQuery := getMainQuery(ctx, m).TeamMember
	wrappers := []gen.Condition{
		memberQuery.TeamID.Eq(teamID),
		memberQuery.ID.In(memberIds...),
	}
	_, err := memberQuery.WithContext(ctx).Where(wrappers...).UpdateSimple(memberQuery.Status.Value(req.GetStatus().GetValue()))
	return err
}

func (m *memberRepoImpl) UpdatePosition(ctx context.Context, req bo.UpdateMemberPosition) error {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorPermissionDenied("team id not found")
	}

	memberQuery := getMainQuery(ctx, m).TeamMember
	wrappers := []gen.Condition{
		memberQuery.TeamID.Eq(teamID),
		memberQuery.ID.Eq(req.GetMember().GetID()),
	}
	_, err := memberQuery.WithContext(ctx).Where(wrappers...).UpdateSimple(memberQuery.Position.Value(req.GetPosition().GetValue()))
	return err
}

func (m *memberRepoImpl) UpdateRoles(ctx context.Context, req bo.UpdateMemberRoles) error {
	memberDo := &system.TeamMember{
		TeamModel: do.TeamModel{
			CreatorModel: do.CreatorModel{
				BaseModel: do.BaseModel{ID: req.GetMember().GetID()},
			},
		},
	}

	roles := slices.MapFilter(req.GetRoles(), func(role do.TeamRole) (*system.TeamRole, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return nil, false
		}
		return &system.TeamRole{
			TeamModel: do.TeamModel{
				CreatorModel: do.CreatorModel{
					BaseModel: do.BaseModel{ID: role.GetID()},
				},
			},
		}, true
	})

	memberMutation := getMainQuery(ctx, m).TeamMember
	rolesAssociation := memberMutation.Roles.WithContext(ctx).Model(memberDo)
	if len(roles) == 0 {
		return rolesAssociation.Clear()
	}
	return rolesAssociation.Replace(roles...)
}

func (m *memberRepoImpl) Get(ctx context.Context, id uint32) (do.TeamMember, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	memberQuery := getMainQuery(ctx, m).TeamMember
	wrappers := []gen.Condition{
		memberQuery.TeamID.Eq(teamID),
		memberQuery.ID.Eq(id),
	}
	member, err := memberQuery.WithContext(ctx).Where(wrappers...).Preload(field.Associations).First()
	if err != nil {
		return nil, teamMemberNotFound(err)
	}
	return member, nil
}

func (m *memberRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.TeamMember, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	query := getMainQuery(ctx, m)
	memberQuery := query.TeamMember
	wrappers := []gen.Condition{
		memberQuery.TeamID.Eq(teamID),
		memberQuery.ID.In(ids...),
	}
	members, err := memberQuery.WithContext(ctx).Where(wrappers...).Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	memberDos := slices.Map(members, func(member *system.TeamMember) do.TeamMember { return member })
	return memberDos, nil
}

func (m *memberRepoImpl) FindByUserID(ctx context.Context, userID uint32) (do.TeamMember, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	query := getMainQuery(ctx, m)
	memberQuery := query.TeamMember
	wrappers := []gen.Condition{
		memberQuery.TeamID.Eq(teamID),
		memberQuery.UserID.Eq(userID),
	}
	member, err := memberQuery.WithContext(ctx).Where(wrappers...).Preload(field.Associations).First()
	if err != nil {
		return nil, teamMemberNotFound(err)
	}
	return member, nil
}
