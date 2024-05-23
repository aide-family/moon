package repoimpl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/query"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewTeamRepo(data *data.Data) repo.TeamRepo {
	return &teamRepoImpl{
		data: data,
	}
}

type teamRepoImpl struct {
	data *data.Data
}

func (l *teamRepoImpl) CreateTeam(ctx context.Context, team *bo.CreateTeamParams) (*model.SysTeam, error) {
	sysTeamModel := &model.SysTeam{
		Name:      team.Name,
		Status:    team.Status,
		Remark:    team.Remark,
		Logo:      team.Logo,
		LeaderID:  team.LeaderID,
		CreatorID: team.CreatorID,
	}

	err := query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建基础信息
		if err := tx.SysTeam.Create(sysTeamModel); err != nil {
			return err
		}
		teamId := sysTeamModel.ID
		// 添加管理员成员
		adminMembers := types.SliceToWithFilter(team.Admins, func(memberId uint32) (*model.SysTeamMember, bool) {
			return &model.SysTeamMember{
				UserID: memberId,
				TeamID: teamId,
				Status: vobj.StatusEnable,
				Role:   vobj.RoleAdmin,
			}, true
		})
		if err := tx.SysTeamMember.Create(adminMembers...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sysTeamModel, nil
}

func (l *teamRepoImpl) UpdateTeam(ctx context.Context, team *bo.UpdateTeamParams) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.
		Where(query.SysTeam.ID.Eq(team.ID)).
		UpdateColumnSimple(
			query.SysTeam.Name.Value(team.Name),
			query.SysTeam.Remark.Value(team.Remark),
		)
	return err
}

func (l *teamRepoImpl) GetTeamDetail(ctx context.Context, teamID uint32) (*model.SysTeam, error) {
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.Where(query.SysTeam.ID.Eq(teamID)).First()
}

func (l *teamRepoImpl) GetTeamList(ctx context.Context, params *bo.QueryTeamListParams) ([]*model.SysTeam, error) {
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam
	if !types.TextIsNull(params.Keyword) {
		q = q.Where(query.SysTeam.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		q = q.Where(query.SysTeam.Status.Eq(params.Status))
	}
	if params.CreatorID > 0 {
		q = q.Where(query.SysTeam.CreatorID.Eq(params.CreatorID))
	}
	if params.LeaderID > 0 {
		q = q.Where(query.SysTeam.LeaderID.Eq(params.LeaderID))
	}
	var teamIds []uint32
	queryTeamIds := false
	if params.UserID > 0 {
		queryTeamIds = true
		if err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
			query.SysTeamMember.UserID.Eq(params.UserID),
		).Pluck(query.SysTeamMember.TeamID, &teamIds); err != nil {
			return nil, err
		}
	}
	if len(params.IDs) > 0 {
		queryTeamIds = true
		teamIds = append(teamIds, params.IDs...)
	}
	if queryTeamIds {
		q = q.Where(query.SysTeam.ID.In(teamIds...))
	}

	if !types.IsNil(params.Page) {
		total, err := q.Count()
		if err != nil {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		page := params.Page
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Order(query.SysTeam.ID.Desc()).Find()
}

func (l *teamRepoImpl) UpdateTeamStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.Where(query.SysTeam.ID.In(ids...)).
		UpdateColumnSimple(query.SysTeam.Status.Value(status))
	return err
}

func (l *teamRepoImpl) GetUserTeamList(ctx context.Context, userID uint32) ([]*model.SysTeam, error) {
	var teamIds []uint32
	if err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
		query.SysTeamMember.UserID.Eq(userID),
	).Pluck(query.SysTeamMember.TeamID, &teamIds); err != nil {
		return nil, err
	}
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.Where(query.SysTeam.ID.In(teamIds...)).Find()
}

func (l *teamRepoImpl) AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error {
	roles := make([]*model.SysTeamMemberRole, 0, len(params.Members))
	members := types.SliceToWithFilter(params.Members, func(memberItem *bo.AddTeamMemberItem) (*model.SysTeamMember, bool) {
		if types.IsNil(memberItem) {
			return nil, false
		}
		userRoles := types.SliceToWithFilter(memberItem.RoleIds, func(roleId uint32) (*model.SysTeamMemberRole, bool) {
			if roleId == 0 {
				return nil, false
			}
			return &model.SysTeamMemberRole{
				UserID: memberItem.UserID,
				TeamID: params.ID,
				RoleID: roleId,
			}, true
		})
		roles = append(roles, userRoles...)
		return &model.SysTeamMember{
			UserID: memberItem.UserID,
			TeamID: params.ID,
			Status: vobj.StatusEnable,
			Role:   memberItem.Role,
		}, true
	})
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if err := tx.SysTeamMember.WithContext(ctx).Create(members...); err != nil {
			return err
		}
		if len(roles) > 0 {
			return tx.SysTeamMemberRole.WithContext(ctx).Create(roles...)
		}
		return nil
	})
}

