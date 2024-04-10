package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type (
	ApiListApiReq struct {
		Keyword string      `json:"keyword"`
		Status  vobj.Status `json:"status"`
		Curr    int32       `json:"curr"`
		Size    int32       `json:"size"`
	}

	ApiBO struct {
		Id     uint32      `json:"id"`
		Name   string      `json:"name"`
		Path   string      `json:"path"`
		Method string      `json:"method"`
		Status vobj.Status `json:"status"`
		Remark string      `json:"remark"`
		Module vobj.Module `json:"module"`
		Domain vobj.Domain `json:"domain"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// String json string
func (b *ApiBO) String() string {
	if b == nil {
		return "{}"
	}
	marshal, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

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
		BaseModel: do.BaseModel{
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
