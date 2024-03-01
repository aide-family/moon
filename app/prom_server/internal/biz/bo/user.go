package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	UserBO struct {
		Id        uint32    `json:"id"`
		Username  string    `json:"username"`
		Nickname  string    `json:"nickname"`
		Password  string    `json:"password"`
		Salt      string    `json:"salt"`
		Email     string    `json:"email"`
		Phone     string    `json:"phone"`
		Status    vo.Status `json:"status"`
		Remark    string    `json:"remark"`
		Avatar    string    `json:"avatar"`
		CreatedAt int64     `json:"createdAt"`
		UpdatedAt int64     `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
		Roles     []*RoleBO `json:"roles"`
		Gender    vo.Gender `json:"gender"`
	}
)

// String json string
func (l *UserBO) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetRoles 获取角色列表
func (l *UserBO) GetRoles() []*RoleBO {
	if l == nil {
		return nil
	}
	return l.Roles
}

func (l *UserBO) ToApiSelectV1() *api.UserSelectV1 {
	if l == nil {
		return nil
	}

	return &api.UserSelectV1{
		Value:    l.Id,
		Label:    l.Username,
		Status:   l.Status.Value(),
		Avatar:   l.Avatar,
		Nickname: l.Nickname,
		Gender:   l.Gender.Value(),
	}
}

func (l *UserBO) ToApiV1() *api.UserV1 {
	if l == nil {
		return nil
	}

	return &api.UserV1{
		Id:        l.Id,
		Username:  l.Username,
		Email:     l.Email,
		Phone:     l.Phone,
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		Avatar:    l.Avatar,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
		Roles: slices.To(l.GetRoles(), func(bo *RoleBO) *api.RoleSelectV1 {
			return bo.ApiRoleSelectV1()
		}),
		Nickname: l.Nickname,
		Gender:   l.Gender.Value(),
	}
}

func (l *UserBO) ToModel() *do.SysUser {
	if l == nil {
		return nil
	}

	return &do.SysUser{
		BaseModel: do.BaseModel{
			ID: l.Id,
		},
		Username: l.Username,
		Nickname: l.Nickname,
		Password: l.Password,
		Email:    l.Email,
		Salt:     l.Salt,
		Phone:    l.Phone,
		Status:   l.Status,
		Remark:   l.Remark,
		Avatar:   l.Avatar,
		Gender:   l.Gender,
		Roles: slices.To(l.GetRoles(), func(bo *RoleBO) *do.SysRole {
			return bo.ToModel()
		}),
	}
}

// UserModelToBO .
func UserModelToBO(m *do.SysUser) *UserBO {
	if m == nil {
		return nil
	}

	return &UserBO{
		Id:        m.ID,
		Username:  m.Username,
		Nickname:  m.Nickname,
		Password:  m.Password,
		Email:     m.Email,
		Phone:     m.Phone,
		Status:    m.Status,
		Remark:    m.Remark,
		Salt:      m.Salt,
		Avatar:    m.Avatar,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		Roles: slices.To(m.GetRoles(), func(m *do.SysRole) *RoleBO {
			return RoleModelToBO(m)
		}),
		Gender: m.Gender,
	}
}
