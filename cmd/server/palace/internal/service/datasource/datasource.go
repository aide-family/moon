package datasource

import (
	"context"
	"io"
	nethttp "net/http"
	"net/url"
	"regexp"
	"strings"

	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
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
	params := builder.NewParamsBuild(ctx).DatasourceModuleBuilder().WithCreateDatasourceRequest(req).ToBo()
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
	params := builder.NewParamsBuild(ctx).DatasourceModuleBuilder().WithUpdateDatasourceRequest(req).ToBo()
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
		Detail: builder.NewParamsBuild(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToAPI(datasourceDetail),
	}, nil
}

// ListDatasource 获取数据源列表
func (s *Service) ListDatasource(ctx context.Context, req *datasourceapi.ListDatasourceRequest) (*datasourceapi.ListDatasourceReply, error) {
	params := builder.NewParamsBuild(ctx).DatasourceModuleBuilder().WithListDatasourceRequest(req).ToBo()
	datasourceList, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.ListDatasourceReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToAPIs(datasourceList),
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
	params := builder.NewParamsBuild(ctx).DatasourceModuleBuilder().WithListDatasourceRequest(req).ToBo()
	list, err := s.datasourceBiz.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &datasourceapi.GetDatasourceSelectReply{
		List: builder.NewParamsBuild(ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToSelects(list),
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
		List: builder.NewParamsBuild(ctx).MetricDataModuleBuilder().BoMetricDataBuilder().ToAPIs(query),
	}, nil
}

// DataSourceProxy 数据源健康检查
func (s *Service) DataSourceProxy() http.HandlerFunc {
	return func(ctx http.Context) error {
		var in datasourceapi.DataSourceHealthRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		if !isValidURL(in.Url) {
			return merr.ErrorAlert("数据源地址错误，请检查")
		}
		toURL, err := url.JoinPath(in.Url, "/-/ready")
		if !types.IsNil(err) {
			return err
		}
		log.Debugw("to", toURL)
		return s.proxy(ctx, toURL)
	}
}

// isValidURL 验证URL是否有效
func isValidURL(url string) bool {
	// 定义正则表达式来匹配网址
	regex := `^(https?|ftp):\/\/(?:www\.)?((?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}|(?:\d{1,3}\.){3}\d{1,3}|(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{1,4})(?::\d{1,5})?(\/[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;%=]*)?$`
	re := regexp.MustCompile(regex)

	// 使用正则表达式进行匹配
	return re.MatchString(url)
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

// MetricProxy 指标数据源代理
func (s *Service) MetricProxy() http.HandlerFunc {
	return func(ctx http.Context) error {
		isContentType := false
		for k := range ctx.Request().Header {
			if strings.EqualFold(k, "Content-Type") {
				isContentType = true
				break
			}
		}
		if !isContentType {
			ctx.Header().Set("Content-Type", "application/json")
			ctx.Request().Header.Set("Content-Type", "application/json")
		}

		var in datasourceapi.ProxyMetricDatasourceQueryRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}

		// 获取请求头JWT
		//token := ctx.Header().Get("Authorization")
		//auths := strings.SplitN(token, " ", 2)
		//if len(auths) != 2 || !strings.EqualFold(auths[0], "Bearer") {
		//	return jwt.ErrMissingJwtToken
		//}
		//jwtToken := auths[1]
		//log.Debugw("jwtToken", jwtToken)

		_ctx := middleware.WithTeamIDContextKey(ctx, in.GetTeamID())
		log.Debugw("teamID", middleware.GetTeamID(_ctx), "req", &in)
		datasourceDetail, err := s.datasourceBiz.GetDatasource(_ctx, in.GetId())
		if !types.IsNil(err) {
			log.Errorw("err", err)
			return err
		}

		to, err := url.JoinPath(datasourceDetail.Endpoint, in.To)
		if !types.IsNil(err) {
			return err
		}
		log.Debugw("to", to)
		// 直接转发请求
		return s.proxy(ctx, to)
	}
}

// proxy
func (s *Service) proxy(ctx http.Context, to string) error {
	w := ctx.Response()
	r := ctx.Request()

	// 获取query data
	query := r.URL.Query()
	// 绑定query到to
	toURL, err := url.Parse(to)
	if !types.IsNil(err) {
		return err
	}
	toURL.RawQuery = query.Encode()
	// body
	body := r.Body
	//
	// 发起一个新请求， 把数据写回w
	proxyReq, err := nethttp.NewRequestWithContext(ctx, r.Method, toURL.String(), body)
	if !types.IsNil(err) {
		return err
	}
	proxyReq.Header = r.Header
	proxyReq.Form = r.Form
	proxyReq.Body = r.Body
	client := &nethttp.Client{}
	resp, err := client.Do(proxyReq)
	if !types.IsNil(err) {
		return err
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		if len(v) == 0 {
			continue
		}
		w.Header().Set(k, v[0])
	}
	_, err = io.Copy(w, resp.Body)
	return err
}
