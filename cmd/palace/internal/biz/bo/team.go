package bo

import (
	"github.com/google/uuid"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type CreateTeamRequest interface {
	GetName() string
	GetRemark() string
	GetLogo() string
	GetLeader() do.User
	GetCapacity() vobj.TeamCapacity
	GetBizDBConfig() *config.Database
	GetAlarmDBConfig() *config.Database
	GetUUID() uuid.UUID
	GetStatus() vobj.TeamStatus
}

type UpdateTeamRequest interface {
	GetTeam() do.Team
	GetName() string
	GetRemark() string
	GetLogo() string
}

func (o *SaveOneTeamRequest) WithCreateTeamRequest(leader do.User) CreateTeamRequest {
	o.leader = leader
	return o
}

func (o *SaveOneTeamRequest) WithUpdateTeamRequest(team do.Team) UpdateTeamRequest {
	o.teamDo = team
	return o
}

type SaveOneTeamRequest struct {
	TeamID     uint32
	TeamName   string
	TeamRemark string
	TeamLogo   string

	leader do.User
	teamDo do.Team
}

func (o *SaveOneTeamRequest) GetCapacity() vobj.TeamCapacity {
	if o == nil {
		return vobj.TeamCapacityUnknown
	}
	return vobj.TeamCapacityDefault
}

func (o *SaveOneTeamRequest) GetBizDBConfig() *config.Database {
	return nil
}

func (o *SaveOneTeamRequest) GetAlarmDBConfig() *config.Database {
	return nil
}

func (o *SaveOneTeamRequest) GetUUID() uuid.UUID {
	return uuid.New()
}

func (o *SaveOneTeamRequest) GetStatus() vobj.TeamStatus {
	return vobj.TeamStatusNormal
}

func (o *SaveOneTeamRequest) GetLeader() do.User {
	return o.leader
}

func (o *SaveOneTeamRequest) GetTeam() do.Team {
	return o.teamDo
}

func (o *SaveOneTeamRequest) GetName() string {
	return o.TeamName
}

func (o *SaveOneTeamRequest) GetRemark() string {
	return o.TeamRemark
}

func (o *SaveOneTeamRequest) GetLogo() string {
	return o.TeamLogo
}

type TeamListRequest struct {
	*PaginationRequest
	Keyword   string
	Status    []vobj.TeamStatus
	UserIds   []uint32
	LeaderId  uint32
	CreatorId uint32
}

func (r *TeamListRequest) ToTeamListReply(items []*system.Team) *TeamListReply {
	return &TeamListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(items, func(team *system.Team) do.Team { return team }),
	}
}

type TeamListReply = ListReply[do.Team]

type TeamMemberListRequest struct {
	*PaginationRequest
	Keyword   string
	Status    []vobj.MemberStatus
	Positions []vobj.Role
	TeamId    uint32
}

func (r *TeamMemberListRequest) ToTeamMemberListReply(items []*system.TeamMember) *TeamMemberListReply {
	return &TeamMemberListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(items, func(member *system.TeamMember) do.TeamMember { return member }),
	}
}

type TeamMemberListReply = ListReply[do.TeamMember]

type UpdateMemberPosition interface {
	GetMember() do.TeamMember
	GetPosition() vobj.Role
}

type UpdateMemberPositionReq struct {
	operator do.TeamMember
	member   do.TeamMember
	MemberID uint32
	Position vobj.Role
}

func (r *UpdateMemberPositionReq) GetMember() do.TeamMember {
	if r == nil {
		return nil
	}
	return r.member
}

func (r *UpdateMemberPositionReq) GetPosition() vobj.Role {
	if r == nil {
		return vobj.RoleUnknown
	}
	return r.Position
}

func (r *UpdateMemberPositionReq) WithOperator(operator do.TeamMember) *UpdateMemberPositionReq {
	r.operator = operator
	return r
}

func (r *UpdateMemberPositionReq) WithMember(member do.TeamMember) *UpdateMemberPositionReq {
	r.member = member
	return r
}

