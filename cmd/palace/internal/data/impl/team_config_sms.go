package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamConfigSMSRepo(data *data.Data) repository.TeamSMSConfig {
	return &teamConfigSMSRepoImpl{
		Data: data,
	}
}

type teamConfigSMSRepoImpl struct {
	*data.Data
}

func (t *teamConfigSMSRepoImpl) List(ctx context.Context, req *bo.ListSMSConfigRequest) (*bo.ListSMSConfigListReply, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizSMSConfigQuery := bizQuery.SmsConfig
	wrapper := bizSMSConfigQuery.WithContext(ctx).Where(bizSMSConfigQuery.TeamID.Eq(teamID))

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizSMSConfigQuery.Status.Eq(req.Status.GetValue()))
	}
	if !req.Provider.IsUnknown() {
		wrapper = wrapper.Where(bizSMSConfigQuery.Provider.Eq(req.Provider.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(bizSMSConfigQuery.Name.Like(req.Keyword))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}

	smsConfigs, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(smsConfigs, func(smsConfig *team.SmsConfig) do.TeamSMSConfig { return smsConfig })
	return req.ToListReply(rows), nil
}

func (t *teamConfigSMSRepoImpl) Create(ctx context.Context, config bo.TeamSMSConfig) error {
	smsConfigDo := &team.SmsConfig{
		TeamModel: do.TeamModel{},
		Name:      config.GetName(),
		Remark:    config.GetRemark(),
		Status:    config.GetStatus(),
		Provider:  config.GetProviderType(),
		Sms:       crypto.NewObject(config.GetSMSConfig()),
	}
	smsConfigDo.WithContext(ctx)
	bizQuery := getTeamBizQuery(ctx, t)
	bizSMSConfigQuery := bizQuery.SmsConfig
	return bizSMSConfigQuery.WithContext(ctx).Create(smsConfigDo)
}

func (t *teamConfigSMSRepoImpl) Update(ctx context.Context, config bo.TeamSMSConfig) error {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizSMSConfigQuery := bizQuery.SmsConfig
	wrappers := []gen.Condition{
		bizSMSConfigQuery.TeamID.Eq(teamID),
		bizSMSConfigQuery.ID.Eq(config.GetID()),
	}
	mutations := []field.AssignExpr{
		bizSMSConfigQuery.Name.Value(config.GetName()),
		bizSMSConfigQuery.Remark.Value(config.GetRemark()),
		bizSMSConfigQuery.Status.Value(config.GetStatus().GetValue()),
		bizSMSConfigQuery.Provider.Value(config.GetProviderType().GetValue()),
		bizSMSConfigQuery.Sms.Value(crypto.NewObject(config.GetSMSConfig())),
	}
	_, err := bizSMSConfigQuery.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

func (t *teamConfigSMSRepoImpl) Get(ctx context.Context, id uint32) (do.TeamSMSConfig, error) {
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizSMSConfigQuery := bizQuery.SmsConfig
	wrapper := []gen.Condition{
		bizSMSConfigQuery.TeamID.Eq(teamID),
		bizSMSConfigQuery.ID.Eq(id),
	}
	smsConfig, err := bizSMSConfigQuery.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, teamSMSConfigNotFound(err)
	}
	return smsConfig, nil
}

func (t *teamConfigSMSRepoImpl) FindByIds(ctx context.Context, ids []uint32) ([]do.TeamSMSConfig, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	bizQuery, teamID := getTeamBizQueryWithTeamID(ctx, t)
	bizSMSConfigQuery := bizQuery.SmsConfig
	wrapper := bizSMSConfigQuery.WithContext(ctx).Where(bizSMSConfigQuery.TeamID.Eq(teamID), bizSMSConfigQuery.ID.In(ids...))
	smsConfigs, err := wrapper.Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(smsConfigs, func(smsConfig *team.SmsConfig) do.TeamSMSConfig { return smsConfig }), nil
}
