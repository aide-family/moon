package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ ITeamModuleBuilder = (*teamModuleBuilder)(nil)

type (
	teamModuleBuilder struct {
		ctx context.Context
	}

	// ITeamModuleBuilder 团队模块条目构造器
	ITeamModuleBuilder interface {
		// WithCreateTeamRequest 设置创建团队请求
		WithCreateTeamRequest(*teamapi.CreateTeamRequest) ICreateTeamRequestBuilder
		// WithUpdateTeamRequest 设置更新团队请求
		WithUpdateTeamRequest(*teamapi.UpdateTeamRequest) IUpdateTeamRequestBuilder
		// WithListTeamRequest 设置获取团队列表请求
		WithListTeamRequest(*teamapi.ListTeamRequest) IListTeamRequestBuilder
		// WithRemoveTeamMemberRequest 设置移除团队成员请求
		WithRemoveTeamMemberRequest(*teamapi.RemoveTeamMemberRequest) IRemoveTeamMemberRequestBuilder
		// WithSetTeamAdminRequest 设置设置团队管理员请求
		WithSetTeamAdminRequest(*teamapi.SetTeamAdminRequest) ISetTeamAdminRequestBuilder
		// WithRemoveTeamAdminRequest 设置移除团队管理员请求
		WithRemoveTeamAdminRequest(*teamapi.RemoveTeamAdminRequest) IRemoveTeamAdminRequestBuilder
		// WithSetMemberRoleRequest 设置设置成员角色请求
		WithSetMemberRoleRequest(*teamapi.SetMemberRoleRequest) ISetMemberRoleRequestBuilder
		// WithListTeamMemberRequest 设置获取团队成员列表请求
		WithListTeamMemberRequest(*teamapi.ListTeamMemberRequest) IListTeamMemberRequestBuilder
		// WithTransferTeamLeaderRequest 设置转移团队领导请求
		WithTransferTeamLeaderRequest(*teamapi.TransferTeamLeaderRequest) ITransferTeamLeaderRequestBuilder
		// WithSetTeamMailConfigRequest 设置设置团队邮箱配置请求
		WithSetTeamMailConfigRequest(*teamapi.SetTeamMailConfigRequest) ISetTeamMailConfigRequestBuilder
		// DoTeamBuilder 获取团队条目构造器
		DoTeamBuilder() IDoTeamBuilder
	}

	// ICreateTeamRequestBuilder 创建团队请求参数构造器
	ICreateTeamRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateTeamParams
	}

	createTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.CreateTeamRequest
	}

	// IUpdateTeamRequestBuilder 更新团队请求参数构造器
	IUpdateTeamRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateTeamParams
	}

	updateTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.UpdateTeamRequest
	}

	// IListTeamRequestBuilder 获取团队列表请求参数构造器
	IListTeamRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryTeamListParams
	}

	listTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.ListTeamRequest
	}

	// IAddTeamMemberRequestBuilder 添加团队成员请求参数构造器
	IAddTeamMemberRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.AddTeamMemberParams
	}

	// IRemoveTeamMemberRequestBuilder 移除团队成员请求参数构造器
	IRemoveTeamMemberRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.RemoveTeamMemberParams
	}

	removeTeamMemberRequestBuilder struct {
		ctx context.Context
		*teamapi.RemoveTeamMemberRequest
	}

	// ISetTeamAdminRequestBuilder 设置团队管理员请求参数构造器
	ISetTeamAdminRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.SetMemberAdminParams
	}

	setTeamAdminRequestBuilder struct {
		ctx context.Context
		*teamapi.SetTeamAdminRequest
	}

	// IRemoveTeamAdminRequestBuilder 移除团队管理员请求参数构造器
	IRemoveTeamAdminRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.SetMemberAdminParams
	}

	removeTeamAdminRequestBuilder struct {
		ctx context.Context
		*teamapi.RemoveTeamAdminRequest
	}

	// ISetMemberRoleRequestBuilder 设置成员角色请求参数构造器
	ISetMemberRoleRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.SetMemberRoleParams
	}

	setMemberRoleRequestBuilder struct {
		ctx context.Context
		*teamapi.SetMemberRoleRequest
	}

	// IListTeamMemberRequestBuilder 获取团队成员列表请求参数构造器
	IListTeamMemberRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.ListTeamMemberParams
	}

	listTeamMemberRequestBuilder struct {
		ctx context.Context
		*teamapi.ListTeamMemberRequest
	}

	// ITransferTeamLeaderRequestBuilder 转移团队领导请求参数构造器
	ITransferTeamLeaderRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.TransferTeamLeaderParams
	}

	transferTeamLeaderRequestBuilder struct {
		ctx context.Context
		*teamapi.TransferTeamLeaderRequest
	}

	// ISetTeamMailConfigRequestBuilder 设置团队邮箱配置请求参数构造器
	ISetTeamMailConfigRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.SetTeamMailConfigParams
	}

	setTeamMailConfigRequestBuilder struct {
		ctx context.Context
		*teamapi.SetTeamMailConfigRequest
	}

	// IDoTeamBuilder 团队条目构造器
	IDoTeamBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.SysTeam, ...map[uint32]*adminapi.UserItem) *adminapi.TeamItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*model.SysTeam) []*adminapi.TeamItem
		// ToSelect 转换为选择对象
		ToSelect(*model.SysTeam) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*model.SysTeam) []*adminapi.SelectItem
	}

	doTeamBuilder struct {
		ctx context.Context
	}
)

