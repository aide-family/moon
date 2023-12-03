package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/endpoint"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/valueobj"
)

type EndpointService struct {
	pb.UnimplementedEndpointServer
	log *log.Helper

	endpointBiz *biz.EndpointBiz
}

// NewEndpointService 实例化EndpointService
func NewEndpointService(endpointBiz *biz.EndpointBiz, logger log.Logger) *EndpointService {
	return &EndpointService{
		log:         log.NewHelper(log.With(logger, "module", "service.prom.endpoint")),
		endpointBiz: endpointBiz,
	}
}

// AppendEndpoint 新增
func (s *EndpointService) AppendEndpoint(ctx context.Context, req *pb.AppendEndpointRequest) (*pb.AppendEndpointReply, error) {
	endpointBo := make([]*bo.EndpointBO, 0, len(req.Endpoints))
	for _, endpoint := range req.GetEndpoints() {
		endpointBo = append(endpointBo, &bo.EndpointBO{
			Uuid:          endpoint.GetId(),
			Name:          endpoint.GetName(),
			Endpoint:      endpoint.GetEndpoint(),
			Status:        valueobj.Status(endpoint.GetStatus()),
			Remark:        endpoint.GetRemark(),
			CreatedAt:     endpoint.GetCreatedAt(),
			UpdatedAt:     endpoint.GetUpdatedAt(),
			AgentEndpoint: endpoint.GetAgentEndpoint(),
			AgentCheck:    endpoint.GetAgentCheck(),
		})
	}

	if err := s.endpointBiz.AppendEndpoint(ctx, endpointBo); err != nil {
		return nil, err
	}
	return &pb.AppendEndpointReply{
		Response: &api.Response{
			Code: 0,
			Msg:  "append endpoint's is OK",
		},
	}, nil
}

// DeleteEndpoint 删除
func (s *EndpointService) DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest) (*pb.DeleteEndpointReply, error) {
	endpointList := make([]*bo.EndpointBO, 0, len(req.GetUuids()))
	for _, uuid := range req.GetUuids() {
		endpointList = append(endpointList, &bo.EndpointBO{
			Uuid:          uuid,
			AgentEndpoint: req.GetAgentName(),
		})
	}
	if err := s.endpointBiz.DeleteEndpoint(ctx, endpointList); err != nil {
		return nil, err
	}
	return &pb.DeleteEndpointReply{
		Response: &api.Response{
			Code: 0,
			Msg:  "delete endpoint's is OK",
		},
	}, nil
}

// ListEndpoint 查询
func (s *EndpointService) ListEndpoint(ctx context.Context, req *pb.ListEndpointRequest) (*pb.ListEndpointReply, error) {
	listEndpoint, err := s.endpointBiz.ListEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*api.PrometheusServer, 0, len(listEndpoint))
	for _, endpoint := range listEndpoint {
		list = append(list, endpoint.ToApiSelectV1())
	}
	return &pb.ListEndpointReply{
		Response: &api.Response{
			Code: 0,
			Msg:  "query endpoint's is OK",
		},
		Endpoints: list,
	}, nil
}
