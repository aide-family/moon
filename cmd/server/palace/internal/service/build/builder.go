package build

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	teamapi "github.com/aide-family/moon/api/admin/team"
	userapi "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

func NewBuilder() *builder {
	return &builder{}
}

type (
	builder struct {
		ctx context.Context
	}

	Builder interface {
		WithContext(ctx context.Context) Builder

		// TODO 注册新的数据转换方法写在这里

		WithDoDatasource(d *bizmodel.Datasource) DatasourceModelBuilder
		WithCreateDatasourceBo(user *datasourceapi.CreateDatasourceRequest) DatasourceRequestBuilder
		WithListDatasourceBo(user *datasourceapi.ListDatasourceRequest) DatasourceRequestBuilder

		WithBoDatasourceQueryData(d *bo.DatasourceQueryData) DatasourceQueryDataBuilder

		WithApiTemplateStrategy(template *model.StrategyTemplate) TemplateModelBuilder
		WithCreateBoTemplateStrategy(template *strategyapi.CreateTemplateStrategyRequest) TemplateRequestBuilder
		WithUpdateBoTemplateStrategy(template *strategyapi.UpdateTemplateStrategyRequest) TemplateRequestBuilder

		WithApiTemplateStrategyLevel(*model.StrategyLevelTemplate) TemplateLevelBuilder

		WithApiStrategy(strategy *bizmodel.Strategy) StrategyModelBuilder

		WithCreateBoStrategy(strategy *strategyapi.CreateStrategyRequest) StrategyRequestBuilder

		WithUpdateBoStrategy(strategy *strategyapi.UpdateStrategyRequest) StrategyRequestBuilder

		WithApiStrategyLevel(strategy *bizmodel.StrategyLevel) StrategyLevelModelBuilder

		WithCreateBoDict(dict *dictapi.CreateDictRequest) DictRequestBuilder

		WithUpdateBoDict(dict *dictapi.UpdateDictRequest) DictRequestBuilder

		WithApiDict(dict *model.SysDict) DictModelBuilder

		WithApiDictSelect(dict *model.SysDict) DictModelBuilder

		WithCreateMenuBo(menu *menuapi.CreateMenuRequest) MenuRequestBuilder

		WithUpdateMenuBo(menu *menuapi.UpdateMenuRequest) MenuRequestBuilder

		WithApiMenu(menu *model.SysMenu) MenuModelBuilder

		WithBatchCreateMenuBo(menus *menuapi.BatchCreateMenuRequest) MenuRequestBuilder

		WithApiMenuTree(menuList []*admin.Menu, parentID uint32) MenuTreeBuilder

		WithApiTeam(team *model.SysTeam) TeamModelBuilder

		WithSelectTeamRole(team *bizmodel.SysTeamRole) TeamRoleBuilder

		WithApiTeamRole(team *bizmodel.SysTeamRole) TeamRoleBuilder

		WithCreateTeamBo(team *teamapi.CreateTeamRequest, LeaderId uint32) TeamRequestBuilder

		WithUpdateTeamBo(team *teamapi.UpdateTeamRequest) TeamRequestBuilder

		WithListTeamBo(team *teamapi.ListTeamRequest) TeamRequestBuilder

		WithListTeamTeamMemberBo(team *teamapi.ListTeamMemberRequest) TeamRequestBuilder

		WithAddTeamMemberBo(team *teamapi.AddTeamMemberRequest) TeamRequestBuilder

		WithLeaderIdBo(leaderId uint32) TeamRequestBuilder

		WithApiTeamMember(teamMember *bizmodel.SysTeamMember) TeamMemberBuilder

		WithApiUserBo(user *model.SysUser) UserModelBuilder

		WithCreateUserBo(user *userapi.CreateUserRequest) UserRequestBuilder

		WithUpdateUserBo(user *userapi.UpdateUserRequest) UserRequestBuilder

		WithApiDatasourceMetric(metric *bizmodel.DatasourceMetric) DatasourceMetricModelBuilder

		WithApiDatasourceMetricLabel(metric *bizmodel.MetricLabel) DatasourceMetricLabelModelBuilder

		WithApiDatasourceMetricLabelValue(metric *bizmodel.MetricLabelValue) DatasourceMetricLabelValueBuilder
	}
)

func (b *builder) WithBoDatasourceQueryData(d *bo.DatasourceQueryData) DatasourceQueryDataBuilder {
	return &datasourceQueryDataBuilder{
		DatasourceQueryData: d,
		ctx:                 b.ctx,
	}
}

