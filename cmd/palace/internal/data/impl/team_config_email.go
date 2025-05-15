package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/pkg/util/crypto"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewTeamConfigEmailRepo(data *data.Data, logger log.Logger) repository.TeamEmailConfig {
	return &teamConfigEmailImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team_config_email")),
	}
}

type teamConfigEmailImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamConfigEmailImpl) Get(ctx context.Context, id uint32) (do.TeamEmailConfig, error) {
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

func (t *teamConfigEmailImpl) List(ctx context.Context, req *bo.ListEmailConfigRequest) (*bo.ListEmailConfigListReply, error) {
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
	return req.ToListEmailConfigListReply(emailConfigs), nil
}

func (t *teamConfigEmailImpl) Create(ctx context.Context, config bo.TeamEmailConfig) error {
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
	return bizEmailConfigQuery.WithContext(ctx).Create(emailConfigDo)
}

func (t *teamConfigEmailImpl) Update(ctx context.Context, config bo.TeamEmailConfig) error {
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
