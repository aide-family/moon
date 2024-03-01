package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	DictBO struct {
		Id        uint32      `json:"id"`
		Name      string      `json:"name"`
		Category  vo.Category `json:"category"`
		Status    vo.Status   `json:"status"`
		Remark    string      `json:"remark"`
		Color     string      `json:"color"`
		CreatedAt int64       `json:"createdAt"`
		UpdatedAt int64       `json:"updatedAt"`
		DeletedAt int64       `json:"deletedAt"`
	}
)

// String json string
func (d *DictBO) String() string {
	if d == nil {
		return "{}"
	}
	marshal, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

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

func (d *DictBO) ToModel() *do.PromDict {
	if d == nil {
		return nil
	}
	return &do.PromDict{
		Name:     d.Name,
		Category: d.Category,
		Status:   d.Status,
		Remark:   d.Remark,
		Color:    d.Color,
		BaseModel: do.BaseModel{
			ID: d.Id,
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
func DictModelToBO(m *do.PromDict) *DictBO {
	if m == nil {
		return nil
	}
	return &DictBO{
		Id:        m.ID,
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