func (b *builder) WithDoDatasource(d *bizmodel.Datasource) DatasourceModelBuilder {
	return &datasourceBuilder{
		Datasource: d,
		ctx:        b.ctx,
	}
}
func (b *builder) WithApiTemplateStrategy(template *model.StrategyTemplate) TemplateModelBuilder {
	return &templateStrategyBuilder{
		StrategyTemplate: template,
		ctx:              b.ctx,
	}
}

func (b *builder) WithCreateBoTemplateStrategy(template *strategyapi.CreateTemplateStrategyRequest) TemplateRequestBuilder {
	return &templateStrategyBuilder{
		CreateStrategy: template,
		ctx:            b.ctx,
	}
}

func (b *builder) WithUpdateBoTemplateStrategy(template *strategyapi.UpdateTemplateStrategyRequest) TemplateRequestBuilder {
	return &templateStrategyBuilder{
		UpdateStrategy: template,
		ctx:            b.ctx,
	}
}

func (b *builder) WithApiTemplateStrategyLevel(template *model.StrategyLevelTemplate) TemplateLevelBuilder {
	return &templateStrategyLevelBuilder{
		StrategyLevelTemplate: template,
		ctx:                   b.ctx,
	}
}

func (b *builder) WithApiStrategy(strategy *bizmodel.Strategy) StrategyModelBuilder {
	return &strategyBuilder{
		Strategy: strategy,
		ctx:      b.ctx,
	}
}

func (b *builder) WithCreateBoStrategy(strategy *strategyapi.CreateStrategyRequest) StrategyRequestBuilder {
	return &strategyBuilder{
		CreateStrategy: strategy,
		ctx:            b.ctx,
	}
}

func (b *builder) WithUpdateBoStrategy(strategy *strategyapi.UpdateStrategyRequest) StrategyRequestBuilder {
	return &strategyBuilder{
		UpdateStrategy: strategy,
		ctx:            b.ctx,
	}
}

func (b *builder) WithCreateBoDict(dict *dictapi.CreateDictRequest) DictRequestBuilder {
	return &dictBuilder{
		CreateDictRequest: dict,
		ctx:               b.ctx,
	}
}

func (b *builder) WithUpdateBoDict(dict *dictapi.UpdateDictRequest) DictRequestBuilder {
	return &dictBuilder{
		UpdateDictRequest: dict,
		ctx:               b.ctx,
	}
}

func (b *builder) WithApiDict(dict *model.SysDict) DictModelBuilder {
	return &dictBuilder{
		SysDict: dict,
		ctx:     b.ctx,
	}
}

func (b *builder) WithApiDictSelect(dict *model.SysDict) DictModelBuilder {
	return &dictBuilder{
		SysDict: dict,
		ctx:     b.ctx,
	}
}

func (b *builder) WithCreateMenuBo(menu *menuapi.CreateMenuRequest) MenuRequestBuilder {
	return &menuBuilder{
		CreateMenuRequest: menu,
		ctx:               b.ctx,
	}
}

func (b *builder) WithUpdateMenuBo(menu *menuapi.UpdateMenuRequest) MenuRequestBuilder {
	return &menuBuilder{
		UpdateMenuRequest: menu,
		ctx:               b.ctx,
	}
}

func (b *builder) WithApiMenu(menu *model.SysMenu) MenuModelBuilder {
	return &menuBuilder{
		Menu: menu,
		ctx:  b.ctx,
	}
}

func (b *builder) WithBatchCreateMenuBo(menu *menuapi.BatchCreateMenuRequest) MenuRequestBuilder {
	return &menuBuilder{
		BatchCreateMenuRequest: menu,
		ctx:                    b.ctx,
	}
}

func (b *builder) WithApiMenuTree(menuList []*admin.Menu, parentID uint32) MenuTreeBuilder {
	menuMap := make(map[uint32][]*admin.Menu)
	// 按照父级ID分组
	for _, menu := range menuList {
		if _, ok := menuMap[menu.GetParentId()]; !ok {
			menuMap[menu.GetParentId()] = make([]*admin.Menu, 0)
		}
		menuMap[menu.GetParentId()] = append(menuMap[menu.GetParentId()], menu)
	}
	return &menuTreeBuilder{
		MenuMap:  menuMap,
		ParentID: parentID,
		ctx:      b.ctx,
	}
}

func (b *builder) WithApiTeam(team *model.SysTeam) TeamModelBuilder {
	return &teamBuilder{
		SysTeam: team,
		ctx:     b.ctx,
	}
}

