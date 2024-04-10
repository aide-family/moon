package bo

import (
	"encoding"
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/strategy"
)

var _ encoding.BinaryMarshaler = (*EndpointBO)(nil)
var _ encoding.BinaryUnmarshaler = (*EndpointBO)(nil)

type (
	ListEndpointReq struct {
		Page    Pagination
		Keyword string
		Status  vobj.Status
	}
	CreateEndpointReq struct {
		Name     string `json:"name"`
		Endpoint string `json:"endpoint"`
		Remark   string `json:"remark"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	UpdateEndpointReq struct {
		Id uint32 `json:"id"`
		*CreateEndpointReq
	}
	EndpointBO struct {
		Id        uint32              `json:"id"`
		Name      string              `json:"name"`
		Endpoint  string              `json:"endpoint"`
		Status    vobj.Status         `json:"status"`
		Remark    string              `json:"remark"`
		CreatedAt int64               `json:"createdAt"`
		UpdatedAt int64               `json:"updatedAt"`
		DeletedAt int64               `json:"deletedAt"`
		BasicAuth *strategy.BasicAuth `json:"basicAuth"`
	}
)

// GetBasicAuth get basic auth
func (l *EndpointBO) GetBasicAuth() *strategy.BasicAuth {
	if l == nil {
		return nil
	}
	return l.BasicAuth
}

// String json string
func (l *EndpointBO) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

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

// ToModel EndpointBO to do.PromEndpoint
func (l *EndpointBO) ToModel() *do.Endpoint {
	if l == nil {
		return nil
	}
	return &do.Endpoint{
		BaseModel: do.BaseModel{ID: l.Id},
		Name:      l.Name,
		Endpoint:  l.Endpoint,
		Remark:    l.Remark,
		Status:    l.Status,
		BasicAuth: l.GetBasicAuth(),
	}
}

// EndpointModelToBO do.PromEndpoint to EndpointBO
func EndpointModelToBO(m *do.Endpoint) *EndpointBO {
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
		BasicAuth: m.BasicAuth,
	}
}