func (r *UpdateMemberPositionReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if validate.IsNil(r.member) {
		return merr.ErrorParamsError("invalid member")
	}
	if r.Position.IsUnknown() {
		return merr.ErrorParamsError("invalid position")
	}
	operatorPosition := r.operator.GetPosition()
	if !(operatorPosition.GT(r.Position) && operatorPosition.IsAdminOrSuperAdmin()) {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}

type UpdateMemberStatus interface {
	GetMembers() []do.TeamMember
	GetStatus() vobj.MemberStatus
}

type UpdateMemberStatusReq struct {
	operator  do.TeamMember
	members   []do.TeamMember
	MemberIds []uint32
	Status    vobj.MemberStatus
}

func (r *UpdateMemberStatusReq) GetMembers() []do.TeamMember {
	if r == nil {
		return nil
	}
	return r.members
}

func (r *UpdateMemberStatusReq) GetStatus() vobj.MemberStatus {
	if r == nil {
		return vobj.MemberStatusUnknown
	}
	return r.Status
}

func (r *UpdateMemberStatusReq) WithOperator(operator do.TeamMember) *UpdateMemberStatusReq {
	r.operator = operator
	return r
}

func (r *UpdateMemberStatusReq) WithMembers(members []do.TeamMember) *UpdateMemberStatusReq {
	r.members = slices.MapFilter(members, func(member do.TeamMember) (do.TeamMember, bool) {
		if validate.IsNil(member) || member.GetID() <= 0 {
			return nil, false
		}
		return member, true
	})
	return r
}

func (r *UpdateMemberStatusReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if len(r.members) == 0 {
		return merr.ErrorParamsError("invalid members")
	}
	if r.Status.IsUnknown() {
		return merr.ErrorParamsError("invalid status")
	}
	operatorPosition := r.operator.GetPosition()
	for _, member := range r.members {
		if !(operatorPosition.GT(member.GetPosition()) && operatorPosition.IsAdminOrSuperAdmin()) {
			return merr.ErrorParamsError("invalid position")
		}
	}
	return nil
}

type UpdateMemberRoles interface {
	GetMember() do.TeamMember
	GetRoles() []do.TeamRole
}

type UpdateMemberRolesReq struct {
	operator do.TeamMember
	member   do.TeamMember
	roles    []do.TeamRole
	MemberId uint32
	RoleIds  []uint32
}

func (r *UpdateMemberRolesReq) GetMember() do.TeamMember {
	if r == nil {
		return nil
	}
	return r.member
}

func (r *UpdateMemberRolesReq) GetRoles() []do.TeamRole {
	if r == nil {
		return nil
	}
	return r.roles
}

func (r *UpdateMemberRolesReq) WithOperator(operator do.TeamMember) *UpdateMemberRolesReq {
	r.operator = operator
	return r
}

func (r *UpdateMemberRolesReq) WithMember(member do.TeamMember) *UpdateMemberRolesReq {
	r.member = member
	return r
}

func (r *UpdateMemberRolesReq) WithRoles(roles []do.TeamRole) *UpdateMemberRolesReq {
	r.roles = slices.MapFilter(roles, func(role do.TeamRole) (do.TeamRole, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return nil, false
		}
		return role, true
	})
	return r
}

func (r *UpdateMemberRolesReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if validate.IsNil(r.member) {
		return merr.ErrorParamsError("invalid member")
	}
	operatorPosition := r.operator.GetPosition()
	if !(operatorPosition.GT(r.member.GetPosition()) && operatorPosition.IsAdminOrSuperAdmin()) {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}

type RemoveMemberReq struct {
	operator do.TeamMember
	member   do.TeamMember
	MemberID uint32
}

func (r *RemoveMemberReq) GetMembers() []do.TeamMember {
	if r == nil {
		return nil
	}
	return []do.TeamMember{r.member}
}

func (r *RemoveMemberReq) GetStatus() vobj.MemberStatus {
	if r == nil {
		return vobj.MemberStatusUnknown
	}
	return vobj.MemberStatusDeleted
}

func (r *RemoveMemberReq) WithOperator(operator do.TeamMember) *RemoveMemberReq {
	r.operator = operator
	return r
}

func (r *RemoveMemberReq) WithMember(member do.TeamMember) *RemoveMemberReq {
	r.member = member
	return r
}

func (r *RemoveMemberReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if validate.IsNil(r.member) {
		return merr.ErrorParamsError("invalid member")
	}
	operatorPosition := r.operator.GetPosition()
	if !(operatorPosition.GT(r.member.GetPosition()) && operatorPosition.IsAdminOrSuperAdmin()) {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}

type InviteMember interface {
	GetTeam() do.Team
	GetInviteUser() do.User
	GetPosition() vobj.Role
	GetRoles() []do.TeamRole
	GetSendEmailFun() SendEmailFun
	GetOperator() do.TeamMember
}

type InviteMemberReq struct {
	operator     do.TeamMember
	team         do.Team
	invitee      do.User
	roles        []do.TeamRole
	UserEmail    string
	Position     vobj.Role
	RoleIds      []uint32
	SendEmailFun SendEmailFun
}

func (r *InviteMemberReq) GetTeam() do.Team {
	if r == nil {
		return nil
	}
	return r.team
}

func (r *InviteMemberReq) GetInviteUser() do.User {
	if r == nil {
		return nil
	}
	return r.invitee
}

func (r *InviteMemberReq) GetPosition() vobj.Role {
	if r == nil {
		return vobj.RoleUnknown
	}
	return r.Position
}

func (r *InviteMemberReq) GetRoles() []do.TeamRole {
	if r == nil {
		return nil
	}
	return r.roles
}

func (r *InviteMemberReq) GetSendEmailFun() SendEmailFun {
	if r == nil {
		return nil
	}
	return r.SendEmailFun
}

func (r *InviteMemberReq) GetOperator() do.TeamMember {
	if r == nil {
		return nil
	}
	return r.operator
}

func (r *InviteMemberReq) Validate() error {
	if validate.IsNil(r.team) {
		return merr.ErrorParamsError("invalid team")
	}
	if validate.IsNil(r.invitee) {
		return merr.ErrorParamsError("invalid invitee")
	}
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if r.Position.IsUnknown() {
		return merr.ErrorParamsError("invalid position")
	}
	if !r.operator.GetPosition().IsAdminOrSuperAdmin() {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}

func (r *InviteMemberReq) WithTeam(team do.Team) *InviteMemberReq {
	r.team = team
	return r
}

func (r *InviteMemberReq) WithInviteUser(invitee do.User) *InviteMemberReq {
	r.invitee = invitee
	return r
}

func (r *InviteMemberReq) WithOperator(operator do.TeamMember) *InviteMemberReq {
	r.operator = operator
	return r
}

func (r *InviteMemberReq) WithRoles(roles []do.TeamRole) *InviteMemberReq {
	r.roles = slices.MapFilter(roles, func(role do.TeamRole) (do.TeamRole, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return nil, false
		}
		return role, true
	})
	return r
}

type CreateTeamMemberReq struct {
	Team     do.Team
	User     do.User
	Status   vobj.MemberStatus
	Position vobj.Role
}
