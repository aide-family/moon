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
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamConfigEmailRepo(data *data.Data, logger log.Logger) repository.TeamEmailConfig {
	return &teamConfigEmailRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team_config_email")),
	}
}

type teamConfigEmailRepoImpl struct {
	*data.Data
	helper *log.Helper
}

// CheckNameUnique implements repository.TeamEmailConfig.
func (t *teamConfigEmailRepoImpl) CheckNameUnique(ctx context.Context, name string, configID uint32) error {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	wrappers := []gen.Condition{
		bizEmailConfigQuery.TeamID.Eq(teamID),
		bizEmailConfigQuery.Name.Eq(name),
	}
	if configID != 0 {
		wrappers = append(wrappers, bizEmailConfigQuery.ID.Neq(configID))
	}
	emailConfig, err := bizEmailConfigQuery.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if emailConfig != nil {
		return merr.ErrorConflict("name already exists")
	}
	return nil
}

func (t *teamConfigEmailRepoImpl) Get(ctx context.Context, id uint32) (do.TeamEmailConfig, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	wrappers := []gen.Condition{
		bizEmailConfigQuery.TeamID.Eq(teamID),
		bizEmailConfigQuery.ID.Eq(id),
	}
	emailConfig, err := bizEmailConfigQuery.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		return nil, teamEmailConfigNotFound(err)
	}
	return emailConfig, nil
}

func (t *teamConfigEmailRepoImpl) List(ctx context.Context, req *bo.ListEmailConfigRequest) (*bo.ListEmailConfigListReply, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	wrapper := bizEmailConfigQuery.WithContext(ctx).Where(bizEmailConfigQuery.TeamID.Eq(teamID))
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizEmailConfigQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(bizEmailConfigQuery.Name.Like(req.Keyword))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	emailConfigs, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(emailConfigs, func(emailConfig *team.EmailConfig) do.TeamEmailConfig { return emailConfig })
	return req.ToListReply(rows), nil
}

func (t *teamConfigEmailRepoImpl) Create(ctx context.Context, config bo.TeamEmailConfig) (uint32, error) {
	emailConfigDo := &team.EmailConfig{
		TeamModel: do.TeamModel{},
		Name:      config.GetName(),
		Remark:    config.GetRemark(),
		Status:    config.GetStatus(),
		Email:     crypto.NewObject(config.GetEmailConfig()),
	}
	emailConfigDo.WithContext(ctx)
	bizQuery := getTeamBizQuery(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	if err := bizEmailConfigQuery.WithContext(ctx).Create(emailConfigDo); err != nil {
		return 0, err
	}
	return emailConfigDo.ID, nil
}

func (t *teamConfigEmailRepoImpl) Update(ctx context.Context, config bo.TeamEmailConfig) error {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	wrappers := []gen.Condition{
		bizEmailConfigQuery.TeamID.Eq(teamID),
		bizEmailConfigQuery.ID.Eq(config.GetID()),
	}
	mutations := []field.AssignExpr{
		bizEmailConfigQuery.Name.Value(config.GetName()),
		bizEmailConfigQuery.Remark.Value(config.GetRemark()),
		bizEmailConfigQuery.Status.Value(config.GetStatus().GetValue()),
		bizEmailConfigQuery.Email.Value(crypto.NewObject(config.GetEmailConfig())),
	}
	_, err := bizEmailConfigQuery.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

func (t *teamConfigEmailRepoImpl) FindByIds(ctx context.Context, ids []uint32) ([]do.TeamEmailConfig, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizEmailConfigQuery := bizQuery.EmailConfig
	wrapper := bizEmailConfigQuery.WithContext(ctx).Where(bizEmailConfigQuery.TeamID.Eq(teamID), bizEmailConfigQuery.ID.In(ids...))
	emailConfigs, err := wrapper.Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(emailConfigs, func(emailConfig *team.EmailConfig) do.TeamEmailConfig { return emailConfig }), nil
}
