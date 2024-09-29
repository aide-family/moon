package datasource

import (
	"context"
	"io"
	nethttp "net/http"

	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Service 数据源服务
type Service struct {
	datasourceapi.UnimplementedDatasourceServer

	datasourceBiz *biz.DatasourceBiz
}

// NewDatasourceService 创建数据源服务
func NewDatasourceService(datasourceBiz *biz.DatasourceBiz) *Service {
	return &Service{
		datasourceBiz: datasourceBiz,
	}
}

// CreateDatasource 创建数据源
func (s *Service) CreateDatasource(ctx context.Context, req *datasourceapi.CreateDatasourceRequest) (*datasourceapi.CreateDatasourceReply, error) {
	params := builder.NewParamsBuild().DatasourceModuleBuilder().WithCreateDatasourceRequest(req).ToBo()
	datasourceDetail, err := s.datasourceBiz.CreateDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	// 记录操作日志
	log.Debugw("datasourceDetail", datasourceDetail)
	return &datasourceapi.CreateDatasourceReply{}, nil
}

// UpdateDatasource 更新数据源
func (s *Service) UpdateDatasource(ctx context.Context, req *datasourceapi.UpdateDatasourceRequest) (*datasourceapi.UpdateDatasourceReply, error) {
	params := builder.NewParamsBuild().DatasourceModuleBuilder().WithUpdateDatasourceRequest(req).ToBo()
	if err := s.datasourceBiz.UpdateDatasourceBaseInfo(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.UpdateDatasourceReply{}, nil
}

// DeleteDatasource 删除数据源
func (s *Service) DeleteDatasource(ctx context.Context, req *datasourceapi.DeleteDatasourceRequest) (*datasourceapi.DeleteDatasourceReply, error) {
	if err := s.datasourceBiz.DeleteDatasource(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.DeleteDatasourceReply{}, nil
}

// GetDatasource 获取数据源详情
func (s *Service) GetDatasource(ctx context.Context, req *datasourceapi.GetDatasourceRequest) (*datasourceapi.GetDatasourceReply, error) {
	datasourceDetail, err := s.datasourceBiz.GetDatasource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.GetDatasourceReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToAPI(datasourceDetail),
	}, nil
}

// ListDatasource 获取数据源列表
func (s *Service) ListDatasource(ctx context.Context, req *datasourceapi.ListDatasourceRequest) (*datasourceapi.ListDatasourceReply, error) {
	params := builder.NewParamsBuild().DatasourceModuleBuilder().WithListDatasourceRequest(req).ToBo()
	datasourceList, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.ListDatasourceReply{
		Pagination: builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild().WithContext(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToAPIs(datasourceList),
	}, nil
}

// UpdateDatasourceStatus 更新数据源状态
func (s *Service) UpdateDatasourceStatus(ctx context.Context, req *datasourceapi.UpdateDatasourceStatusRequest) (*datasourceapi.UpdateDatasourceStatusReply, error) {
	if err := s.datasourceBiz.UpdateDatasourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.UpdateDatasourceStatusReply{}, nil
}

// GetDatasourceSelect 获取数据源下拉列表
func (s *Service) GetDatasourceSelect(ctx context.Context, req *datasourceapi.ListDatasourceRequest) (*datasourceapi.GetDatasourceSelectReply, error) {
	params := builder.NewParamsBuild().DatasourceModuleBuilder().WithListDatasourceRequest(req).ToBo()
	list, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.GetDatasourceSelectReply{
		List: builder.NewParamsBuild().WithContext(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToSelects(list),
	}, nil
}

// SyncDatasourceMeta 同步数据源元数据
func (s *Service) SyncDatasourceMeta(ctx context.Context, req *datasourceapi.SyncDatasourceMetaRequest) (*datasourceapi.SyncDatasourceMetaReply, error) {
	if err := s.datasourceBiz.SyncDatasourceMetaV2(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.SyncDatasourceMetaReply{}, nil
}

// DatasourceQuery 查询数据
func (s *Service) DatasourceQuery(ctx context.Context, req *datasourceapi.DatasourceQueryRequest) (*datasourceapi.DatasourceQueryReply, error) {
	params := &bo.DatasourceQueryParams{
		DatasourceID: req.GetId(),
		Query:        req.GetQuery(),
		Step:         req.GetStep(),
		TimeRange:    req.GetRange(),
	}
	query, err := s.datasourceBiz.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	return &datasourceapi.DatasourceQueryReply{
		List: builder.NewParamsBuild().MetricDataModuleBuilder().BoMetricDataBuilder().ToAPIs(query),
	}, nil
}

// ProxyQuery 查询数据
func (s *Service) ProxyQuery(ctx http.Context) error {
	// 代理http请求
	var in datasourceapi.ProxyQueryRequest
	if err := ctx.BindQuery(&in); err != nil {
		return err
	}
	if err := ctx.BindVars(&in); err != nil {
		return err
	}
	to := in.GetTo()
	if in.GetDatasourceID() > 0 {
		datasourceDetail, err := s.datasourceBiz.GetDatasource(ctx, in.GetDatasourceID())
		if !types.IsNil(err) {
			return err
		}
		to = datasourceDetail.Endpoint
	}

	req := ctx.Request()
	method := req.Method
	body := req.Body
	// 转发请求
	proxyReq, err := nethttp.NewRequestWithContext(ctx, method, to, body)
	if err != nil {
		return err
	}
	for key, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// 发起请求
	client := &nethttp.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	response := ctx.Response()
	// 将响应头复制到客户端
	for key, values := range resp.Header {
		for _, value := range values {
			response.Header().Add(key, value)
		}
	}
	// 设置响应状态码
	response.WriteHeader(resp.StatusCode)
	// 将响应体复制到客户端
	_, err = io.Copy(response, resp.Body)
	return err
}
