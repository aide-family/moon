package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/endpoint"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type EndpointService struct {
	pb.UnimplementedEndpointServer
	log *log.Helper

	endpointBiz *biz.EndpointBiz
}

func NewEndpointService(endpointBiz *biz.EndpointBiz, logger log.Logger) *EndpointService {
	return &EndpointService{
		log:         log.NewHelper(log.With(logger, "module", "service.endpoint")),
		endpointBiz: endpointBiz,
	}
}

func (s *EndpointService) AppendEndpoint(ctx context.Context, req *pb.AppendEndpointRequest) (*pb.AppendEndpointReply, error) {
	endpointBo := make([]*dobo.EndpointBO, 0, len(req.Endpoints))

	if err := s.endpointBiz.AppendEndpoint(ctx, endpointBo); err != nil {
		return nil, err
	}
	return &pb.AppendEndpointReply{}, nil
}

func (s *EndpointService) DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest) (*pb.DeleteEndpointReply, error) {
	return &pb.DeleteEndpointReply{}, nil
}

func (s *EndpointService) ListEndpoint(ctx context.Context, req *pb.ListEndpointRequest) (*pb.ListEndpointReply, error) {
	return &pb.ListEndpointReply{}, nil
}
