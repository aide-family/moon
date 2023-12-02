package dobo

import (
	"encoding"
	"encoding/json"
	"time"

	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ encoding.BinaryMarshaler = (*EndpointDO)(nil)
var _ encoding.BinaryUnmarshaler = (*EndpointDO)(nil)

type (
	EndpointDO struct {
		Id            uint      `json:"id"`
		Uuid          string    `json:"uuid"`
		Name          string    `json:"name"`
		Endpoint      string    `json:"endpoint"`
		Status        int32     `json:"status"`
		Remark        string    `json:"remark"`
		CreatedAt     time.Time `json:"createdAt"`
		UpdatedAt     time.Time `json:"updatedAt"`
		DeletedAt     int64     `json:"deletedAt"`
		AgentEndpoint string    `json:"agentEndpoint"`
		AgentCheck    string    `json:"agentCheck"`
	}

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

func (l *EndpointDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *EndpointDO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

// NewEndpointDO new endpoint do
func NewEndpointDO(values ...*EndpointDO) IDO[*EndpointBO, *EndpointDO] {
	return NewDO[*EndpointBO, *EndpointDO](
		DOWithValues[*EndpointBO, *EndpointDO](values...),
		DOWithBToD[*EndpointBO, *EndpointDO](endpointBoToDo),
		DOWithDToB[*EndpointBO, *EndpointDO](endpointDoToBo),
	)
}

// NewEndpointBO new endpoint bo
func NewEndpointBO(values ...*EndpointBO) IBO[*EndpointBO, *EndpointDO] {
	return NewBO[*EndpointBO, *EndpointDO](
		BOWithValues[*EndpointBO, *EndpointDO](values...),
		BOWithBToD[*EndpointBO, *EndpointDO](endpointBoToDo),
		BOWithDToB[*EndpointBO, *EndpointDO](endpointDoToBo),
	)
}

func endpointDoToBo(d *EndpointDO) *EndpointBO {
	if d == nil {
		return nil
	}
	return &EndpointBO{
		Id:            d.Id,
		Uuid:          d.Uuid,
		Name:          d.Name,
		Endpoint:      d.Endpoint,
		Status:        valueobj.Status(d.Status),
		Remark:        d.Remark,
		CreatedAt:     d.CreatedAt.Unix(),
		UpdatedAt:     d.UpdatedAt.Unix(),
		DeletedAt:     d.DeletedAt,
		AgentEndpoint: d.AgentEndpoint,
		AgentCheck:    d.AgentCheck,
	}
}

func endpointBoToDo(b *EndpointBO) *EndpointDO {
	if b == nil {
		return nil
	}
	return &EndpointDO{
		Id:            b.Id,
		Uuid:          b.Uuid,
		Name:          b.Name,
		Endpoint:      b.Endpoint,
		Status:        int32(b.Status),
		Remark:        b.Remark,
		CreatedAt:     time.Unix(b.CreatedAt, 0),
		UpdatedAt:     time.Unix(b.UpdatedAt, 0),
		AgentEndpoint: b.AgentEndpoint,
		AgentCheck:    b.AgentCheck,
	}
}

func (l *EndpointBO) ToApiEndpointSelectV1() *api.PrometheusServer {
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
