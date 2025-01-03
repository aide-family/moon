package template

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/template"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
)

// SendTemplateService send template service
type SendTemplateService struct {
	pb.UnimplementedSendTemplateServer

	sendTemplate *biz.SendTemplateBiz
}

// NewSendTemplateService new a send template service
func NewSendTemplateService(sendTemplate *biz.SendTemplateBiz) *SendTemplateService {
	return &SendTemplateService{sendTemplate: sendTemplate}
}

// CreateSendTemplate create send template
func (s *SendTemplateService) CreateSendTemplate(ctx context.Context, req *pb.CreateSendTemplateRequest) (*pb.CreateSendTemplateReply, error) {
	param := builder.NewParamsBuild(ctx).SendTemplateModuleBuild().WithSendTemplateCreateRequest(req).ToBo()
	err := s.sendTemplate.CreateSendTemplate(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.CreateSendTemplateReply{}, nil
}

// UpdateSendTemplate update send template
func (s *SendTemplateService) UpdateSendTemplate(ctx context.Context, req *pb.UpdateSendTemplateRequest) (*pb.UpdateSendTemplateReply, error) {
	param := builder.NewParamsBuild(ctx).SendTemplateModuleBuild().WithSendTemplateUpdateRequest(req).ToBo()
	err := s.sendTemplate.UpdateSendTemplate(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateSendTemplateReply{}, nil
}

// DeleteSendTemplate delete send template
func (s *SendTemplateService) DeleteSendTemplate(ctx context.Context, req *pb.DeleteSendTemplateRequest) (*pb.DeleteSendTemplateReply, error) {
	err := s.sendTemplate.DeleteSendTemplate(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.DeleteSendTemplateReply{}, nil
}

// GetSendTemplate get send template
func (s *SendTemplateService) GetSendTemplate(ctx context.Context, req *pb.GetSendTemplateRequest) (*pb.GetSendTemplateReply, error) {
	detail, err := s.sendTemplate.GetSendTemplateDetail(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.GetSendTemplateReply{
		Detail: builder.NewParamsBuild(ctx).SendTemplateModuleBuild().IDoSendTemplateBuilder().ToAPI(detail),
	}, nil
}

// ListSendTemplate list send template
func (s *SendTemplateService) ListSendTemplate(ctx context.Context, req *pb.ListSendTemplateRequest) (*pb.ListSendTemplateReply, error) {
	param := builder.NewParamsBuild(ctx).SendTemplateModuleBuild().WithSendTemplateListRequest(req).ToBo()
	list, err := s.sendTemplate.SendTemplateList(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.ListSendTemplateReply{
		List:       builder.NewParamsBuild(ctx).SendTemplateModuleBuild().IDoSendTemplateBuilder().ToAPIs(list),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(param.Page),
	}, nil
}

// UpdateStatus update send template status
func (s *SendTemplateService) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.UpdateStatusReply, error) {
	param := builder.NewParamsBuild(ctx).SendTemplateModuleBuild().WithSendTemplateStatusUpdateRequest(req).ToBo()
	err := s.sendTemplate.UpdateSendTemplateStatus(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateStatusReply{}, nil
}
