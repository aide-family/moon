package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	valueobj2 "prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	UserBO struct {
		Id        uint             `json:"id"`
		Username  string           `json:"username"`
		Nickname  string           `json:"nickname"`
		Password  string           `json:"password"`
		Salt      string           `json:"salt"`
		Email     string           `json:"email"`
		Phone     string           `json:"phone"`
		Status    valueobj2.Status `json:"status"`
		Remark    string           `json:"remark"`
		Avatar    string           `json:"avatar"`
		CreatedAt int64            `json:"createdAt"`
		UpdatedAt int64            `json:"updatedAt"`
		DeletedAt int64            `json:"deletedAt"`
		Roles     []*RoleBO        `json:"roles"`
		Gender    valueobj2.Gender `json:"gender"`
	}
)

func (l *UserBO) ToApiSelectV1() *api.UserSelectV1 {
	if l == nil {
		return nil
	}

	return &api.UserSelectV1{
		Value:    uint32(l.Id),
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
		Id:        uint32(l.Id),
		Username:  l.Username,
		Email:     l.Email,
		Phone:     l.Phone,
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		Avatar:    l.Avatar,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
		Roles: slices.To(l.Roles, func(bo *RoleBO) *api.RoleSelectV1 {
			if bo == nil {
				return nil
			}
			return bo.ApiRoleSelectV1()
		}),
		Nickname: l.Nickname,
		Gender:   l.Gender.Value(),
	}
}

func (l *UserBO) ToModel() *model.SysUser {
	if l == nil {
		return nil
	}

	return &model.SysUser{
		BaseModel: query.BaseModel{
			ID: l.Id,
		},
		Username: l.Username,
		Nickname: l.Nickname,
		Password: l.Password,
		Email:    l.Email,
		Salt:     l.Salt,
		Phone:    l.Phone,
		Status:   l.Status.Value(),
		Remark:   l.Remark,
		Avatar:   l.Avatar,
		Gender:   l.Gender.Value(),
		Roles: slices.To(l.Roles, func(bo *RoleBO) *model.SysRole {
			return bo.ToModel()
		}),
	}
}

// UserModelToBO .
func UserModelToBO(m *model.SysUser) *UserBO {
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
		Status:    valueobj2.Status(m.Status),
		Remark:    m.Remark,
		Salt:      m.Salt,
		Avatar:    m.Avatar,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		Roles: slices.To(m.Roles, func(m *model.SysRole) *RoleBO {
			return RoleModelToBO(m)
		}),
		Gender: valueobj2.Gender(m.Gender),
	}
}
