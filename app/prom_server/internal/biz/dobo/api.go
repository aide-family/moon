package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper/model"
)

type (
	ApiDO struct {
		Id        uint      `json:"id"`
		Name      string    `json:"name"`
		Path      string    `json:"path"`
		Method    string    `json:"method"`
		Status    int32     `json:"status"`
		Remark    string    `json:"remark"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
	}

	ApiBO struct {
		Id        uint            `json:"id"`
		Name      string          `json:"name"`
		Path      string          `json:"path"`
		Method    string          `json:"method"`
		Status    valueobj.Status `json:"status"`
		Remark    string          `json:"remark"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
	}
)

// NewApiBO .
func NewApiBO(values ...*ApiBO) IBO[*ApiBO, *ApiDO] {
	return NewBO[*ApiBO, *ApiDO](
		BOWithValues[*ApiBO, *ApiDO](values...),
		BOWithDToB[*ApiBO, *ApiDO](apiDoToBo),
		BOWithBToD[*ApiBO, *ApiDO](apiBoToDo),
	)
}

// NewApiDO .
func NewApiDO(values ...*ApiDO) IDO[*ApiBO, *ApiDO] {
	return NewDO[*ApiBO, *ApiDO](
		DOWithValues[*ApiBO, *ApiDO](values...),
		DOWithBToD[*ApiBO, *ApiDO](apiBoToDo),
		DOWithDToB[*ApiBO, *ApiDO](apiDoToBo),
	)
}

// apiBoToDo .
func apiBoToDo(b *ApiBO) *ApiDO {
	if b == nil {
		return nil
	}
	return &ApiDO{
		Id:        b.Id,
		Name:      b.Name,
		Path:      b.Path,
		Method:    b.Method,
		Status:    int32(b.Status),
		Remark:    b.Remark,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
	}
}

func apiDoToBo(d *ApiDO) *ApiBO {
	if d == nil {
		return nil
	}
	return &ApiBO{
		Id:        d.Id,
		Name:      d.Name,
		Path:      d.Path,
		Method:    d.Method,
		Status:    valueobj.Status(d.Status),
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

// ToModel .
func (b *ApiDO) ToModel() *model.SysApi {
	if b == nil {
		return nil
	}

	return &model.SysApi{
		BaseModel: query.BaseModel{
			ID: b.Id,
		},
		Name:   b.Name,
		Path:   b.Path,
		Method: b.Method,
		Status: b.Status,
		Remark: b.Remark,
	}
}

// ToSelectV1 .
func (b *ApiBO) ToSelectV1() *api.ApiSelectV1 {
	if b == nil {
		return nil
	}

	return &api.ApiSelectV1{
		Value:  uint32(b.Id),
		Label:  b.Name,
		Status: api.Status(b.Status),
		Remark: b.Remark,
	}
}

// ToV1 .
func (b *ApiBO) ToV1() *api.ApiV1 {
	if b == nil {
		return nil
	}

	return &api.ApiV1{
		Id:        uint32(b.Id),
		Name:      b.Name,
		Path:      b.Path,
		Method:    b.Method,
		Status:    api.Status(b.Status),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: b.DeletedAt,
	}
}

// ApiModelToDO .
func ApiModelToDO(m *model.SysApi) *ApiDO {
	if m == nil {
		return nil
	}

	return &ApiDO{
		Id:        m.ID,
		Name:      m.Name,
		Path:      m.Path,
		Method:    m.Method,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}
