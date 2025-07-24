package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamRepo(d *data.Data, logger log.Logger) repository.Team {
	return &teamRepoImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team")),
	}
}

type teamRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func (r *teamRepoImpl) Create(ctx context.Context, team bo.CreateTeamRequest) error {
	teamMutation := getMainQuery(ctx, r).Team
	teamDo := &system.Team{
		Name:          team.GetName(),
		Status:        team.GetStatus(),
		Remark:        team.GetRemark(),
		Logo:          team.GetLogo(),
		LeaderID:      team.GetLeader().GetID(),
		UUID:          team.GetUUID(),
		Capacity:      team.GetCapacity(),
		Leader:        nil,
		Admins:        nil,
		Menus:         nil,
		BizDBConfig:   crypto.NewObject(team.GetBizDBConfig()),
		AlarmDBConfig: crypto.NewObject(team.GetAlarmDBConfig()),
	}
	teamDo.WithContext(ctx)
	return teamMutation.WithContext(ctx).Create(teamDo)
}

func (r *teamRepoImpl) Update(ctx context.Context, team bo.UpdateTeamRequest) error {
	teamMutation := getMainQuery(ctx, r).Team
	wrappers := []gen.Condition{
		teamMutation.ID.Eq(team.GetTeam().GetID()),
	}
	mutations := []field.AssignExpr{
		teamMutation.Name.Value(team.GetName()),
		teamMutation.Remark.Value(team.GetRemark()),
		teamMutation.Logo.Value(team.GetLogo()),
	}
	_, err := teamMutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

func (r *teamRepoImpl) Delete(ctx context.Context, id uint32) error {
	teamMutation := getMainQuery(ctx, r).Team
	wrappers := []gen.Condition{
		teamMutation.ID.Eq(id),
	}
	_, err := teamMutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

func (r *teamRepoImpl) FindByID(ctx context.Context, id uint32) (do.Team, error) {
	systemQuery := getMainQuery(ctx, r).Team
	teamDo, err := systemQuery.WithContext(ctx).Preload(field.Associations).Where(systemQuery.ID.Eq(id)).First()
	if err != nil {
		return nil, teamNotFound(err)
	}
	return teamDo, nil
}

func (r *teamRepoImpl) List(ctx context.Context, req *bo.TeamListRequest) (*bo.TeamListReply, error) {
	query := getMainQuery(ctx, r)
	teamQuery := query.Team
	wrapper := teamQuery.WithContext(ctx).Preload(field.Associations)
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(teamQuery.Name.Like(req.Keyword))
	}
	if len(req.Status) > 0 {
		status := slices.Map(req.Status, func(statusItem vobj.TeamStatus) int8 { return statusItem.GetValue() })
		wrapper = wrapper.Where(teamQuery.Status.In(status...))
	}
	if req.LeaderID > 0 {
		wrapper = wrapper.Where(teamQuery.LeaderID.Eq(req.LeaderID))
	}
	if req.CreatorID > 0 {
		wrapper = wrapper.Where(teamQuery.CreatorID.Eq(req.CreatorID))
	}
	if len(req.UserIds) > 0 {
		userQuery := query.User
		users, err := userQuery.WithContext(ctx).Where(userQuery.ID.In(req.UserIds...)).Preload(userQuery.Teams).Find()
		if err != nil {
			return nil, err
		}
		if len(users) > 0 {
			var teamIds []uint32
			for _, user := range users {
				teamIds = append(teamIds, slices.Map(user.GetTeams(), func(team do.Team) uint32 { return team.GetID() })...)
			}
			if len(teamIds) > 0 {
				wrapper = wrapper.Where(teamQuery.ID.In(teamIds...))
			}
		}
		wrapper = wrapper.Where(teamQuery.LeaderID.In(req.UserIds...))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}

	teamDos, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(teamDos, func(teamDo *system.Team) do.Team { return teamDo })
	return req.ToListReply(rows), nil
}

func (r *teamRepoImpl) CheckNameUnique(ctx context.Context, name string, teamID uint32) error {
	teamQuery := getMainQuery(ctx, r).Team
	wrapper := teamQuery.WithContext(ctx)
	if teamID > 0 {
		wrapper = wrapper.Where(teamQuery.ID.Neq(teamID))
	}
	teamDo, err := wrapper.Where(teamQuery.Name.Eq(name)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if teamDo != nil {
		return merr.ErrorConflict("team name already exists")
	}
	return nil
}

func (r *teamRepoImpl) FindByName(ctx context.Context, name string) (do.Team, error) {
	teamQuery := getMainQuery(ctx, r).Team
	wrapper := teamQuery.WithContext(ctx).Preload(field.Associations)
	teamDo, err := wrapper.Where(teamQuery.Name.Eq(name)).First()
	if err != nil {
		return nil, teamNotFound(err)
	}
	if teamDo == nil {
		return nil, merr.ErrorNotFound("team not found")
	}
	return teamDo, nil
}
