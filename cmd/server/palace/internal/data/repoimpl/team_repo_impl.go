package repoimpl

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel/bizquery"
	"github.com/aide-cloud/moon/pkg/helper/model/query"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/utils/random"
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
		Status:    team.Status.GetValue(),
		Remark:    team.Remark,
		Logo:      team.Logo,
		LeaderID:  team.LeaderID,
		CreatorID: team.CreatorID,
		UUID:      random.UUIDToUpperCase(true),
	}
	// 判断团队名称是否重复
	_, err := query.Use(l.data.GetMainDB(ctx)).SysTeam.WithContext(ctx).Where(query.SysTeam.Name.Eq(team.Name)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, bo.TeamNameExistErr
	}

	sysApis, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysAPI.Find()
	if err != nil {
		return nil, err
	}
	teamApis := types.SliceToWithFilter(sysApis, func(apiItem *model.SysAPI) (*bizmodel.SysTeamAPI, bool) {
		return &bizmodel.SysTeamAPI{
			ID:        apiItem.ID,
			CreatedAt: apiItem.CreatedAt,
			UpdatedAt: apiItem.UpdatedAt,
			DeletedAt: apiItem.DeletedAt,
			Name:      apiItem.Name,
			Path:      apiItem.Path,
			Status:    apiItem.Status,
			Remark:    apiItem.Remark,
			Module:    apiItem.Module,
			Domain:    apiItem.Domain,
		}, true
	})

	err = query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建基础信息
		if err = tx.SysTeam.Create(sysTeamModel); err != nil {
			return err
		}
		teamId := sysTeamModel.ID
		flag := true
		// 添加管理员成员
		adminMembers := types.SliceToWithFilter(team.Admins, func(memberId uint32) (*bizmodel.SysTeamMember, bool) {
			if memberId <= 0 {
				return nil, false
			}
			if memberId == sysTeamModel.LeaderID {
				flag = false
			}
			return &bizmodel.SysTeamMember{
				UserID: memberId,
				TeamID: teamId,
				Status: vobj.StatusEnable.GetValue(),
				Role:   vobj.RoleAdmin.GetValue(),
			}, true
		})
		if flag {
			adminMembers = append(adminMembers, &bizmodel.SysTeamMember{
				UserID: sysTeamModel.LeaderID,
				TeamID: teamId,
				Status: vobj.StatusEnable.GetValue(),
				Role:   vobj.RoleAdmin.GetValue(),
			})
		}

		// 创建团队数据库
		_, err = l.data.GetBizDB(ctx).Exec("CREATE DATABASE IF NOT EXISTS " + "`" + data.GenBizDatabaseName(teamId) + "`")
		if err != nil {
			return err
		}
		bizDB, err := l.data.GetBizGormDB(teamId)
		if err != nil {
			return err
		}
		// TODO 初始化数据表
		if err = bizDB.AutoMigrate(bizmodel.Models()...); err != nil {
			return err
		}

		if err = bizquery.Use(bizDB).SysTeamMember.WithContext(ctx).Create(adminMembers...); err != nil {
			return err
		}

		// 迁移api数据到团队数据库
		return bizquery.Use(bizDB).SysTeamAPI.Create(teamApis...)
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
		q = q.Where(query.SysTeam.Status.Eq(params.Status.GetValue()))
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
		// TODO 需要缓存用户的全部团队ID， 然后取出来使用
		//queryTeamIds = true
		//bizDB, err := l.data.GetBizGormDB(teamId)
		//if err != nil {
		//	return nil, err
		//}
		//if err := bizquery.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
		//	bizquery.SysTeamMember.UserID.Eq(params.UserID),
		//).Pluck(bizquery.SysTeamMember.TeamID, &teamIds); err != nil {
		//	return nil, err
		//}
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
		UpdateColumnSimple(query.SysTeam.Status.Value(status.GetValue()))
	return err
}

func (l *teamRepoImpl) GetUserTeamList(ctx context.Context, userID uint32) ([]*model.SysTeam, error) {
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam
	// TODO 需要缓存用户的全部团队ID， 然后取出来使用
	//var teamIds []uint32
	//if err := bizquery.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeamMember.Where(
	//	bizquery.SysTeamMember.UserID.Eq(userID),
	//).Pluck(bizquery.SysTeamMember.TeamID, &teamIds); err != nil {
	//	return nil, err
	//}
	//q = q.Where(query.SysTeam.ID.In(teamIds...))

	return q.Find()
}

