package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertmaintenance"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// AlertMaintenanceService 把维护窗口接口转成 biz 入参；匹配与抑制语义都在 biz/data。
type AlertMaintenanceService struct {
	v1.UnimplementedAlertMaintenanceServiceServer
	gateway *alertmaintenance.Gateway
}

func NewAlertMaintenanceService(gateway *alertmaintenance.Gateway) *AlertMaintenanceService {
	return &AlertMaintenanceService{gateway: gateway}
}

func (s *AlertMaintenanceService) CreateAlertMaintenanceWindow(ctx context.Context, req *v1.CreateAlertMaintenanceWindowRequest) (*v1.CreateAlertMaintenanceWindowResponse, error) {
	startsAt, endsAt, err := parseMaintenanceRange(req.GetStartsAt(), req.GetEndsAt())
	if err != nil {
		return nil, err
	}
	item, err := s.gateway.Create(ctx, bo.AlertMaintenanceWindow{
		Name: req.GetName(), Description: req.GetDescription(), Labels: req.GetLabels(),
		LevelUIDs: req.GetLevelUids(), DatasourceUIDs: req.GetDatasourceUids(), RuleUIDs: req.GetRuleUids(),
		StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertMaintenanceWindowResponse{Window: maintenanceWindowItem(*item)}, nil
}

func (s *AlertMaintenanceService) UpdateAlertMaintenanceWindow(ctx context.Context, req *v1.UpdateAlertMaintenanceWindowRequest) (*v1.UpdateAlertMaintenanceWindowResponse, error) {
	startsAt, endsAt, err := parseMaintenanceRange(req.GetStartsAt(), req.GetEndsAt())
	if err != nil {
		return nil, err
	}
	item, err := s.gateway.Update(ctx, bo.AlertMaintenanceWindow{
		UID: req.GetUid(), Name: req.GetName(), Description: req.GetDescription(), Labels: req.GetLabels(),
		LevelUIDs: req.GetLevelUids(), DatasourceUIDs: req.GetDatasourceUids(), RuleUIDs: req.GetRuleUids(),
		StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertMaintenanceWindowResponse{Window: maintenanceWindowItem(*item)}, nil
}

func (s *AlertMaintenanceService) UpdateAlertMaintenanceWindowStatus(ctx context.Context, req *v1.UpdateAlertMaintenanceWindowStatusRequest) (*v1.UpdateAlertMaintenanceWindowStatusResponse, error) {
	item, err := s.gateway.UpdateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertMaintenanceWindowStatusResponse{Window: maintenanceWindowItem(*item)}, nil
}

func (s *AlertMaintenanceService) DeleteAlertMaintenanceWindow(ctx context.Context, req *v1.DeleteAlertMaintenanceWindowRequest) (*v1.DeleteAlertMaintenanceWindowResponse, error) {
	if err := s.gateway.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertMaintenanceWindowResponse{}, nil
}

func (s *AlertMaintenanceService) GetAlertMaintenanceWindow(ctx context.Context, req *v1.GetAlertMaintenanceWindowRequest) (*v1.GetAlertMaintenanceWindowResponse, error) {
	item, err := s.gateway.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertMaintenanceWindowResponse{Window: maintenanceWindowItem(*item)}, nil
}

func (s *AlertMaintenanceService) ListAlertMaintenanceWindows(ctx context.Context, req *v1.ListAlertMaintenanceWindowsRequest) (*v1.ListAlertMaintenanceWindowsResponse, error) {
	items, total, page, pageSize, err := s.gateway.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.AlertMaintenanceWindowItem, 0, len(items))
	for _, item := range items {
		out = append(out, maintenanceWindowItem(item))
	}
	return &v1.ListAlertMaintenanceWindowsResponse{
		Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

// parseMaintenanceRange 解析维护窗口的起止时间，接受 RFC3339 与带纳秒的写法。
func parseMaintenanceRange(startsAt, endsAt string) (time.Time, time.Time, error) {
	start, err := parseMaintenanceTime(startsAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseMaintenanceTime(endsAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func parseMaintenanceTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, value)
	}
	if err != nil {
		return time.Time{}, merr.Invalid("invalid time")
	}
	return parsed.UTC(), nil
}

func maintenanceWindowItem(item bo.AlertMaintenanceWindow) *v1.AlertMaintenanceWindowItem {
	return &v1.AlertMaintenanceWindowItem{
		Uid: item.UID, Name: item.Name, Description: item.Description, Labels: item.Labels,
		LevelUids: item.LevelUIDs, DatasourceUids: item.DatasourceUIDs, RuleUids: item.RuleUIDs,
		StartsAt: item.StartsAt.UTC().Format(time.RFC3339), EndsAt: item.EndsAt.UTC().Format(time.RFC3339),
		Status: item.Status, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339), CreatedBy: item.CreatedBy,
	}
}
