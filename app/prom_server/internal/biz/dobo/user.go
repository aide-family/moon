package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

type (
	UserDO struct {
		Id        uint      `json:"id"`
		Username  string    `json:"username"`
		Nickname  string    `json:"nickname"`
		Password  string    `json:"password"`
		Email     string    `json:"email"`
		Phone     string    `json:"phone"`
		Status    int32     `json:"status"`
		Remark    string    `json:"remark"`
		Avatar    string    `json:"avatar"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
		Salt      string    `json:"salt"`
		Roles     []*RoleDO `json:"roles"`
		Gender    int32     `json:"gender"`
	}

	UserBO struct {
		Id        uint            `json:"id"`
		Username  string          `json:"username"`
		Nickname  string          `json:"nickname"`
		Password  string          `json:"password"`
		Email     string          `json:"email"`
		Phone     string          `json:"phone"`
		Status    valueobj.Status `json:"status"`
		Remark    string          `json:"remark"`
		Avatar    string          `json:"avatar"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
		Roles     []*RoleBO       `json:"roles"`
		Gender    valueobj.Gender `json:"gender"`
	}
)

// NewUserDO .
func NewUserDO(values ...*UserDO) IDO[*UserBO, *UserDO] {
	return NewDO[*UserBO, *UserDO](
		DOWithValues[*UserBO, *UserDO](values...),
		DOWithBToD[*UserBO, *UserDO](userBoToDo),
		DOWithDToB[*UserBO, *UserDO](userDoToBo),
	)
}

// NewUserBO .
func NewUserBO(values ...*UserBO) IBO[*UserBO, *UserDO] {
	return NewBO[*UserBO, *UserDO](
		BOWithValues[*UserBO, *UserDO](values...),
		BOWithDToB[*UserBO, *UserDO](userDoToBo),
		BOWithBToD[*UserBO, *UserDO](userBoToDo),
	)
}

func userDoToBo(d *UserDO) *UserBO {
	if d == nil {
		return nil
	}
	return &UserBO{
		Id:        d.Id,
		Username:  d.Username,
		Nickname:  d.Nickname,
		Password:  d.Password,
		Email:     d.Email,
		Phone:     d.Phone,
		Status:    valueobj.Status(d.Status),
		Remark:    d.Remark,
		Avatar:    d.Avatar,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
		Roles:     NewRoleDO(d.Roles...).BO().List(),
		Gender:    valueobj.Gender(d.Gender),
	}
}

func userBoToDo(b *UserBO) *UserDO {
	if b == nil {
		return nil
	}
	return &UserDO{
		Id:        b.Id,
		Username:  b.Username,
		Nickname:  b.Nickname,
		Password:  b.Password,
		Email:     b.Email,
		Phone:     b.Phone,
		Status:    int32(b.Status),
		Remark:    b.Remark,
		Avatar:    b.Avatar,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
		Roles:     NewRoleBO(b.Roles...).DO().List(),
		Gender:    b.Gender.Value(),
	}
}

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

func (l *UserDO) ToModel() *model.SysUser {
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
		Phone:    l.Phone,
		Status:   l.Status,
		Remark:   l.Remark,
		Avatar:   l.Avatar,
		Salt:     l.Salt,
		Gender:   l.Gender,
		Roles: slices.To(l.Roles, func(do *RoleDO) *model.SysRole {
			if do == nil {
				return nil
			}
			return do.ToModel()
		}),
	}
}

// UserModelToDO .
func UserModelToDO(m *model.SysUser) *UserDO {
	if m == nil {
		return nil
	}

	return &UserDO{
		Id:        m.ID,
		Username:  m.Username,
		Nickname:  m.Nickname,
		Password:  m.Password,
		Email:     m.Email,
		Phone:     m.Phone,
		Status:    m.Status,
		Remark:    m.Remark,
		Avatar:    m.Avatar,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
		Salt:      m.Salt,
		Roles: slices.To(m.Roles, func(m *model.SysRole) *RoleDO {
			if m == nil {
				return nil
			}
			return RoleModelToDO(m)
		}),
		Gender: m.Gender,
	}
}
