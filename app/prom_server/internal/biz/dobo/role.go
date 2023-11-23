package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/plugin/soft_delete"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper/model"
)

type (
	RoleDO struct {
		Id        uint      `json:"id"`
		Name      string    `json:"name"`
		Status    int32     `json:"status"`
		Remark    string    `json:"remark"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
	}

	RoleBO struct {
		Id        uint            `json:"id"`
		Name      string          `json:"name"`
		Status    valueobj.Status `json:"status"`
		Remark    string          `json:"remark"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
	}
)

// NewRoleDO .
func NewRoleDO(values ...*RoleDO) IDO[*RoleBO, *RoleDO] {
	return NewDO[*RoleBO, *RoleDO](
		DOWithValues[*RoleBO, *RoleDO](values...),
		DOWithBToD[*RoleBO, *RoleDO](roleBoToDo),
		DOWithDToB[*RoleBO, *RoleDO](roleDoToBo),
	)
}

// NewRoleBO .
func NewRoleBO(values ...*RoleBO) IBO[*RoleBO, *RoleDO] {
	return NewBO[*RoleBO, *RoleDO](
		BOWithValues[*RoleBO, *RoleDO](values...),
		BOWithBToD[*RoleBO, *RoleDO](roleBoToDo),
		BOWithDToB[*RoleBO, *RoleDO](roleDoToBo),
	)
}

func roleBoToDo(b *RoleBO) *RoleDO {
	if b == nil {
		return nil
	}
	return &RoleDO{
		Id:        b.Id,
		Name:      b.Name,
		Status:    int32(b.Status),
		Remark:    b.Remark,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
	}
}

func roleDoToBo(d *RoleDO) *RoleBO {
	if d == nil {
		return nil
	}
	return &RoleBO{
		Id:        d.Id,
		Name:      d.Name,
		Status:    valueobj.Status(d.Status),
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

func (l *RoleBO) ApiRoleSelectV1() *api.RoleSelectV1 {
	return &api.RoleSelectV1{
		Value:  uint32(l.Id),
		Label:  l.Name,
		Status: l.Status.ApiStatus(),
		Remark: l.Remark,
	}
}

func (l *RoleBO) ApiRoleV1() *api.RoleV1 {
	return &api.RoleV1{
		Id:        uint32(l.Id),
		Name:      l.Name,
		Status:    l.Status.ApiStatus(),
		Remark:    l.Remark,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
	}
}

func (l *RoleDO) ModelRole() *model.SysRole {
	if l == nil {
		return nil
	}
	return &model.SysRole{
		BaseModel: query.BaseModel{
			ID:        l.Id,
			CreatedAt: l.CreatedAt,
			UpdatedAt: l.UpdatedAt,
			DeletedAt: soft_delete.DeletedAt(l.DeletedAt),
		},
		Remark: l.Remark,
		Name:   l.Name,
		Status: l.Status,
	}
}

// RoleModelToDO .
func RoleModelToDO(m *model.SysRole) *RoleDO {
	if m == nil {
		return nil
	}
	return &RoleDO{
		Id:        m.ID,
		Name:      m.Name,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}
