package service

import (
	"context"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"

	pb "github.com/aide-family/moon/api/admin/template"
)

type SendTemplateService struct {
	pb.UnimplementedSendTemplateServer

	sendTemplate *biz.SendTemplateBiz
}

func NewSendTemplateService(sendTemplate *biz.SendTemplateBiz) *SendTemplateService {
	return &SendTemplateService{sendTemplate: sendTemplate}
}

func (s *SendTemplateService) CreateSendTemplate(ctx context.Context, req *pb.CreateSendTemplateRequest) (*pb.CreateSendTemplateReply, error) {
	return &pb.CreateSendTemplateReply{}, nil
}
func (s *SendTemplateService) UpdateSendTemplate(ctx context.Context, req *pb.UpdateSendTemplateRequest) (*pb.UpdateSendTemplateReply, error) {
	return &pb.UpdateSendTemplateReply{}, nil
}
func (s *SendTemplateService) DeleteSendTemplate(ctx context.Context, req *pb.DeleteSendTemplateRequest) (*pb.DeleteSendTemplateReply, error) {
	return &pb.DeleteSendTemplateReply{}, nil
}
func (s *SendTemplateService) GetSendTemplate(ctx context.Context, req *pb.GetSendTemplateRequest) (*pb.GetSendTemplateReply, error) {
	return &pb.GetSendTemplateReply{}, nil
}
func (s *SendTemplateService) ListSendTemplate(ctx context.Context, req *pb.ListSendTemplateRequest) (*pb.ListSendTemplateReply, error) {
	return &pb.ListSendTemplateReply{}, nil
}
func (s *SendTemplateService) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.UpdateStatusReply, error) {
	return &pb.UpdateStatusReply{}, nil
}
