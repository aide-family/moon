package impl

import (
	"context"

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

func NewTeamConfigSMSRepo(data *data.Data) repository.TeamSMSConfig {
	return &teamConfigSMSImpl{
		Data: data,
	}
}

type teamConfigSMSImpl struct {
	*data.Data
}

func (t *teamConfigSMSImpl) List(ctx context.Context, req *bo.ListSMSConfigRequest) (*bo.ListSMSConfigListReply, error) {
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
	return req.ToListSMSConfigListReply(smsConfigs), nil
}

func (t *teamConfigSMSImpl) Create(ctx context.Context, config bo.TeamSMSConfig) error {
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

func (t *teamConfigSMSImpl) Update(ctx context.Context, config bo.TeamSMSConfig) error {
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

func (t *teamConfigSMSImpl) Get(ctx context.Context, id uint32) (do.TeamSMSConfig, error) {
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
