package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/inspection"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// InspectionPlanService 把巡检计划接口转成 biz 入参。
type InspectionPlanService struct {
	v1.UnimplementedInspectionPlanServiceServer
	gateway *inspection.Gateway
}

func NewInspectionPlanService(gateway *inspection.Gateway) *InspectionPlanService {
	return &InspectionPlanService{gateway: gateway}
}

func (s *InspectionPlanService) CreateInspectionPlan(ctx context.Context, req *v1.CreateInspectionPlanRequest) (*v1.CreateInspectionPlanResponse, error) {
	item, err := s.gateway.CreatePlan(ctx, toInspectionPlanInput(req.GetPlan()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateInspectionPlanResponse{Plan: inspectionPlanItem(*item)}, nil
}

func (s *InspectionPlanService) UpdateInspectionPlan(ctx context.Context, req *v1.UpdateInspectionPlanRequest) (*v1.UpdateInspectionPlanResponse, error) {
	in := toInspectionPlanInput(req.GetPlan())
	in.UID = req.GetUid()
	item, err := s.gateway.UpdatePlan(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateInspectionPlanResponse{Plan: inspectionPlanItem(*item)}, nil
}

func (s *InspectionPlanService) UpdateInspectionPlanStatus(ctx context.Context, req *v1.UpdateInspectionPlanStatusRequest) (*v1.UpdateInspectionPlanStatusResponse, error) {
	item, err := s.gateway.UpdatePlanStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateInspectionPlanStatusResponse{Plan: inspectionPlanItem(*item)}, nil
}

func (s *InspectionPlanService) DeleteInspectionPlan(ctx context.Context, req *v1.DeleteInspectionPlanRequest) (*v1.DeleteInspectionPlanResponse, error) {
	if err := s.gateway.DeletePlan(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteInspectionPlanResponse{}, nil
}

func (s *InspectionPlanService) GetInspectionPlan(ctx context.Context, req *v1.GetInspectionPlanRequest) (*v1.GetInspectionPlanResponse, error) {
	item, err := s.gateway.GetPlan(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetInspectionPlanResponse{Plan: inspectionPlanItem(*item)}, nil
}

func (s *InspectionPlanService) ListInspectionPlans(ctx context.Context, req *v1.ListInspectionPlansRequest) (*v1.ListInspectionPlansResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.ListPlans(ctx, req.GetPage(), req.GetPageSize(), bo.InspectionPlanListFilter{
		Name: req.GetName(), Status: req.GetStatus(), TemplateUID: req.GetTemplateUid(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.InspectionPlan, 0, len(items))
	for _, item := range items {
		out = append(out, inspectionPlanItem(item))
	}
	return &v1.ListInspectionPlansResponse{
		Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize),
	}, nil
}

// InspectionTemplateService 把报告模板接口转成 biz 入参。
type InspectionTemplateService struct {
	v1.UnimplementedInspectionTemplateServiceServer
	gateway *inspection.Gateway
}

func NewInspectionTemplateService(gateway *inspection.Gateway) *InspectionTemplateService {
	return &InspectionTemplateService{gateway: gateway}
}

func (s *InspectionTemplateService) CreateInspectionTemplate(ctx context.Context, req *v1.CreateInspectionTemplateRequest) (*v1.CreateInspectionTemplateResponse, error) {
	item, err := s.gateway.CreateTemplate(ctx, bo.InspectionTemplate{
		Name: req.GetName(), Description: req.GetDescription(),
		HTMLBody: req.GetHtmlBody(), MarkdownBody: req.GetMarkdownBody(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateInspectionTemplateResponse{Template: inspectionTemplateItem(*item)}, nil
}

func (s *InspectionTemplateService) UpdateInspectionTemplate(ctx context.Context, req *v1.UpdateInspectionTemplateRequest) (*v1.UpdateInspectionTemplateResponse, error) {
	item, err := s.gateway.UpdateTemplate(ctx, bo.InspectionTemplate{
		UID: req.GetUid(), Name: req.GetName(), Description: req.GetDescription(),
		HTMLBody: req.GetHtmlBody(), MarkdownBody: req.GetMarkdownBody(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateInspectionTemplateResponse{Template: inspectionTemplateItem(*item)}, nil
}

func (s *InspectionTemplateService) DeleteInspectionTemplate(ctx context.Context, req *v1.DeleteInspectionTemplateRequest) (*v1.DeleteInspectionTemplateResponse, error) {
	if err := s.gateway.DeleteTemplate(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteInspectionTemplateResponse{}, nil
}

func (s *InspectionTemplateService) GetInspectionTemplate(ctx context.Context, req *v1.GetInspectionTemplateRequest) (*v1.GetInspectionTemplateResponse, error) {
	item, err := s.gateway.GetTemplate(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetInspectionTemplateResponse{Template: inspectionTemplateItem(*item)}, nil
}

func (s *InspectionTemplateService) ListInspectionTemplates(ctx context.Context, req *v1.ListInspectionTemplatesRequest) (*v1.ListInspectionTemplatesResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.ListTemplates(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.InspectionTemplate, 0, len(items))
	for _, item := range items {
		out = append(out, inspectionTemplateItem(item))
	}
	return &v1.ListInspectionTemplatesResponse{
		Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize),
	}, nil
}

func toInspectionPlanInput(in *v1.InspectionPlanInput) bo.InspectionPlan {
	if in == nil {
		return bo.InspectionPlan{}
	}
	plan := bo.InspectionPlan{
		Name: in.GetName(), Description: in.GetDescription(),
		Cron: in.GetCron(), Timezone: in.GetTimezone(),
		TimeoutSeconds: int(in.GetTimeoutSeconds()),
		TemplateUID:    in.GetTemplateUid(), OutputFormats: in.GetOutputFormats(),
		TitleTemplate: in.GetTitleTemplate(), NotificationGroupUIDs: in.GetNotificationGroupUids(),
		NotifyPolicy: in.GetNotifyPolicy(), Labels: in.GetLabels(), Remark: in.GetRemark(),
	}
	plan.Items = make([]bo.InspectionItem, 0, len(in.GetItems()))
	for _, item := range in.GetItems() {
		plan.Items = append(plan.Items, bo.InspectionItem{
			Kind: item.GetKind(), Title: item.GetTitle(), DatasourceUID: item.GetDatasourceUid(),
			Query: item.GetQuery(), WindowSeconds: int(item.GetWindowSeconds()),
			StepSeconds: int(item.GetStepSeconds()), Display: item.GetDisplay(),
			Params: item.GetParams(), Enabled: item.GetEnabled(),
		})
	}
	return plan
}

func inspectionPlanItem(in bo.InspectionPlan) *v1.InspectionPlan {
	items := make([]*v1.InspectionPlanItem, 0, len(in.Items))
	for _, item := range in.Items {
		items = append(items, &v1.InspectionPlanItem{
			Uid: item.UID, Kind: item.Kind, Title: item.Title, Sort: int32(item.Sort),
			DatasourceUid: item.DatasourceUID, Query: item.Query,
			WindowSeconds: int32(item.WindowSeconds), StepSeconds: int32(item.StepSeconds),
			Display: item.Display, Params: item.Params, Enabled: item.Enabled,
		})
	}
	return &v1.InspectionPlan{
		Uid: in.UID, Name: in.Name, Description: in.Description, Status: in.Status,
		Cron: in.Cron, Timezone: in.Timezone, TimeoutSeconds: int32(in.TimeoutSeconds),
		TemplateUid: in.TemplateUID, OutputFormats: in.OutputFormats,
		TitleTemplate: in.TitleTemplate, NotificationGroupUids: in.NotificationGroupUIDs,
		NotifyPolicy: in.NotifyPolicy, Labels: in.Labels, Remark: in.Remark,
		Items: items, NextRunAt: formatOptionalTime(in.NextRunAt), LastRunAt: formatOptionalTime(in.LastRunAt),
		CreatedBy: in.CreatedBy, CreatedAt: formatTime(in.CreatedAt), UpdatedAt: formatTime(in.UpdatedAt),
	}
}

func inspectionTemplateItem(in bo.InspectionTemplate) *v1.InspectionTemplate {
	return &v1.InspectionTemplate{
		Uid: in.UID, Name: in.Name, Description: in.Description,
		HtmlBody: in.HTMLBody, MarkdownBody: in.MarkdownBody,
		IsBuiltin: in.IsBuiltin, Status: in.Status,
		CreatedAt: formatTime(in.CreatedAt), UpdatedAt: formatTime(in.UpdatedAt),
	}
}

// InspectionRunService 把执行记录接口转成 biz 入参。
type InspectionRunService struct {
	v1.UnimplementedInspectionRunServiceServer
	gateway *inspection.Gateway
}

func NewInspectionRunService(gateway *inspection.Gateway) *InspectionRunService {
	return &InspectionRunService{gateway: gateway}
}

func (s *InspectionRunService) RunInspectionPlan(ctx context.Context, req *v1.RunInspectionPlanRequest) (*v1.RunInspectionPlanResponse, error) {
	run, err := s.gateway.Trigger(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RunInspectionPlanResponse{Run: inspectionRunItem(*run)}, nil
}

func (s *InspectionRunService) ListInspectionRuns(ctx context.Context, req *v1.ListInspectionRunsRequest) (*v1.ListInspectionRunsResponse, error) {
	filter := bo.InspectionRunListFilter{
		PlanUID: req.GetPlanUid(), Status: req.GetStatus(), Trigger: req.GetTrigger(),
	}
	if req.GetFrom() != "" {
		value, err := time.Parse(time.RFC3339, req.GetFrom())
		if err != nil {
			return nil, merr.Invalid("invalid from time")
		}
		filter.From = &value
	}
	if req.GetTo() != "" {
		value, err := time.Parse(time.RFC3339, req.GetTo())
		if err != nil {
			return nil, merr.Invalid("invalid to time")
		}
		filter.To = &value
	}
	items, total, pageNo, pageSize, err := s.gateway.ListRuns(ctx, req.GetPage(), req.GetPageSize(), filter)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.InspectionRun, 0, len(items))
	for _, item := range items {
		out = append(out, inspectionRunItem(item))
	}
	return &v1.ListInspectionRunsResponse{
		Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize),
	}, nil
}

func (s *InspectionRunService) GetInspectionRun(ctx context.Context, req *v1.GetInspectionRunRequest) (*v1.GetInspectionRunResponse, error) {
	run, err := s.gateway.GetRun(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetInspectionRunResponse{Run: inspectionRunItem(*run)}, nil
}

func inspectionRunItem(in bo.InspectionRun) *v1.InspectionRun {
	return &v1.InspectionRun{
		Uid: in.UID, PlanUid: in.PlanUID, PlanName: in.PlanName, Trigger: in.Trigger, Status: in.Status,
		ScheduledAt: formatTime(in.ScheduledAt), StartedAt: formatOptionalTime(in.StartedAt),
		FinishedAt:  formatOptionalTime(in.FinishedAt),
		WindowStart: formatOptionalTime(in.WindowStart), WindowEnd: formatOptionalTime(in.WindowEnd),
		ItemTotal: int32(in.ItemTotal), ItemOk: int32(in.ItemOK), ItemWarn: int32(in.ItemWarn),
		ItemFailed: int32(in.ItemFailed), Summary: in.Summary, Error: in.Error,
		OperatorName: in.OperatorName, NotifiedAt: formatOptionalTime(in.NotifiedAt),
	}
}

func (s *InspectionRunService) ListRunReports(ctx context.Context, req *v1.ListRunReportsRequest) (*v1.ListRunReportsResponse, error) {
	items, err := s.gateway.ListReports(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.InspectionReport, 0, len(items))
	for _, item := range items {
		out = append(out, inspectionReportItem(item, false))
	}
	return &v1.ListRunReportsResponse{Items: out}, nil
}

func (s *InspectionRunService) GetInspectionReport(ctx context.Context, req *v1.GetInspectionReportRequest) (*v1.GetInspectionReportResponse, error) {
	item, err := s.gateway.GetReport(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetInspectionReportResponse{Report: inspectionReportItem(*item, true)}, nil
}

func inspectionReportItem(in bo.InspectionReport, includeContent bool) *v1.InspectionReport {
	item := &v1.InspectionReport{
		Uid: in.UID, RunUid: in.RunUID, Format: in.Format, Title: in.Title,
		Size: int32(in.Size), ShareEnabled: in.ShareEnabled,
		ShareExpiresAt: formatOptionalTime(in.ShareExpiresAt),
		CreatedAt:      formatTime(in.CreatedAt),
	}
	if includeContent {
		item.Content = in.Content
	}
	return item
}