func (b *builder) WithSelectTeamRole(team *bizmodel.SysTeamRole) TeamRoleBuilder {
	return &teamRoleBuilder{
		SysTeamRole: team,
		ctx:         b.ctx,
	}
}

func (b *builder) WithApiTeamRole(team *bizmodel.SysTeamRole) TeamRoleBuilder {
	return &teamRoleBuilder{
		SysTeamRole: team,
		ctx:         b.ctx,
	}
}

func (b *builder) WithCreateTeamBo(req *teamapi.CreateTeamRequest, leaderId uint32) TeamRequestBuilder {
	return &teamBuilder{
		CreateRoleRequest: req,
		ctx:               b.ctx,
		LeaderId:          leaderId,
	}
}

func (b *builder) WithUpdateTeamBo(req *teamapi.UpdateTeamRequest) TeamRequestBuilder {
	return &teamBuilder{
		UpdateTeamRequest: req,
		ctx:               b.ctx,
	}
}

func (b *builder) WithListTeamBo(req *teamapi.ListTeamRequest) TeamRequestBuilder {
	return &teamBuilder{
		ListTeamRequest: req,
		ctx:             b.ctx,
	}
}

func (b *builder) WithListTeamTeamMemberBo(req *teamapi.ListTeamMemberRequest) TeamRequestBuilder {
	return &teamBuilder{
		ListTeamMemberRequest: req,
		ctx:                   b.ctx,
	}
}

func (b *builder) WithAddTeamMemberBo(req *teamapi.AddTeamMemberRequest) TeamRequestBuilder {
	return &teamBuilder{
		AddTeamMemberRequest: req,
		ctx:                  b.ctx,
	}
}

func (b *builder) WithApiTeamMember(teamMember *bizmodel.SysTeamMember) TeamMemberBuilder {
	return &teamMemberBuilder{
		SysTeamMember: teamMember,
		ctx:           b.ctx,
	}
}

func (b *builder) WithApiUserBo(user *model.SysUser) UserModelBuilder {
	return &userBuilder{
		SysUser: user,
		ctx:     b.ctx,
	}
}

func (b *builder) WithCreateUserBo(req *userapi.CreateUserRequest) UserRequestBuilder {
	return &userBuilder{
		CreateUserRequest: req,
		ctx:               b.ctx,
	}
}

func (b *builder) WithUpdateUserBo(req *userapi.UpdateUserRequest) UserRequestBuilder {
	return &userBuilder{
		UpdateUserRequest: req,
		ctx:               b.ctx,
	}
}

func (b *builder) WithCreateDatasourceBo(req *datasourceapi.CreateDatasourceRequest) DatasourceRequestBuilder {
	return &datasourceBuilder{
		CreateDatasourceRequest: req, ctx: b.ctx}
}

func (b *builder) WithUpdateDatasourceBo(req *datasourceapi.UpdateDatasourceRequest) DatasourceRequestBuilder {
	return &datasourceBuilder{
		UpdateDatasourceRequest: req,
		ctx:                     b.ctx,
	}
}

func (b *builder) WithListDatasourceBo(req *datasourceapi.ListDatasourceRequest) DatasourceRequestBuilder {
	return &datasourceBuilder{
		ListDatasourceRequest: req,
		ctx:                   b.ctx,
	}
}

func (b *builder) WithLeaderIdBo(leaderId uint32) TeamRequestBuilder {
	return &teamBuilder{
		LeaderId: leaderId,
		ctx:      b.ctx,
	}
}

func (b *builder) WithApiDatasourceMetric(metric *bizmodel.DatasourceMetric) DatasourceMetricModelBuilder {
	return &datasourceMetricModelBuilder{
		DatasourceMetric: metric,
		ctx:              b.ctx,
	}
}

func (b *builder) WithApiDatasourceMetricLabel(metric *bizmodel.MetricLabel) DatasourceMetricLabelModelBuilder {
	return &datasourceMetricLabelModelBuilder{
		MetricLabel: metric,
		ctx:         b.ctx,
	}
}

func (b *builder) WithApiDatasourceMetricLabelValue(metric *bizmodel.MetricLabelValue) DatasourceMetricLabelValueBuilder {

	return &datasourceMetricLabelValueBuilder{
		MetricLabelValue: metric,
		ctx:              b.ctx,
	}
}

func (b *builder) WithApiStrategyLevel(strategy *bizmodel.StrategyLevel) StrategyLevelModelBuilder {
	return &strategyLevelBuilder{
		StrategyLevel: strategy,
		ctx:           b.ctx,
	}

}
func (b *builder) WithContext(ctx context.Context) Builder {
	b.ctx = ctx
	return b
}
