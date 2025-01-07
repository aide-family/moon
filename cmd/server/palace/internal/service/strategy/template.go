package strategy

import (
	"context"
	"time"

	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
)

// TemplateService 模板策略服务
type TemplateService struct {
	strategyapi.UnimplementedTemplateServer

	templateBiz   *biz.TemplateBiz
	datasourceBiz *biz.DatasourceBiz
}

// NewTemplateService 创建模板策略服务
func NewTemplateService(templateBiz *biz.TemplateBiz, datasourceBiz *biz.DatasourceBiz) *TemplateService {
	return &TemplateService{
		templateBiz:   templateBiz,
		datasourceBiz: datasourceBiz,
	}
}

// CreateTemplateStrategy 创建模板策略
func (s *TemplateService) CreateTemplateStrategy(ctx context.Context, req *strategyapi.CreateTemplateStrategyRequest) (*strategyapi.CreateTemplateStrategyReply, error) {
	params := builder.NewParamsBuild(ctx).StrategyModuleBuilder().WithCreateTemplateStrategyRequest(req).ToBo()
	if err := s.templateBiz.CreateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.CreateTemplateStrategyReply{}, nil
}

// UpdateTemplateStrategy 更新模板策略
func (s *TemplateService) UpdateTemplateStrategy(ctx context.Context, req *strategyapi.UpdateTemplateStrategyRequest) (*strategyapi.UpdateTemplateStrategyReply, error) {
	params := builder.NewParamsBuild(ctx).StrategyModuleBuilder().WithUpdateTemplateStrategyRequest(req).ToBo()
	if err := s.templateBiz.UpdateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyReply{}, nil
}

// DeleteTemplateStrategy 删除模板策略
func (s *TemplateService) DeleteTemplateStrategy(ctx context.Context, req *strategyapi.DeleteTemplateStrategyRequest) (*strategyapi.DeleteTemplateStrategyReply, error) {
	if err := s.templateBiz.DeleteTemplateStrategy(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &strategyapi.DeleteTemplateStrategyReply{}, nil
}

// GetTemplateStrategy 获取模板策略
func (s *TemplateService) GetTemplateStrategy(ctx context.Context, req *strategyapi.GetTemplateStrategyRequest) (*strategyapi.GetTemplateStrategyReply, error) {
	detail, err := s.templateBiz.GetTemplateStrategy(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &strategyapi.GetTemplateStrategyReply{
		Detail: builder.NewParamsBuild(ctx).StrategyModuleBuilder().DoTemplateStrategyBuilder().ToAPI(detail),
	}, nil
}

// ListTemplateStrategy 列表模板策略
func (s *TemplateService) ListTemplateStrategy(ctx context.Context, req *strategyapi.ListTemplateStrategyRequest) (*strategyapi.ListTemplateStrategyReply, error) {
	params := &bo.QueryTemplateStrategyListParams{
		Page:    types.NewPagination(req.GetPagination()),
		Status:  vobj.Status(req.GetStatus()),
		Keyword: req.GetKeyword(),
	}
	list, err := s.templateBiz.ListTemplateStrategy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &strategyapi.ListTemplateStrategyReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).StrategyModuleBuilder().DoTemplateStrategyBuilder().ToAPIs(list),
	}, nil
}

// UpdateTemplateStrategyStatus 更新模板策略状态
func (s *TemplateService) UpdateTemplateStrategyStatus(ctx context.Context, req *strategyapi.UpdateTemplateStrategyStatusRequest) (*strategyapi.UpdateTemplateStrategyStatusReply, error) {
	params := builder.NewParamsBuild(ctx).StrategyModuleBuilder().WithUpdateTemplateStrategyStatusRequest(req).ToBo()
	if err := s.templateBiz.UpdateTemplateStrategyStatus(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyStatusReply{}, nil
}

// ValidateAnnotationsTemplate 验证模板策略告警模板
func (s *TemplateService) ValidateAnnotationsTemplate(ctx context.Context, req *strategyapi.ValidateAnnotationsTemplateRequest) (*strategyapi.ValidateAnnotationsTemplateReply, error) {
	timeNow := time.Now()
	data := map[string]any{
		// 策略告警时候的值
		"value": 0.00,
		// 策略告警unix时间戳
		"eventAt": timeNow.Unix(),
		// 策略告警标签
		"labels": label.NewLabels(req.GetLabels()),
		// 策略明细
		"strategy": map[string]any{
			// 策略名称
			"alert": req.GetAlert(),
			// 策略等级
			"level": req.GetLevel(),
			// 策略告警表达式
			"expr": req.GetExpr(),
			// 持续时间
			"duration": req.GetDuration(),
			// 持续次数
			"count": req.GetCount(),
			// 持续类型
			"sustainType": vobj.Sustain(req.GetSustainType()).String(),
			// 告警条件
			"condition": vobj.Condition(req.GetCondition()).String(),
			// 告警阈值
			"threshold": req.GetThreshold(),
		},
	}
	labels := label.NewLabels(req.GetLabels())
	queryParams := &bo.DatasourceQueryParams{
		DatasourceID: req.GetDatasourceId(), // TODO 增加数据源支持
		Query:        req.GetExpr(),
		Step:         0,
		TimeRange:    []string{timeNow.Format(time.DateTime)},
	}
	if req.GetDatasource() != "" {
		queryParams.Datasource = &bizmodel.Datasource{
			Endpoint:    req.GetDatasource(),
			StorageType: vobj.StorageTypePrometheus,
			Category:    vobj.DatasourceTypeMetrics,
		}
	}
	queryData, err := s.datasourceBiz.Query(ctx, queryParams)
	if err != nil {
		return nil, err
	}
	log.Debugw("queryData", queryData)
	if len(queryData) > 0 {
		for _, datum := range queryData {
			if types.IsNil(datum) {
				continue
			}
			labelsTmp := types.MapsMerge(labels.Map(), datum.Labels)
			labels = label.NewLabels(labelsTmp)
		}
		data["labels"] = labels
		data["value"] = queryData[0].Value.Value
		data["eventAt"] = queryData[0].Value.Timestamp
	}

	log.Debugw("labels", labels)

	formatterWithErr, err := format.FormatterWithErr(req.GetAnnotations(), data)

	errorString := ""
	if err != nil {
		errorString = err.Error()
	}
	labelsString := make([]string, 0, len(labels.Map()))
	for k := range labels.Map() {
		labelsString = append(labelsString, k)
	}
	return &strategyapi.ValidateAnnotationsTemplateReply{
		Annotations: formatterWithErr,
		Errors:      errorString,
		Labels:      labelsString,
	}, nil
}
