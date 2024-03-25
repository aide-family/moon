package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/server/prom/notify"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/pkg/util/slices"
)

type TemplateService struct {
	pb.UnimplementedTemplateServer

	log               *log.Helper
	notifyTemplateBiz *biz.NotifyTemplateBiz
}

func NewTemplateService(notifyTemplateBiz *biz.NotifyTemplateBiz, logger log.Logger) *TemplateService {
	return &TemplateService{
		log:               log.NewHelper(log.With(logger, "module", "server.prom.notify.template")),
		notifyTemplateBiz: notifyTemplateBiz,
	}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateReply, error) {
	createReq := &bo.NotifyTemplateCreateBO{
		Content:    req.GetContent(),
		StrategyID: req.GetStrategyId(),
		NotifyType: vobj.NotifyTemplateType(req.GetNotifyType()),
	}
	_, err := s.notifyTemplateBiz.CreateTemplate(ctx, createReq)
	if err != nil {
		s.log.Warnw("CreateTemplate err", err)
		return nil, err
	}
	return &pb.CreateTemplateReply{}, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.UpdateTemplateReply, error) {
	updateReq := &bo.NotifyTemplateUpdateBo{
		Id:         req.GetId(),
		Content:    req.GetContent(),
		StrategyID: req.GetStrategyId(),
		NotifyType: vobj.NotifyTemplateType(req.GetNotifyType()),
	}
	if err := s.notifyTemplateBiz.UpdateTemplate(ctx, updateReq); err != nil {
		s.log.Warnw("UpdateTemplate err", err)
		return nil, err
	}
	return &pb.UpdateTemplateReply{}, nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*pb.DeleteTemplateReply, error) {
	if err := s.notifyTemplateBiz.DeleteTemplate(ctx, req.GetId()); err != nil {
		s.log.Warnw("DeleteTemplate err", err)
		return nil, err
	}
	return &pb.DeleteTemplateReply{}, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.GetTemplateReply, error) {
	notifyTemplateDetail, err := s.notifyTemplateBiz.GetTemplate(ctx, req.GetId())
	if err != nil {
		s.log.Warnw("GetTemplate err", err)
		return nil, err
	}
	return &pb.GetTemplateReply{
		Detail: notifyTemplateDetail.ToApi(),
	}, nil
}

func (s *TemplateService) ListTemplate(ctx context.Context, req *pb.ListTemplateRequest) (*pb.ListTemplateReply, error) {
	page := bo.NewPage(req.GetPage().GetCurr(), req.GetPage().GetSize())
	listParams := bo.NotifyTemplateListBo{
		Page:       page,
		StrategyId: req.GetStrategyId(),
	}
	notifyTemplates, err := s.notifyTemplateBiz.ListTemplate(ctx, &listParams)
	if err != nil {
		s.log.Warnw("ListTemplate err", err)
		return nil, err
	}
	return &pb.ListTemplateReply{
		Page: &api.PageReply{
			Curr:  page.GetRespCurr(),
			Size:  page.GetSize(),
			Total: page.GetTotal(),
		},
		List: slices.To(notifyTemplates, func(item *bo.NotifyTemplateBO) *api.NotifyTemplateItem {
			return item.ToApi()
		}),
	}, nil
}
