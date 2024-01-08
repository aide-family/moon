package bo

import (
	query "github.com/aide-cloud/gorm-normalize"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	ApiBO struct {
		Id     uint32    `json:"id"`
		Name   string    `json:"name"`
		Path   string    `json:"path"`
		Method string    `json:"method"`
		Status vo.Status `json:"status"`
		Remark string    `json:"remark"`
		Module vo.Module `json:"module"`
		Domain vo.Domain `json:"domain"`

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
		Value:  b.Id,
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
		Id:        b.Id,
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
func (b *ApiBO) ToModel() *do.SysAPI {
	if b == nil {
		return nil
	}

	return &do.SysAPI{
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
func ApiModelToBO(m *do.SysAPI) *ApiBO {
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
