package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/plugin/soft_delete"
	"prometheus-manager/api"
	"prometheus-manager/pkg/model"
)

type (
	DictBO struct {
		Id        uint32
		Name      string
		Category  api.Category
		Status    api.Status
		Remark    string
		Color     string
		CreatedAt int64
		UpdatedAt int64
		DeletedAt int64
	}

	DictDO struct {
		Id        uint
		Name      string
		Category  int32
		Status    int32
		Remark    string
		Color     string
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt int64
	}
)

// NewDictBO 创建字典业务对象
func NewDictBO(values ...*DictBO) IBO[*DictBO, *DictDO] {
	return NewBO[*DictBO, *DictDO](
		BOWithValues[*DictBO, *DictDO](values...),
		BOWithDToB[*DictBO, *DictDO](dictDoToBo),
		BOWithBToD[*DictBO, *DictDO](dictBoToDo),
	)
}

// NewDictDO 创建字典数据对象
func NewDictDO(values ...*DictDO) IDO[*DictBO, *DictDO] {
	return NewDO[*DictBO, *DictDO](
		DOWithValues[*DictBO, *DictDO](values...),
		DOWithBToD[*DictBO, *DictDO](dictBoToDo),
		DOWithDToB[*DictBO, *DictDO](dictDoToBo),
	)
}

// dictDoToBo 字典数据对象转换为字典业务对象
func dictDoToBo(d *DictDO) *DictBO {
	if d == nil {
		return nil
	}
	return &DictBO{
		Id:        uint32(d.Id),
		Name:      d.Name,
		Category:  api.Category(d.Category),
		Status:    api.Status(d.Status),
		Remark:    d.Remark,
		Color:     d.Color,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

// dictBoToDo 字典业务对象转换为字典数据对象
func dictBoToDo(b *DictBO) *DictDO {
	if b == nil {
		return nil
	}
	return &DictDO{
		Id:        uint(b.Id),
		Name:      b.Name,
		Category:  int32(b.Category),
		Status:    int32(b.Status),
		Remark:    b.Remark,
		Color:     b.Color,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
	}
}

// ToApiDictSelectV1 转换为api字典查询对象
func (b *DictBO) ToApiDictSelectV1() *api.DictSelectV1 {
	return &api.DictSelectV1{
		Value:     b.Id,
		Label:     b.Name,
		Category:  b.Category,
		Color:     b.Color,
		Status:    b.Status,
		Remark:    b.Remark,
		IsDeleted: b.DeletedAt > 0,
	}
}

func ListToApiDictSelectV1(values ...*DictBO) []*api.DictSelectV1 {
	list := make([]*api.DictSelectV1, 0, len(values))
	for _, v := range values {
		list = append(list, v.ToApiDictSelectV1())
	}
	return list
}

// DictModelToDO dict model to dict do
func DictModelToDO(m *model.PromDict) *DictDO {
	if m == nil {
		return nil
	}
	return &DictDO{
		Id:        m.ID,
		Name:      m.Name,
		Category:  m.Category,
		Status:    m.Status,
		Remark:    m.Remark,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}

// DictDOToModel dict do to dict model
func DictDOToModel(d *DictDO) *model.PromDict {
	if d == nil {
		return nil
	}
	return &model.PromDict{
		Name:     d.Name,
		Category: d.Category,
		Status:   d.Status,
		Remark:   d.Remark,
		Color:    d.Color,
		BaseModel: query.BaseModel{
			ID:        d.Id,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			DeletedAt: soft_delete.DeletedAt(d.DeletedAt),
		},
	}
}