func (l *teamRepoImpl) RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if _, err := tx.SysTeamMember.WithContext(ctx).
			Where(query.SysTeamMember.TeamID.Eq(params.ID), query.SysTeamMember.UserID.In(params.MemberIds...)).
			Delete(); err != nil {
			return err
		}
		if _, err := tx.SysTeamMemberRole.WithContext(ctx).
			Where(query.SysTeamMemberRole.TeamID.Eq(params.ID), query.SysTeamMemberRole.UserID.In(params.MemberIds...)).
			Delete(); err != nil {
			return err
		}
		return nil
	})
}

func (l *teamRepoImpl) SetMemberAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
		query.SysTeamMember.TeamID.Eq(params.ID),
		query.SysTeamMember.UserID.In(params.MemberIds...),
	).UpdateColumnSimple(query.SysTeamMember.Role.Value(params.Role))
	return err
}

func (l *teamRepoImpl) SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error {
	roles := types.SliceToWithFilter(params.RoleIds, func(roleId uint32) (*model.SysTeamMemberRole, bool) {
		if roleId == 0 {
			return nil, false
		}
		return &model.SysTeamMemberRole{
			UserID: params.MemberID,
			TeamID: params.ID,
			RoleID: roleId,
		}, true
	})
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 删除之前的全部角色信息
		if _, err := tx.SysTeamMemberRole.WithContext(ctx).
			Where(query.SysTeamMemberRole.TeamID.Eq(params.ID), query.SysTeamMemberRole.UserID.Eq(params.MemberID)).
			Delete(); err != nil {
			return err
		}
		if len(roles) > 0 {
			// 创建新的角色信息
			return tx.SysTeamMemberRole.WithContext(ctx).Create(roles...)
		}
		return nil
	})
}

func (l *teamRepoImpl) ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*model.SysTeamMember, error) {
	var ons []field.Expr
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		ons = append(ons, query.SysUser.Username.Like(params.Keyword))
	}
	if !params.Gender.IsUnknown() {
		ons = append(ons, query.SysUser.Gender.Eq(params.Gender))
	}
	if !params.Role.IsAll() {
		wheres = append(wheres, query.SysTeamMember.Role.Eq(params.Role))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, query.SysTeamMember.Status.Eq(params.Status))
	}
	if len(params.MemberIDs) > 0 {
		wheres = append(wheres, query.SysTeamMember.UserID.In(params.MemberIDs...))
	}
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(wheres...).LeftJoin(query.SysUser, ons...)
	if !types.IsNil(params.Page) {
		total, err := q.Count()
		if err != nil {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		page := params.Page
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Order(query.SysTeamMember.Role.Desc(), query.SysTeamMember.ID.Asc()).Find()
}

func (l *teamRepoImpl) TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 设置新管理员
		if _, err := tx.SysTeamMember.WithContext(ctx).Where(
			query.SysTeamMember.TeamID.Eq(params.ID),
			query.SysTeamMember.UserID.Eq(params.LeaderID),
		).UpdateColumnSimple(query.SysTeamMember.Role.Value(vobj.RoleSuperAdmin)); err != nil {
			return err
		}
		// 设置老管理员
		if _, err := tx.SysTeamMember.WithContext(ctx).Where(
			query.SysTeamMember.TeamID.Eq(params.ID),
			query.SysTeamMember.UserID.Neq(params.OldLeaderID),
		).UpdateColumnSimple(query.SysTeamMember.Role.Value(vobj.RoleAdmin)); err != nil {
			return err
		}
		// 系统团队信息
		if _, err := tx.SysTeam.WithContext(ctx).Where(
			query.SysTeam.ID.Eq(params.ID),
		).UpdateColumnSimple(query.SysTeam.LeaderID.Value(params.LeaderID)); err != nil {
			return err
		}
		return nil
	})
}

func (l *teamRepoImpl) GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*model.SysTeamMember, error) {
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
		query.SysTeamMember.TeamID.Eq(teamID),
		query.SysTeamMember.UserID.Eq(userID),
	).First()
}

func (l *teamRepoImpl) GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*model.SysTeamMemberRole, error) {
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMemberRole.Where(
		query.SysTeamMemberRole.TeamID.Eq(teamID),
		query.SysTeamMemberRole.UserID.Eq(userID),
	).Find()
}

func (l *teamRepoImpl) GetUserTeamRole(ctx context.Context, userID, teamID uint32) (*model.SysTeamMemberRole, error) {
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMemberRole.Where(
		query.SysTeamMemberRole.TeamID.Eq(teamID),
		query.SysTeamMemberRole.UserID.Eq(userID),
	).First()
}
