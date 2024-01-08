package bo

import (
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	RoleBO struct {
		Id        uint32    `json:"id"`
		Name      string    `json:"name"`
		Status    vo.Status `json:"status"`
		Remark    string    `json:"remark"`
		CreatedAt int64     `json:"createdAt"`
		UpdatedAt int64     `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
		Users     []*UserBO `json:"users"`
		Apis      []*ApiBO  `json:"apis"`
	}
)

// GetUsers 获取用户列表
func (l *RoleBO) GetUsers() []*UserBO {
	if l == nil {
		return nil
	}
	return l.Users
}

// GetApis 获取api列表
func (l *RoleBO) GetApis() []*ApiBO {
	if l == nil {
		return nil
	}
	return l.Apis
}

func (l *RoleBO) ApiRoleSelectV1() *api.RoleSelectV1 {
	if l == nil {
		return nil
	}
	return &api.RoleSelectV1{
		Value:  l.Id,
		Label:  l.Name,
		Status: l.Status.Value(),
		Remark: l.Remark,
	}
}

func (l *RoleBO) ToApiV1() *api.RoleV1 {
	if l == nil {
		return nil
	}
	return &api.RoleV1{
		Id:        l.Id,
		Name:      l.Name,
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
		Users: slices.To(l.GetUsers(), func(i *UserBO) *api.UserSelectV1 {
			return i.ToApiSelectV1()
		}),
		Apis: slices.To(l.GetApis(), func(i *ApiBO) *api.ApiSelectV1 {
			return i.ToApiSelectV1()
		}),
	}
}

func (l *RoleBO) ToModel() *do.SysRole {
	if l == nil {
		return nil
	}
	return &do.SysRole{
		BaseModel: do.BaseModel{
			ID: l.Id,
		},
		Remark: l.Remark,
		Name:   l.Name,
		Status: l.Status,
		Users: slices.To(l.GetUsers(), func(i *UserBO) *do.SysUser {
			return i.ToModel()
		}),
		Apis: slices.To(l.GetApis(), func(i *ApiBO) *do.SysAPI {
			return i.ToModel()
		}),
	}
}

// RoleModelToBO .
func RoleModelToBO(m *do.SysRole) *RoleBO {
	if m == nil {
		return nil
	}
	return &RoleBO{
		Id:        m.ID,
		Name:      m.Name,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		Users: slices.To(m.GetUsers(), func(i *do.SysUser) *UserBO {
			return UserModelToBO(i)
		}),
		Apis: slices.To(m.GetApis(), func(i *do.SysAPI) *ApiBO {
			return ApiModelToBO(i)
		}),
	}
}
