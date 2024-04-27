package promservice

import (
	"context"

	"github.com/aide-cloud/universal/base/slices"
	"github.com/aide-family/moon/api"
	pb "github.com/aide-family/moon/api/server/prom/endpoint"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/go-kratos/kratos/v2/log"
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
	createEndpointBo := &bo.CreateEndpointReq{
		Name:               req.GetName(),
		Endpoint:           req.GetEndpoint(),
		Remark:             req.GetRemark(),
		Username:           req.GetUsername(),
		Password:           req.GetPassword(),
		DatasourceCategory: agent.DatasourceCategory(req.GetDatasourceType()),
	}

	endpointBo, err := s.endpointBiz.AppendEndpoint(ctx, createEndpointBo)
	if err != nil {
		return nil, err
	}
	return &pb.AppendEndpointReply{
		Id: endpointBo.Id,
	}, nil
}

// DeleteEndpoint 删除
func (s *EndpointService) DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest) (*pb.DeleteEndpointReply, error) {
	if err := s.endpointBiz.DeleteEndpointById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteEndpointReply{
		Id: req.GetId(),
	}, nil
}

// EditEndpoint 编辑
func (s *EndpointService) EditEndpoint(ctx context.Context, req *pb.EditEndpointRequest) (*pb.EditEndpointReply, error) {
	editEndpointBo := &bo.UpdateEndpointReq{
		Id: req.GetId(),
		CreateEndpointReq: &bo.CreateEndpointReq{
			Name:               req.GetName(),
			Endpoint:           req.GetEndpoint(),
			Remark:             req.GetRemark(),
			Username:           req.GetUsername(),
			Password:           req.GetPassword(),
			DatasourceCategory: agent.DatasourceCategory(req.GetDatasourceType()),
		},
	}
	endpointBo, err := s.endpointBiz.UpdateEndpointById(ctx, editEndpointBo)
	if err != nil {
		return nil, err
	}
	return &pb.EditEndpointReply{
		Id: endpointBo.Id,
	}, nil
}

// BatchEditEndpointStatus 批量编辑状态
func (s *EndpointService) BatchEditEndpointStatus(ctx context.Context, req *pb.BatchEditEndpointStatusRequest) (*pb.BatchEditEndpointStatusReply, error) {
	if err := s.endpointBiz.UpdateStatusByIds(ctx, req.GetIds(), vobj.Status(req.GetStatus())); err != nil {
		return nil, err
	}
	return &pb.BatchEditEndpointStatusReply{
		Ids: req.GetIds(),
	}, nil
}

// GetEndpoint 详情
func (s *EndpointService) GetEndpoint(ctx context.Context, req *pb.GetEndpointRequest) (*pb.GetEndpointReply, error) {
	detail, err := s.endpointBiz.DetailById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetEndpointReply{
		Detail: detail.ToApiV1(),
	}, nil
}

// ListEndpoint 查询
func (s *EndpointService) ListEndpoint(ctx context.Context, req *pb.ListEndpointRequest) (*pb.ListEndpointReply, error) {
	pgReq := req.GetPage()
	pageInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	listEndpoint, err := s.endpointBiz.ListEndpoint(ctx, &bo.ListEndpointReq{
		Page:    pageInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
		DatasourceCategoryList: slices.To(req.GetDatasourceTypes(), func(i api.DatasourceType) agent.DatasourceCategory {
			return agent.DatasourceCategory(i)
		}),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*api.PrometheusServerItem, 0, len(listEndpoint))
	for _, endpoint := range listEndpoint {
		list = append(list, endpoint.ToApiV1())
	}
	return &pb.ListEndpointReply{
		Page: &api.PageReply{
			Curr:  pageInfo.GetRespCurr(),
			Size:  pageInfo.GetSize(),
			Total: pageInfo.GetTotal(),
		},
		List: list,
	}, nil
}

// SelectEndpoint 查询
func (s *EndpointService) SelectEndpoint(ctx context.Context, req *pb.SelectEndpointRequest) (*pb.SelectEndpointReply, error) {
	pgReq := req.GetPage()
	pageInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	listEndpoint, err := s.endpointBiz.ListEndpoint(ctx, &bo.ListEndpointReq{
		Page:    pageInfo,
		Keyword: req.GetKeyword(),
		Status:  vobj.Status(req.GetStatus()),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*api.PrometheusServerSelectItem, 0, len(listEndpoint))
	for _, endpoint := range listEndpoint {
		list = append(list, endpoint.ToApiSelectV1())
	}
	return &pb.SelectEndpointReply{
		Page: &api.PageReply{
			Curr:  pageInfo.GetRespCurr(),
			Size:  pageInfo.GetSize(),
			Total: pageInfo.GetTotal(),
		},
		List: list,
	}, nil
}
