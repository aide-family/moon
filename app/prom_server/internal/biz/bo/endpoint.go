package bo

import (
	"encoding"
	"encoding/json"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ encoding.BinaryMarshaler = (*EndpointBO)(nil)
var _ encoding.BinaryUnmarshaler = (*EndpointBO)(nil)

type (
	EndpointBO struct {
		Id        uint32          `json:"id"`
		Name      string          `json:"name"`
		Endpoint  string          `json:"endpoint"`
		Status    valueobj.Status `json:"status"`
		Remark    string          `json:"remark"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
	}
)

func (l *EndpointBO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

func (l *EndpointBO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *EndpointBO) ToApiV1() *api.PrometheusServerItem {
	if l == nil {
		return nil
	}
	return &api.PrometheusServerItem{
		Id:        l.Id,
		Name:      l.Name,
		Endpoint:  l.Endpoint,
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func (l *EndpointBO) ToApiSelectV1() *api.PrometheusServerSelectItem {
	if l == nil {
		return nil
	}
	return &api.PrometheusServerSelectItem{
		Value:    l.Id,
		Label:    l.Name,
		Status:   l.Status.Value(),
		Remark:   l.Remark,
		Endpoint: l.Endpoint,
	}
}

// ToModel EndpointBO to model.PromEndpoint
func (l *EndpointBO) ToModel() *model.Endpoint {
	if l == nil {
		return nil
	}
	return &model.Endpoint{
		BaseModel: query.BaseModel{ID: l.Id},
		Name:      l.Name,
		Endpoint:  l.Endpoint,
		Remark:    l.Remark,
		Status:    l.Status,
	}
}

// EndpointModelToBO model.PromEndpoint to EndpointBO
func EndpointModelToBO(m *model.Endpoint) *EndpointBO {
	if m == nil {
		return nil
	}
	return &EndpointBO{
		Id:        m.ID,
		Name:      m.Name,
		Endpoint:  m.Endpoint,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
