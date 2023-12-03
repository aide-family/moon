package bo

import (
	"encoding"
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ encoding.BinaryMarshaler = (*EndpointBO)(nil)
var _ encoding.BinaryUnmarshaler = (*EndpointBO)(nil)

type (
	EndpointBO struct {
		Id            uint            `json:"id"`
		Uuid          string          `json:"uuid"`
		Name          string          `json:"name"`
		Endpoint      string          `json:"endpoint"`
		Status        valueobj.Status `json:"status"`
		Remark        string          `json:"remark"`
		CreatedAt     int64           `json:"createdAt"`
		UpdatedAt     int64           `json:"updatedAt"`
		DeletedAt     int64           `json:"deletedAt"`
		AgentEndpoint string          `json:"agentEndpoint"`
		AgentCheck    string          `json:"agentCheck"`
	}
)

func (l *EndpointBO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

func (l *EndpointBO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *EndpointBO) ToApiSelectV1() *api.PrometheusServer {
	if l == nil {
		return nil
	}
	return &api.PrometheusServer{
		Id:            l.Uuid,
		Name:          l.Name,
		Endpoint:      l.Endpoint,
		Status:        api.Status(l.Status),
		Remark:        l.Remark,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
		AgentEndpoint: l.AgentEndpoint,
		AgentCheck:    l.AgentCheck,
	}
}
