package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertpage"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type AlertPageService struct {
	v1.UnimplementedAlertPageServiceServer
	pages *alertpage.Usecase
}

func NewAlertPageService(uc *alertpage.Usecase) *AlertPageService {
	return &AlertPageService{pages: uc}
}

func (s *AlertPageService) CreateAlertPage(ctx context.Context, req *v1.CreateAlertPageRequest) (*v1.CreateAlertPageResponse, error) {
	item, err := s.pages.Create(ctx, bo.AlertPage{
		Name: req.GetName(), Labels: req.GetLabels(), Public: req.GetPublic(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertPageResponse{Page: alertPageItem(*item)}, nil
}

func (s *AlertPageService) UpdateAlertPage(ctx context.Context, req *v1.UpdateAlertPageRequest) (*v1.UpdateAlertPageResponse, error) {
	item, err := s.pages.Update(ctx, bo.AlertPage{
		UID: req.GetUid(), Name: req.GetName(), Labels: req.GetLabels(), Public: req.GetPublic(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertPageResponse{Page: alertPageItem(*item)}, nil
}

func (s *AlertPageService) DeleteAlertPage(ctx context.Context, req *v1.DeleteAlertPageRequest) (*v1.DeleteAlertPageResponse, error) {
	if err := s.pages.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertPageResponse{}, nil
}

func (s *AlertPageService) GetAlertPage(ctx context.Context, req *v1.GetAlertPageRequest) (*v1.GetAlertPageResponse, error) {
	item, err := s.pages.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertPageResponse{Page: alertPageItem(*item)}, nil
}

func (s *AlertPageService) ListAlertPages(ctx context.Context, req *v1.ListAlertPagesRequest) (*v1.ListAlertPagesResponse, error) {
	items, total, page, pageSize, err := s.pages.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetScope())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertPagesResponse{
		Items: alertPageItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *AlertPageService) ListMyRealtimeTabs(ctx context.Context, _ *v1.ListMyRealtimeTabsRequest) (*v1.ListMyRealtimeTabsResponse, error) {
	tabs, err := s.pages.ListMyTabs(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListMyRealtimeTabsResponse{Tabs: realtimeTabs(tabs)}, nil
}

func (s *AlertPageService) SetMyRealtimeTabs(ctx context.Context, req *v1.SetMyRealtimeTabsRequest) (*v1.SetMyRealtimeTabsResponse, error) {
	tabs, err := s.pages.SetMyTabs(ctx, req.GetAlertPageUids())
	if err != nil {
		return nil, err
	}
	return &v1.SetMyRealtimeTabsResponse{Tabs: realtimeTabs(tabs)}, nil
}

func alertPageItem(in bo.AlertPage) *v1.AlertPageItem {
	return &v1.AlertPageItem{
		Uid: in.UID, Name: in.Name, OwnerUid: in.OwnerUID,
		Labels: in.Labels, Public: in.Public, Remark: in.Remark,
	}
}

func alertPageItems(items []bo.AlertPage) []*v1.AlertPageItem {
	out := make([]*v1.AlertPageItem, 0, len(items))
	for _, item := range items {
		out = append(out, alertPageItem(item))
	}
	return out
}

func realtimeTabs(items []bo.RealtimeAlertTab) []*v1.RealtimeAlertTab {
	out := make([]*v1.RealtimeAlertTab, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.RealtimeAlertTab{Id: item.ID, Name: item.Name, Kind: item.Kind})
	}
	return out
}