func (l *teamRepoImpl) AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error {
	members := types.SliceToWithFilter(params.Members, func(memberItem *bo.AddTeamMemberItem) (*bizmodel.SysTeamMember, bool) {
		if types.IsNil(memberItem) {
			return nil, false
		}
		return &bizmodel.SysTeamMember{
			UserID: memberItem.UserID,
			TeamID: params.ID,
			Status: vobj.StatusEnable.GetValue(),
			Role:   memberItem.Role.GetValue(),
			TeamRoles: types.SliceToWithFilter(memberItem.RoleIds, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
				if roleId <= 0 {
					return nil, false
				}
				return &bizmodel.SysTeamRole{
					ID: roleId,
				}, true
			}),
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return err
	}
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if err := tx.SysTeamMember.WithContext(ctx).Create(members...); err != nil {
			return err
		}
		return nil
	})
}

func (l *teamRepoImpl) RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return err
	}
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.SysTeamMember.WithContext(ctx).
			Where(tx.SysTeamMember.TeamID.Eq(params.ID), tx.SysTeamMember.UserID.In(params.MemberIds...)).
			Delete(); err != nil {
			return err
		}
		if _, err = tx.SysTeamMemberRole.WithContext(ctx).
			Where(tx.SysTeamMemberRole.SysTeamMemberID.In(params.MemberIds...)).
			Delete(); err != nil {
			return err
		}
		return nil
	})
}

func (l *teamRepoImpl) SetMemberAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return err
	}
	q := bizquery.Use(bizDB)
	_, err = q.WithContext(ctx).SysTeamMember.Where(
		q.SysTeamMember.TeamID.Eq(params.ID),
		q.SysTeamMember.UserID.In(params.MemberIds...),
	).UpdateColumnSimple(q.SysTeamMember.Role.Value(params.Role.GetValue()))
	return err
}

func (l *teamRepoImpl) SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error {
	roles := types.SliceToWithFilter(params.RoleIds, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
		if roleId == 0 {
			return nil, false
		}
		return &bizmodel.SysTeamRole{
			ID: roleId,
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return err
	}
	return bizquery.Use(bizDB).SysTeamMember.TeamRoles.
		Model(&bizmodel.SysTeamMember{ID: params.MemberID}).Replace(roles...)
}

func (l *teamRepoImpl) ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error) {
	var wheres []gen.Condition
	var userWheres []gen.Condition
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return nil, err
	}
	qq := bizquery.Use(bizDB)
	if !types.TextIsNull(params.Keyword) {
		userWheres = append(userWheres, query.SysUser.Username.Like(params.Keyword))
	}
	if !params.Gender.IsUnknown() {
		userWheres = append(userWheres, query.SysUser.Gender.Eq(params.Gender.GetValue()))
	}
	if len(userWheres) > 0 {
		var userIds []uint32
		if err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(userWheres...).
			Pluck(query.SysUser.ID, &userIds); err != nil {
			return nil, err
		}
		wheres = append(wheres, qq.SysTeamMember.UserID.In(userIds...))
	}

	if !params.Role.IsAll() {
		wheres = append(wheres, qq.SysTeamMember.Role.Eq(params.Role.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, qq.SysTeamMember.Status.Eq(params.Status.GetValue()))
	}
	if len(params.MemberIDs) > 0 {
		wheres = append(wheres, qq.SysTeamMember.UserID.In(params.MemberIDs...))
	}

	q := bizquery.Use(bizDB).WithContext(ctx).SysTeamMember.Where(wheres...)
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
	return q.Order(qq.SysTeamMember.Role.Desc(), qq.SysTeamMember.ID.Asc()).Find()
}

func (l *teamRepoImpl) TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error {
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if err != nil {
		return err
	}
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		// 设置新管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.TeamID.Eq(params.ID),
			tx.SysTeamMember.UserID.Eq(params.LeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleSuperAdmin.GetValue())); err != nil {
			return err
		}
		// 设置老管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.TeamID.Eq(params.ID),
			tx.SysTeamMember.UserID.Neq(params.OldLeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleAdmin.GetValue())); err != nil {
			return err
		}
		// 系统团队信息
		if _, err = query.Use(l.data.GetMainDB(ctx)).SysTeam.WithContext(ctx).Where(
			query.SysTeam.ID.Eq(params.ID),
		).UpdateColumnSimple(query.SysTeam.LeaderID.Value(params.LeaderID)); err != nil {
			return err
		}
		return nil
	})
}

func (l *teamRepoImpl) GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*bizmodel.SysTeamMember, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if err != nil {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	return q.WithContext(ctx).SysTeamMember.Where(
		q.SysTeamMember.TeamID.Eq(teamID),
		q.SysTeamMember.UserID.Eq(userID),
	).First()
}
