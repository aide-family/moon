package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	RoleBO struct {
		Id        uint            `json:"id"`
		Name      string          `json:"name"`
		Status    valueobj.Status `json:"status"`
		Remark    string          `json:"remark"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
		Users     []*UserBO       `json:"users"`
	}
)

func (l *RoleBO) ApiRoleSelectV1() *api.RoleSelectV1 {
	if l == nil {
		return nil
	}
	return &api.RoleSelectV1{
		Value:  uint32(l.Id),
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
		Id:        uint32(l.Id),
		Name:      l.Name,
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
		Users: slices.To(l.Users, func(i *UserBO) *api.UserSelectV1 {
			return i.ToApiSelectV1()
		}),
	}
}

func (l *RoleBO) ToModel() *model.SysRole {
	if l == nil {
		return nil
	}
	return &model.SysRole{
		BaseModel: query.BaseModel{
			ID: l.Id,
		},
		Remark: l.Remark,
		Name:   l.Name,
		Status: l.Status.Value(),
		Users: slices.To(l.Users, func(i *UserBO) *model.SysUser {
			return i.ToModel()
		}),
	}
}

// RoleModelToBO .
func RoleModelToBO(m *model.SysRole) *RoleBO {
	if m == nil {
		return nil
	}
	return &RoleBO{
		Id:        m.ID,
		Name:      m.Name,
		Status:    valueobj.Status(m.Status),
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		Users: slices.To(m.Users, func(i *model.SysUser) *UserBO {
			return UserModelToBO(i)
		}),
	}
}
