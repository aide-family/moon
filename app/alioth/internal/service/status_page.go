package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/statuspage"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// StatusPageService 把状态页接口转成 biz 入参。
type StatusPageService struct {
	v1.UnimplementedStatusPageServiceServer
	gateway *statuspage.Gateway
}

func NewStatusPageService(gateway *statuspage.Gateway) *StatusPageService {
	return &StatusPageService{gateway: gateway}
}

func (s *StatusPageService) CreateStatusPage(ctx context.Context, req *v1.CreateStatusPageRequest) (*v1.CreateStatusPageResponse, error) {
	item, err := s.gateway.Create(ctx, bo.StatusPage{
		Name: req.GetName(), Title: req.GetTitle(), Description: req.GetDescription(),
		TargetGroupUIDs: req.GetTargetGroupUids(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateStatusPageResponse{StatusPage: statusPageItem(*item)}, nil
}

func (s *StatusPageService) UpdateStatusPage(ctx context.Context, req *v1.UpdateStatusPageRequest) (*v1.UpdateStatusPageResponse, error) {
	item, err := s.gateway.Update(ctx, bo.StatusPage{
		UID: req.GetUid(), Name: req.GetName(), Title: req.GetTitle(), Description: req.GetDescription(),
		TargetGroupUIDs: req.GetTargetGroupUids(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateStatusPageResponse{StatusPage: statusPageItem(*item)}, nil
}

func (s *StatusPageService) UpdateStatusPageStatus(ctx context.Context, req *v1.UpdateStatusPageStatusRequest) (*v1.UpdateStatusPageStatusResponse, error) {
	item, err := s.gateway.UpdateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateStatusPageStatusResponse{StatusPage: statusPageItem(*item)}, nil
}

func (s *StatusPageService) DeleteStatusPage(ctx context.Context, req *v1.DeleteStatusPageRequest) (*v1.DeleteStatusPageResponse, error) {
	if err := s.gateway.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteStatusPageResponse{}, nil
}

func (s *StatusPageService) GetStatusPage(ctx context.Context, req *v1.GetStatusPageRequest) (*v1.GetStatusPageResponse, error) {
	item, err := s.gateway.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetStatusPageResponse{StatusPage: statusPageItem(*item)}, nil
}

func (s *StatusPageService) ListStatusPages(ctx context.Context, req *v1.ListStatusPagesRequest) (*v1.ListStatusPagesResponse, error) {
	items, total, page, pageSize, err := s.gateway.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.StatusPageItem, 0, len(items))
	for _, item := range items {
		out = append(out, statusPageItem(item))
	}
	return &v1.ListStatusPagesResponse{
		Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *StatusPageService) RegenerateStatusPageToken(ctx context.Context, req *v1.RegenerateStatusPageTokenRequest) (*v1.RegenerateStatusPageTokenResponse, error) {
	item, err := s.gateway.RegenerateToken(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RegenerateStatusPageTokenResponse{StatusPage: statusPageItem(*item)}, nil
}

// GetPublicStatusPage 不需要登录态与租户头，只按 token 读取。
func (s *StatusPageService) GetPublicStatusPage(ctx context.Context, req *v1.GetPublicStatusPageRequest) (*v1.GetPublicStatusPageResponse, error) {
	item, err := s.gateway.Public(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}
	return &v1.GetPublicStatusPageResponse{StatusPage: publicStatusPage(item)}, nil
}

func statusPageItem(item bo.StatusPage) *v1.StatusPageItem {
	return &v1.StatusPageItem{
		Uid: item.UID, Name: item.Name, Title: item.Title, Description: item.Description,
		Status: item.Status, TargetGroupUids: item.TargetGroupUIDs, Token: item.Token,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func publicStatusPage(item *bo.PublicStatusPage) *v1.PublicStatusPage {
	components := make([]*v1.PublicStatusComponent, 0, len(item.Components))
	for _, component := range item.Components {
		lastChecked := ""
		if component.LastCheckedAt != nil {
			lastChecked = component.LastCheckedAt.UTC().Format(time.RFC3339)
		}
		components = append(components, &v1.PublicStatusComponent{
			GroupUid: component.GroupUID, Name: component.Name, Status: component.Status,
			Availability24H: component.Availability24h, Availability7D: component.Availability7d,
			Availability30D: component.Availability30d, LastCheckedAt: lastChecked,
		})
	}
	active := make([]*v1.PublicStatusEvent, 0, len(item.ActiveEvents))
	for _, event := range item.ActiveEvents {
		active = append(active, publicStatusEvent(event))
	}
	recent := make([]*v1.PublicStatusEvent, 0, len(item.RecentEvents))
	for _, event := range item.RecentEvents {
		recent = append(recent, publicStatusEvent(event))
	}
	return &v1.PublicStatusPage{
		Title: item.Title, Description: item.Description,
		UpdatedAt:  item.UpdatedAt.UTC().Format(time.RFC3339),
		Components: components, ActiveEvents: active, RecentEvents: recent,
	}
}

func publicStatusEvent(event bo.PublicStatusEvent) *v1.PublicStatusEvent {
	ended := ""
	if event.EndedAt != nil {
		ended = event.EndedAt.UTC().Format(time.RFC3339)
	}
	return &v1.PublicStatusEvent{
		Uid: event.UID, Title: event.Title, Status: event.Status,
		StartedAt: event.StartedAt.UTC().Format(time.RFC3339), EndedAt: ended,
	}
}
