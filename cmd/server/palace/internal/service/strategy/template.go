package strategy

import (
	"context"
	"time"

	"github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
)

type TemplateService struct {
	strategyapi.UnimplementedTemplateServer

	templateBiz   *biz.TemplateBiz
	datasourceBiz *biz.DatasourceBiz
}

func NewTemplateService(templateBiz *biz.TemplateBiz, datasourceBiz *biz.DatasourceBiz) *TemplateService {
	return &TemplateService{
		templateBiz:   templateBiz,
		datasourceBiz: datasourceBiz,
	}
}

func (s *TemplateService) CreateTemplateStrategy(ctx context.Context, req *strategyapi.CreateTemplateStrategyRequest) (*strategyapi.CreateTemplateStrategyReply, error) {
	params := build.NewBuilder().WithCreateBoTemplateStrategy(req).ToCreateTemplateBO()
	if err := s.templateBiz.CreateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.CreateTemplateStrategyReply{}, nil
}

func (s *TemplateService) UpdateTemplateStrategy(ctx context.Context, req *strategyapi.UpdateTemplateStrategyRequest) (*strategyapi.UpdateTemplateStrategyReply, error) {
	params := build.NewBuilder().WithUpdateBoTemplateStrategy(req).ToUpdateTemplateBO()
	if err := s.templateBiz.UpdateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyReply{}, nil
}

func (s *TemplateService) DeleteTemplateStrategy(ctx context.Context, req *strategyapi.DeleteTemplateStrategyRequest) (*strategyapi.DeleteTemplateStrategyReply, error) {
	if err := s.templateBiz.DeleteTemplateStrategy(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &strategyapi.DeleteTemplateStrategyReply{}, nil
}

func (s *TemplateService) GetTemplateStrategy(ctx context.Context, req *strategyapi.GetTemplateStrategyRequest) (*strategyapi.GetTemplateStrategyReply, error) {
	detail, err := s.templateBiz.GetTemplateStrategy(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &strategyapi.GetTemplateStrategyReply{
		Detail: build.NewBuilder().WithApiTemplateStrategy(detail).ToApi(ctx),
	}, nil
}

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
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *model.StrategyTemplate) *admin.StrategyTemplate {
			return build.NewBuilder().WithApiTemplateStrategy(item).ToApi(ctx)
		}),
	}, nil
}

func (s *TemplateService) UpdateTemplateStrategyStatus(ctx context.Context, req *strategyapi.UpdateTemplateStrategyStatusRequest) (*strategyapi.UpdateTemplateStrategyStatusReply, error) {
	if err := s.templateBiz.UpdateTemplateStrategyStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()...); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyStatusReply{}, nil
}

func (s *TemplateService) ValidateAnnotationsTemplate(ctx context.Context, req *strategyapi.ValidateAnnotationsTemplateRequest) (*strategyapi.ValidateAnnotationsTemplateReply, error) {
	timeNow := time.Now()
	data := map[string]any{
		"alert":     req.GetAlert(),
		"level":     req.GetLevel(),
		"value":     0.00,
		"timestamp": timeNow.Unix(),
		"labels":    vobj.LabelsJSON(req.GetLabels()),
	}
	labels := req.GetLabels()
	queryParams := &bo.DatasourceQueryParams{
		DatasourceID: 1, // TODO 增加数据源支持
		Query:        req.GetExpr(),
		Step:         0,
		TimeRange:    []string{timeNow.Format(time.DateTime)},
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
			labels = types.MapsMerge(labels, datum.Labels)
		}
		data["labels"] = vobj.LabelsJSON(labels)
		data["value"] = queryData[0].Value.Value
		data["timestamp"] = queryData[0].Value.Timestamp
	}

	log.Debugw("labels", labels)

	formatterWithErr, err := format.FormatterWithErr(req.GetAnnotations(), data)

	errorString := ""
	if err != nil {
		errorString = err.Error()
	}

	return &strategyapi.ValidateAnnotationsTemplateReply{
		Annotations: formatterWithErr,
		Errors:      errorString,
	}, nil
}
