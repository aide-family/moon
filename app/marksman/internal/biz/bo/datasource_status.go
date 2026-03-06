package bo

import (
	"time"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
	"github.com/bwmarrin/snowflake"
)

// DatasourceStatusSeriesBo is one series of status points for a datasource (from main TSDB query).
type DatasourceStatusSeriesBo struct {
	UID    int64
	Name   string
	Points []DatasourceStatusPointBo
}

// DatasourceStatusPointBo is a single (timestamp, value) point; value 1 = healthy, 0 = unhealthy.
type DatasourceStatusPointBo struct {
	Timestamp int64
	Value     float64
}

// ToAPIV1GetDatasourceStatusReply converts BO to proto reply.
func ToAPIV1GetDatasourceStatusReply(series []*DatasourceStatusSeriesBo) *apiv1.GetDatasourceStatusReply {
	if series == nil {
		return &apiv1.GetDatasourceStatusReply{}
	}
	out := make([]*apiv1.DatasourceStatusSeries, 0, len(series))
	for _, s := range series {
		points := make([]*apiv1.DatasourceStatusPoint, 0, len(s.Points))
		for _, p := range s.Points {
			points = append(points, &apiv1.DatasourceStatusPoint{
				Timestamp: p.Timestamp,
				Value:     p.Value,
			})
		}
		out = append(out, &apiv1.DatasourceStatusSeries{
			Uid:    s.UID,
			Name:   s.Name,
			Points: points,
		})
	}
	return &apiv1.GetDatasourceStatusReply{Series: out}
}

// GetDatasourceStatusRequest is the request for GetDatasourceStatus.
type GetDatasourceStatusRequest struct {
	UID       snowflake.ID
	Name      string
	StartTime int64
	EndTime   int64
	Step      time.Duration
}

func (r *GetDatasourceStatusRequest) GetUID() int64 {
	return r.UID.Int64()
}

func (r *GetDatasourceStatusRequest) GetName() string {
	return r.Name
}

func (r *GetDatasourceStatusRequest) GetStartTime() int64 {
	if r.StartTime <= 0 {
		return time.Now().Unix() - 3600
	}
	return r.StartTime
}

func (r *GetDatasourceStatusRequest) GetEndTime() int64 {
	if r.EndTime <= 0 {
		return time.Now().Unix()
	}
	return r.EndTime
}

func (r *GetDatasourceStatusRequest) GetStep() time.Duration {
	if r.Step <= 0 {
		return 60 * time.Second
	}
	if r.Step > 3600*time.Second {
		return 3600 * time.Second
	}
	return r.Step
}

// NewGetDatasourceStatusRequest converts proto request to BO.
func NewGetDatasourceStatusRequest(req *apiv1.GetDatasourceStatusRequest, datasource *DatasourceItemBo) *GetDatasourceStatusRequest {
	return &GetDatasourceStatusRequest{
		UID:       snowflake.ParseInt64(req.GetUid()),
		Name:      datasource.Name,
		StartTime: req.GetStartTime(),
		EndTime:   req.GetEndTime(),
		Step:      time.Duration(req.GetStepSeconds()) * time.Second,
	}
}
