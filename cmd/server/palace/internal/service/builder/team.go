package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/api/merr"
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

	ITeamModuleBuilder interface {
		WithCreateTeamRequest(*teamapi.CreateTeamRequest) ICreateTeamRequestBuilder
		WithUpdateTeamRequest(*teamapi.UpdateTeamRequest) IUpdateTeamRequestBuilder
		WithListTeamRequest(*teamapi.ListTeamRequest) IListTeamRequestBuilder
		WithAddTeamMemberRequest(*teamapi.AddTeamMemberRequest) IAddTeamMemberRequestBuilder
		WithRemoveTeamMemberRequest(*teamapi.RemoveTeamMemberRequest) IRemoveTeamMemberRequestBuilder
		WithSetTeamAdminRequest(*teamapi.SetTeamAdminRequest) ISetTeamAdminRequestBuilder
		WithRemoveTeamAdminRequest(*teamapi.RemoveTeamAdminRequest) IRemoveTeamAdminRequestBuilder
		WithSetMemberRoleRequest(*teamapi.SetMemberRoleRequest) ISetMemberRoleRequestBuilder
		WithListTeamMemberRequest(*teamapi.ListTeamMemberRequest) IListTeamMemberRequestBuilder
		WithTransferTeamLeaderRequest(*teamapi.TransferTeamLeaderRequest) ITransferTeamLeaderRequestBuilder
		WithSetTeamMailConfigRequest(*teamapi.SetTeamMailConfigRequest) ISetTeamMailConfigRequestBuilder
		DoTeamBuilder() IDoTeamBuilder
	}

	ICreateTeamRequestBuilder interface {
		ToBo() *bo.CreateTeamParams
	}

	createTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.CreateTeamRequest
	}

	IUpdateTeamRequestBuilder interface {
		ToBo() *bo.UpdateTeamParams
	}

	updateTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.UpdateTeamRequest
	}

	IListTeamRequestBuilder interface {
		ToBo() *bo.QueryTeamListParams
	}

	listTeamRequestBuilder struct {
		ctx context.Context
		*teamapi.ListTeamRequest
	}

	IAddTeamMemberRequestBuilder interface {
		ToBo() *bo.AddTeamMemberParams
	}

	addTeamMemberRequestBuilder struct {
		ctx context.Context
		*teamapi.AddTeamMemberRequest
	}

	IRemoveTeamMemberRequestBuilder interface {
		ToBo() *bo.RemoveTeamMemberParams
	}

	removeTeamMemberRequestBuilder struct {
		ctx context.Context
		*teamapi.RemoveTeamMemberRequest
	}

	ISetTeamAdminRequestBuilder interface {
		ToBo() *bo.SetMemberAdminParams
	}

	setTeamAdminRequestBuilder struct {
		ctx context.Context
		*teamapi.SetTeamAdminRequest
	}

	IRemoveTeamAdminRequestBuilder interface {
		ToBo() *bo.SetMemberAdminParams
	}

	removeTeamAdminRequestBuilder struct {
		ctx context.Context
		*teamapi.RemoveTeamAdminRequest
	}

	ISetMemberRoleRequestBuilder interface {
		ToBo() *bo.SetMemberRoleParams
	}

	setMemberRoleRequestBuilder struct {
		ctx context.Context
		*teamapi.SetMemberRoleRequest
	}

	IListTeamMemberRequestBuilder interface {
		ToBo() *bo.ListTeamMemberParams
	}

	listTeamMemberRequestBuilder struct {
		ctx context.Context
		*teamapi.ListTeamMemberRequest
	}

	ITransferTeamLeaderRequestBuilder interface {
		ToBo() *bo.TransferTeamLeaderParams
	}

	transferTeamLeaderRequestBuilder struct {
		ctx context.Context
		*teamapi.TransferTeamLeaderRequest
	}

	ISetTeamMailConfigRequestBuilder interface {
		ToBo() *bo.SetTeamMailConfigParams
	}

	setTeamMailConfigRequestBuilder struct {
		ctx context.Context
		*teamapi.SetTeamMailConfigRequest
	}

	IDoTeamBuilder interface {
		ToAPI(*model.SysTeam) *adminapi.TeamItem
		ToAPIs([]*model.SysTeam) []*adminapi.TeamItem
		ToSelect(*model.SysTeam) *adminapi.SelectItem
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
		TeamID:   s.GetId(),
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

func (d *doTeamBuilder) ToAPI(team *model.SysTeam) *adminapi.TeamItem {
	if types.IsNil(d) || types.IsNil(team) {
		return nil
	}

	return &adminapi.TeamItem{
		Id:        team.ID,
		Name:      team.Name,
		Status:    api.Status(team.Status),
		Remark:    team.Remark,
		CreatedAt: team.CreatedAt.String(),
		UpdatedAt: team.UpdatedAt.String(),
		Leader:    nil, // TODO user
		Creator:   nil, // TODO user
		Logo:      team.Logo,
		Admin:     nil, // TODO user
	}
}

func (d *doTeamBuilder) ToAPIs(teams []*model.SysTeam) []*adminapi.TeamItem {
	if types.IsNil(d) || types.IsNil(teams) {
		return nil
	}

	return types.SliceTo(teams, func(item *model.SysTeam) *adminapi.TeamItem {
		return d.ToAPI(item)
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

	claims, ok := middleware.ParseJwtClaims(t.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(t.ctx))
	}

	return &bo.TransferTeamLeaderParams{
		ID:          t.GetId(),
		LeaderID:    t.GetUserId(),
		OldLeaderID: claims.GetUser(),
	}
}

func (l *listTeamMemberRequestBuilder) ToBo() *bo.ListTeamMemberParams {
	if types.IsNil(l) || types.IsNil(l.ListTeamMemberRequest) {
		return nil
	}

	return &bo.ListTeamMemberParams{
		Page:      types.NewPagination(l.GetPagination()),
		ID:        l.GetId(),
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
		ID:       s.GetId(),
		MemberID: s.GetUserId(),
		RoleIDs:  s.GetRoles(),
	}
}

func (r *removeTeamAdminRequestBuilder) ToBo() *bo.SetMemberAdminParams {
	if types.IsNil(r) || types.IsNil(r.RemoveTeamAdminRequest) {
		return nil
	}

	return &bo.SetMemberAdminParams{
		ID:        r.GetId(),
		MemberIDs: r.GetUserIds(),
		Role:      vobj.RoleUser,
	}
}

func (s *setTeamAdminRequestBuilder) ToBo() *bo.SetMemberAdminParams {
	if types.IsNil(s) || types.IsNil(s.SetTeamAdminRequest) {
		return nil
	}

	return &bo.SetMemberAdminParams{
		ID:        s.GetId(),
		MemberIDs: s.GetUserIds(),
		Role:      vobj.RoleAdmin,
	}
}

func (r *removeTeamMemberRequestBuilder) ToBo() *bo.RemoveTeamMemberParams {
	if types.IsNil(r) || types.IsNil(r.RemoveTeamMemberRequest) {
		return nil
	}

	return &bo.RemoveTeamMemberParams{
		ID:        r.GetId(),
		MemberIds: []uint32{r.GetUserId()},
	}
}

func (a *addTeamMemberRequestBuilder) ToBo() *bo.AddTeamMemberParams {
	if types.IsNil(a) || types.IsNil(a.AddTeamMemberRequest) {
		return nil
	}

	return &bo.AddTeamMemberParams{
		ID: a.GetId(),
		Members: types.SliceTo(a.GetMembers(), func(item *teamapi.AddTeamMemberRequest_MemberItem) *bo.AddTeamMemberItem {
			return &bo.AddTeamMemberItem{
				UserID:  item.GetUserId(),
				Role:    vobj.Role(item.GetRole()),
				RoleIDs: item.GetRoles(),
			}
		}),
	}
}

func (l *listTeamRequestBuilder) ToBo() *bo.QueryTeamListParams {
	if types.IsNil(l) || types.IsNil(l.ListTeamRequest) {
		return nil
	}

	claims, ok := middleware.ParseJwtClaims(l.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(l.ctx))
	}

	return &bo.QueryTeamListParams{
		Page:      types.NewPagination(l.GetPagination()),
		Keyword:   l.GetKeyword(),
		Status:    vobj.Status(l.GetStatus()),
		CreatorID: l.GetCreatorId(),
		LeaderID:  l.GetLeaderId(),
		UserID:    claims.GetUser(),
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

func (t *teamModuleBuilder) WithAddTeamMemberRequest(request *teamapi.AddTeamMemberRequest) IAddTeamMemberRequestBuilder {
	return &addTeamMemberRequestBuilder{ctx: t.ctx, AddTeamMemberRequest: request}
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
