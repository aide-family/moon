package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	ApiBO struct {
		Id     uint            `json:"id"`
		Name   string          `json:"name"`
		Path   string          `json:"path"`
		Method string          `json:"method"`
		Status valueobj.Status `json:"status"`
		Remark string          `json:"remark"`
		Module valueobj.Module `json:"module"`
		Domain valueobj.Domain `json:"domain"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// ToApiSelectV1 .
func (b *ApiBO) ToApiSelectV1() *api.ApiSelectV1 {
	if b == nil {
		return nil
	}

	return &api.ApiSelectV1{
		Value:  uint32(b.Id),
		Label:  b.Name,
		Status: b.Status.Value(),
		Remark: b.Remark,
		Module: b.Module.Value(),
		Domain: b.Domain.Value(),
	}
}

// ToApiV1 .
func (b *ApiBO) ToApiV1() *api.ApiV1 {
	if b == nil {
		return nil
	}

	return &api.ApiV1{
		Id:        uint32(b.Id),
		Name:      b.Name,
		Path:      b.Path,
		Method:    b.Method,
		Status:    b.Status.Value(),
		Remark:    b.Remark,
		Module:    b.Module.Value(),
		Domain:    b.Domain.Value(),
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: b.DeletedAt,
	}
}

// ToModel .
func (b *ApiBO) ToModel() *model.SysAPI {
	if b == nil {
		return nil
	}

	return &model.SysAPI{
		BaseModel: query.BaseModel{
			ID: b.Id,
		},
		Name:   b.Name,
		Path:   b.Path,
		Method: b.Method,
		Status: b.Status,
		Remark: b.Remark,
		Module: b.Module,
		Domain: b.Domain,
	}
}

// ApiModelToBO .
func ApiModelToBO(m *model.SysAPI) *ApiBO {
	if m == nil {
		return nil
	}

	return &ApiBO{
		Id:        m.ID,
		Name:      m.Name,
		Path:      m.Path,
		Method:    m.Method,
		Status:    m.Status,
		Remark:    m.Remark,
		Module:    m.Module,
		Domain:    m.Domain,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