func (s *setTeamMailConfigRequestBuilder) ToBo() *bo.SetTeamMailConfigParams {
	if types.IsNil(s) || types.IsNil(s.SetTeamMailConfigRequest) {
		return nil
	}
	config := s.GetConfig()
	return &bo.SetTeamMailConfigParams{
		User:     config.GetUser(),
		Password: config.GetPass(),
		Host:     config.GetHost(),
		Port:     config.GetPort(),
		Remark:   s.GetRemark(),
	}
}

func (t *teamModuleBuilder) WithSetTeamMailConfigRequest(request *teamapi.SetTeamMailConfigRequest) ISetTeamMailConfigRequestBuilder {
	if types.IsNil(t) || types.IsNil(request) {
		return nil
	}
	return &setTeamMailConfigRequestBuilder{ctx: t.ctx, SetTeamMailConfigRequest: request}
}

func (d *doTeamBuilder) ToAPI(team *model.SysTeam, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.TeamItem {
	if types.IsNil(d) || types.IsNil(team) {
		return nil
	}
	userMap := getUsers(d.ctx, userMaps, append(team.Admins, team.LeaderID, team.CreatorID)...)
	admins := make([]*adminapi.UserItem, 0, len(team.Admins))
	for _, adminID := range team.Admins {
		admins = append(admins, userMap[adminID])
	}

	return &adminapi.TeamItem{
		Id:        team.ID,
		Name:      team.Name,
		Status:    api.Status(team.Status),
		Remark:    team.Remark,
		CreatedAt: team.CreatedAt.String(),
		UpdatedAt: team.UpdatedAt.String(),
		Leader:    userMap[team.LeaderID],
		Creator:   userMap[team.CreatorID],
		Logo:      team.Logo,
		Admins:    admins,
	}
}

func (d *doTeamBuilder) ToAPIs(teams []*model.SysTeam) []*adminapi.TeamItem {
	if types.IsNil(d) || types.IsNil(teams) {
		return nil
	}
	ids := types.SliceTo(teams, func(item *model.SysTeam) uint32 { return item.CreatorID })
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(teams, func(item *model.SysTeam) *adminapi.TeamItem {
		return d.ToAPI(item, userMap)
	})
}

func (d *doTeamBuilder) ToSelect(team *model.SysTeam) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(team) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    team.ID,
		Label:    team.Name,
		Children: nil,
		Disabled: team.DeletedAt > 0 || !team.Status.IsEnable(),
		Extend:   &adminapi.SelectExtend{Remark: team.Remark, Image: team.Logo},
	}
}

