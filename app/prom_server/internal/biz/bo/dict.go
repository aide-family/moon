package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	DictBO struct {
		Id        uint32            `json:"id"`
		Name      string            `json:"name"`
		Category  valueobj.Category `json:"category"`
		Status    valueobj.Status   `json:"status"`
		Remark    string            `json:"remark"`
		Color     string            `json:"color"`
		CreatedAt int64             `json:"createdAt"`
		UpdatedAt int64             `json:"updatedAt"`
		DeletedAt int64             `json:"deletedAt"`
	}
)

// ToApiSelectV1 转换为api字典查询对象
func (d *DictBO) ToApiSelectV1() *api.DictSelectV1 {
	if d == nil {
		return nil
	}
	return &api.DictSelectV1{
		Value:     d.Id,
		Label:     d.Name,
		Category:  api.Category(d.Category),
		Color:     d.Color,
		Status:    api.Status(d.Status),
		Remark:    d.Remark,
		IsDeleted: d.DeletedAt > 0,
	}
}

func ListToApiDictSelectV1(values ...*DictBO) []*api.DictSelectV1 {
	list := make([]*api.DictSelectV1, 0, len(values))
	for _, v := range values {
		list = append(list, v.ToApiSelectV1())
	}
	return list
}

func (d *DictBO) ToModel() *model.PromDict {
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
			ID: uint(d.Id),
		},
	}
}

// ToApiV1 转换为api字典对象
func (d *DictBO) ToApiV1() *api.DictV1 {
	if d == nil {
		return nil
	}
	return &api.DictV1{
		Id:        d.Id,
		Name:      d.Name,
		Category:  d.Category.Value(),
		Color:     d.Color,
		Status:    d.Status.Value(),
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}

// DictModelToBO 字典model数据对象转换为字典业务对象
func DictModelToBO(m *model.PromDict) *DictBO {
	if m == nil {
		return nil
	}
	return &DictBO{
		Id:        uint32(m.ID),
		Name:      m.Name,
		Category:  m.Category,
		Status:    m.Status,
		Remark:    m.Remark,
		Color:     m.Color,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
