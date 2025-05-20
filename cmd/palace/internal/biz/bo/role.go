package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type Role interface {
	GetID() uint32
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetMenus() []do.Menu
	GetMenuIds() []uint32
}

type SaveTeamRoleReq struct {
	teamRole do.TeamRole
	menus    []do.Menu
	ID       uint32
	Name     string
	Remark   string
	MenuIds  []uint32
}

func (r *SaveTeamRoleReq) GetID() uint32 {
	if r == nil {
		return 0
	}
	if validate.IsNil(r.teamRole) {
		return r.ID
	}
	return r.teamRole.GetID()
}

func (r *SaveTeamRoleReq) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveTeamRoleReq) GetRemark() string {
	if r == nil {
		return ""
	}
	return r.Remark
}

func (r *SaveTeamRoleReq) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	if validate.IsNil(r.teamRole) {
		return vobj.GlobalStatusEnable
	}
	return r.teamRole.GetStatus()
}

func (r *SaveTeamRoleReq) GetMenus() []do.Menu {
	if r == nil {
		return nil
	}
	return r.menus
}

func (r *SaveTeamRoleReq) GetMenuIds() []uint32 {
	if r == nil {
		return nil
	}
	return r.MenuIds
}

func (r *SaveTeamRoleReq) WithMenus(menus []do.Menu) Role {
	r.menus = menus
	return r
}

func (r *SaveTeamRoleReq) WithRole(role do.TeamRole) Role {
	r.teamRole = role
	return r
}

type SaveRoleReq struct {
	role    do.Role
	menus   []do.Menu
	ID      uint32
	Name    string
	Remark  string
	MenuIds []uint32
}

func (r *SaveRoleReq) GetID() uint32 {
	if r == nil {
		return 0
	}
	if validate.IsNil(r.role) {
		return r.ID
	}
	return r.role.GetID()
}

func (r *SaveRoleReq) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveRoleReq) GetRemark() string {
	if r == nil {
		return ""
	}
	return r.Remark
}

func (r *SaveRoleReq) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	if validate.IsNil(r.role) {
		return vobj.GlobalStatusEnable
	}
	return r.role.GetStatus()
}

func (r *SaveRoleReq) GetMenus() []do.Menu {
	if r == nil {
		return nil
	}
	return r.menus
}

func (r *SaveRoleReq) GetMenuIds() []uint32 {
	if r == nil {
		return nil
	}
	return r.MenuIds
}

func (r *SaveRoleReq) WithRole(role do.Role) Role {
	r.role = role
	return r
}

func (r *SaveRoleReq) WithMenus(menus []do.Menu) Role {
	r.menus = menus
	return r
}

type ListRoleReq struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *ListRoleReq) ToTeamRoleListReply(roles []do.TeamRole) *ListTeamRoleReply {
	return &ListTeamRoleReply{
		PaginationReply: r.ToReply(),
		Items:           roles,
	}
}

func (r *ListRoleReq) ToListReply(roles []do.Role) *ListRoleReply {
	return &ListRoleReply{
		PaginationReply: r.ToReply(),
		Items:           roles,
	}
}

type ListTeamRoleReply = ListReply[do.TeamRole]

type ListRoleReply = ListReply[do.Role]

type UpdateRoleStatusReq struct {
	RoleID uint32
	Status vobj.GlobalStatus
}

type UpdateRoleUsers interface {
	GetRole() do.Role
	GetUsers() []do.User
}

type UpdateRoleUsersReq struct {
	RoleID   uint32
	UserIDs  []uint32
	users    []do.User
	operator do.User
	role     do.Role
}

func (r *UpdateRoleUsersReq) GetRole() do.Role {
	if r == nil {
		return nil
	}
	return r.role
}

func (r *UpdateRoleUsersReq) GetUsers() []do.User {
	if r == nil {
		return nil
	}
	return r.users
}

func (r *UpdateRoleUsersReq) WithUsers(users []do.User) *UpdateRoleUsersReq {
	r.users = slices.MapFilter(users, func(user do.User) (do.User, bool) {
		if validate.IsNil(user) || user.GetID() <= 0 {
			return nil, false
		}
		return user, true
	})
	return r
}

func (r *UpdateRoleUsersReq) WithRole(role do.Role) *UpdateRoleUsersReq {
	r.role = role
	return r
}

func (r *UpdateRoleUsersReq) WithOperator(operator do.User) *UpdateRoleUsersReq {
	r.operator = operator
	return r
}

func (r *UpdateRoleUsersReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParams("invalid operator")
	}
	if validate.IsNil(r.role) {
		return merr.ErrorParams("invalid role")
	}
	operatorPosition := r.operator.GetPosition()
	if operatorPosition.IsSuperAdmin() {
		return nil
	}
	for _, user := range r.users {
		if !operatorPosition.GT(user.GetPosition()) || !operatorPosition.IsAdminOrSuperAdmin() {
			return merr.ErrorParams("invalid position")
		}
	}
	return nil
}