func (d *doTeamBuilder) ToSelects(teams []*model.SysTeam) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(teams) {
		return nil
	}

	return types.SliceTo(teams, func(item *model.SysTeam) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (t *transferTeamLeaderRequestBuilder) ToBo() *bo.TransferTeamLeaderParams {
	if types.IsNil(t) || types.IsNil(t.TransferTeamLeaderRequest) {
		return nil
	}

	return &bo.TransferTeamLeaderParams{
		LeaderID:    t.GetMemberID(),
		OldLeaderID: middleware.GetUserID(t.ctx),
	}
}

func (l *listTeamMemberRequestBuilder) ToBo() *bo.ListTeamMemberParams {
	if types.IsNil(l) || types.IsNil(l.ListTeamMemberRequest) {
		return nil
	}

	return &bo.ListTeamMemberParams{
		Page:      types.NewPagination(l.GetPagination()),
		Keyword:   l.GetKeyword(),
		Role:      vobj.Role(l.GetRole()),
		Gender:    vobj.Gender(l.GetGender()),
		Status:    vobj.Status(l.GetStatus()),
		MemberIDs: l.GetMemberIds(),
	}
}

func (s *setMemberRoleRequestBuilder) ToBo() *bo.SetMemberRoleParams {
	if types.IsNil(s) || types.IsNil(s.SetMemberRoleRequest) {
		return nil
	}

	return &bo.SetMemberRoleParams{
		MemberID: s.GetMemberID(),
		RoleIDs:  s.GetRoles(),
	}
}

func (r *removeTeamAdminRequestBuilder) ToBo() *bo.SetMemberAdminParams {
	if types.IsNil(r) || types.IsNil(r.RemoveTeamAdminRequest) {
		return nil
	}

	return &bo.SetMemberAdminParams{
		MemberIDs: r.GetMemberIds(),
		Role:      vobj.RoleUser,
	}
}

func (s *setTeamAdminRequestBuilder) ToBo() *bo.SetMemberAdminParams {
	if types.IsNil(s) || types.IsNil(s.SetTeamAdminRequest) {
		return nil
	}

	return &bo.SetMemberAdminParams{
		MemberIDs: s.GetMemberIds(),
		Role:      vobj.RoleAdmin,
	}
}

func (r *removeTeamMemberRequestBuilder) ToBo() *bo.RemoveTeamMemberParams {
	if types.IsNil(r) || types.IsNil(r.RemoveTeamMemberRequest) {
		return nil
	}

	return &bo.RemoveTeamMemberParams{
		MemberIds: []uint32{r.GetMemberID()},
	}
}

func (l *listTeamRequestBuilder) ToBo() *bo.QueryTeamListParams {
	if types.IsNil(l) || types.IsNil(l.ListTeamRequest) {
		return nil
	}

	return &bo.QueryTeamListParams{
		Page:      types.NewPagination(l.GetPagination()),
		Keyword:   l.GetKeyword(),
		Status:    vobj.Status(l.GetStatus()),
		CreatorID: l.GetCreatorId(),
		LeaderID:  l.GetLeaderId(),
		UserID:    middleware.GetUserID(l.ctx),
		IDs:       l.GetIds(),
	}
}

func (u *updateTeamRequestBuilder) ToBo() *bo.UpdateTeamParams {
	if types.IsNil(u) || types.IsNil(u.UpdateTeamRequest) {
		return nil
	}

	return &bo.UpdateTeamParams{
		ID:     u.GetId(),
		Name:   u.GetName(),
		Remark: u.GetRemark(),
		Status: vobj.Status(u.GetStatus()),
		Logo:   u.GetLogo(),
	}
}

func (c *createTeamRequestBuilder) ToBo() *bo.CreateTeamParams {
	if types.IsNil(c) || types.IsNil(c.CreateTeamRequest) {
		return nil
	}

	if c.GetLeaderId() <= 0 {
		c.LeaderId = middleware.GetUserID(c.ctx)
	}

	return &bo.CreateTeamParams{
		Name:     c.GetName(),
		Remark:   c.GetRemark(),
		Logo:     c.GetLogo(),
		Status:   vobj.Status(c.GetStatus()),
		LeaderID: c.GetLeaderId(),
		Admins:   c.GetAdminIds(),
	}
}

func (t *teamModuleBuilder) WithCreateTeamRequest(request *teamapi.CreateTeamRequest) ICreateTeamRequestBuilder {
	return &createTeamRequestBuilder{ctx: t.ctx, CreateTeamRequest: request}
}

func (t *teamModuleBuilder) WithUpdateTeamRequest(request *teamapi.UpdateTeamRequest) IUpdateTeamRequestBuilder {
	return &updateTeamRequestBuilder{ctx: t.ctx, UpdateTeamRequest: request}
}

func (t *teamModuleBuilder) WithListTeamRequest(request *teamapi.ListTeamRequest) IListTeamRequestBuilder {
	return &listTeamRequestBuilder{ctx: t.ctx, ListTeamRequest: request}
}

func (t *teamModuleBuilder) WithRemoveTeamMemberRequest(request *teamapi.RemoveTeamMemberRequest) IRemoveTeamMemberRequestBuilder {
	return &removeTeamMemberRequestBuilder{ctx: t.ctx, RemoveTeamMemberRequest: request}
}

func (t *teamModuleBuilder) WithSetTeamAdminRequest(request *teamapi.SetTeamAdminRequest) ISetTeamAdminRequestBuilder {
	return &setTeamAdminRequestBuilder{ctx: t.ctx, SetTeamAdminRequest: request}
}

func (t *teamModuleBuilder) WithRemoveTeamAdminRequest(request *teamapi.RemoveTeamAdminRequest) IRemoveTeamAdminRequestBuilder {
	return &removeTeamAdminRequestBuilder{ctx: t.ctx, RemoveTeamAdminRequest: request}
}

func (t *teamModuleBuilder) WithSetMemberRoleRequest(request *teamapi.SetMemberRoleRequest) ISetMemberRoleRequestBuilder {
	return &setMemberRoleRequestBuilder{ctx: t.ctx, SetMemberRoleRequest: request}
}

func (t *teamModuleBuilder) WithListTeamMemberRequest(request *teamapi.ListTeamMemberRequest) IListTeamMemberRequestBuilder {
	return &listTeamMemberRequestBuilder{ctx: t.ctx, ListTeamMemberRequest: request}
}

func (t *teamModuleBuilder) WithTransferTeamLeaderRequest(request *teamapi.TransferTeamLeaderRequest) ITransferTeamLeaderRequestBuilder {
	return &transferTeamLeaderRequestBuilder{ctx: t.ctx, TransferTeamLeaderRequest: request}
}

func (t *teamModuleBuilder) DoTeamBuilder() IDoTeamBuilder {
	return &doTeamBuilder{ctx: t.ctx}
}
