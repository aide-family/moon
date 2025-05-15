package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type UserUpdateInfo struct {
	do.User
	UserID   uint32
	Nickname string
	Avatar   string
	Gender   vobj.Gender
}

func (u *UserUpdateInfo) WithUser(user do.User) *UserUpdateInfo {
	u.User = user
	return u
}

func (u *UserUpdateInfo) GetUserID() uint32 {
	if u == nil {
		return 0
	}
	if u.User == nil {
		return u.UserID
	}
	return u.GetID()
}

func (u *UserUpdateInfo) GetNickname() string {
	if u == nil {
		return ""
	}
	return u.Nickname
}

func (u *UserUpdateInfo) GetAvatar() string {
	if u == nil {
		return ""
	}
	return u.Avatar
}

func (u *UserUpdateInfo) GetGender() vobj.Gender {
	if u == nil {
		return vobj.GenderUnknown
	}
	return u.Gender
}

type PasswordUpdateInfo struct {
	OldPassword  string
	NewPassword  string
	SendEmailFun SendEmailFun
}

type UpdateUserPasswordInfo struct {
	UserID         uint32
	Password       string
	Salt           string
	OriginPassword string
	SendEmailFun   SendEmailFun
}

type UpdateUserStatusRequest struct {
	UserIds []uint32
	Status  vobj.UserStatus
}

type ResetUserPasswordRequest struct {
	UserId       uint32
	SendEmailFun SendEmailFun
}

type UpdateUserPosition interface {
	GetUser() do.User
	GetPosition() vobj.Role
}

type UpdateUserPositionRequest struct {
	operator do.User
	user     do.User
	UserId   uint32
	Position vobj.Role
}

func (r *UpdateUserPositionRequest) GetPosition() vobj.Role {
	if r == nil {
		return vobj.RoleUnknown
	}
	return r.Position
}

func (r *UpdateUserPositionRequest) GetUser() do.User {
	if r == nil {
		return nil
	}
	return r.user
}

func (r *UpdateUserPositionRequest) WithOperator(operator do.User) *UpdateUserPositionRequest {
	r.operator = operator
	return r
}

func (r *UpdateUserPositionRequest) WithUser(user do.User) *UpdateUserPositionRequest {
	r.user = user
	return r
}

func (r *UpdateUserPositionRequest) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if validate.IsNil(r.user) {
		return merr.ErrorParamsError("invalid user")
	}
	if r.Position.IsUnknown() {
		return merr.ErrorParamsError("invalid position")
	}
	if r.operator.GetID() == r.user.GetID() {
		return merr.ErrorParamsError("not allowed to update your own position")
	}
	operatorPosition := r.operator.GetPosition()
	if operatorPosition.IsSuperAdmin() {
		return nil
	}
	if !operatorPosition.GT(r.Position) || !operatorPosition.IsAdminOrSuperAdmin() {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}

type UserListRequest struct {
	*PaginationRequest
	Status   []vobj.UserStatus `json:"status"`
	Position []vobj.Role       `json:"position"`
	Keyword  string            `json:"keyword"`
}

func (r *UserListRequest) ToListUserReply(users []*system.User) *UserListReply {
	return &UserListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(users, func(user *system.User) do.User { return user }),
	}
}

type UserListReply = ListReply[do.User]

type UpdateUserRoles interface {
	GetUser() do.User
	GetRoles() []do.Role
}

type UpdateUserRolesReq struct {
	UserID   uint32
	RoleIDs  []uint32
	roles    []do.Role
	operator do.User
	user     do.User
}

func (r *UpdateUserRolesReq) GetUser() do.User {
	if r == nil {
		return nil
	}
	return r.user
}

func (r *UpdateUserRolesReq) GetRoles() []do.Role {
	if r == nil {
		return nil
	}
	return nil
}

func (r *UpdateUserRolesReq) WithRoles(roles []do.Role) UpdateUserRoles {
	r.roles = slices.MapFilter(roles, func(role do.Role) (do.Role, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return nil, false
		}
		return role, true
	})
	return r
}

func (r *UpdateUserRolesReq) WithOperator(operator do.User) *UpdateUserRolesReq {
	r.operator = operator
	return r
}

func (r *UpdateUserRolesReq) WithUser(user do.User) *UpdateUserRolesReq {
	r.user = user
	return r
}

func (r *UpdateUserRolesReq) Validate() error {
	if validate.IsNil(r.operator) {
		return merr.ErrorParamsError("invalid operator")
	}
	if validate.IsNil(r.user) {
		return merr.ErrorParamsError("invalid user")
	}
	operatorPosition := r.operator.GetPosition()
	if operatorPosition.IsSuperAdmin() {
		return nil
	}
	if !operatorPosition.GT(r.user.GetPosition()) || !operatorPosition.IsAdminOrSuperAdmin() {
		return merr.ErrorParamsError("invalid position")
	}
	return nil
}
