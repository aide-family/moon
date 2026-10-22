package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type DatasourceService struct {
	v1.UnimplementedDatasourceServiceServer
	gateway *datasource.Gateway
}

func NewDatasourceService(gateway *datasource.Gateway) *DatasourceService {
	return &DatasourceService{gateway: gateway}
}

func (s *DatasourceService) CreateDatasource(ctx context.Context, req *v1.CreateDatasourceRequest) (*v1.CreateDatasourceResponse, error) {
	item, err := s.gateway.CreateDatasource(ctx, datasourceInput(req))
	if err != nil {
		return nil, err
	}
	return &v1.CreateDatasourceResponse{Datasource: datasourceItem(*item)}, nil
}

func (s *DatasourceService) UpdateDatasource(ctx context.Context, req *v1.UpdateDatasourceRequest) (*v1.UpdateDatasourceResponse, error) {
	in := datasourceInput(req)
	in.UID = req.GetUid()
	item, err := s.gateway.UpdateDatasource(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateDatasourceResponse{Datasource: datasourceItem(*item)}, nil
}

func (s *DatasourceService) UpdateDatasourceStatus(ctx context.Context, req *v1.UpdateDatasourceStatusRequest) (*v1.UpdateDatasourceStatusResponse, error) {
	item, err := s.gateway.UpdateDatasourceStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateDatasourceStatusResponse{Datasource: datasourceItem(*item)}, nil
}

func (s *DatasourceService) DeleteDatasource(ctx context.Context, req *v1.DeleteDatasourceRequest) (*v1.DeleteDatasourceResponse, error) {
	if err := s.gateway.DeleteDatasource(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteDatasourceResponse{}, nil
}

func (s *DatasourceService) GetDatasource(ctx context.Context, req *v1.GetDatasourceRequest) (*v1.GetDatasourceResponse, error) {
	item, err := s.gateway.GetDatasource(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetDatasourceResponse{Datasource: datasourceItem(*item)}, nil
}

func (s *DatasourceService) ListDatasources(ctx context.Context, req *v1.ListDatasourcesRequest) (*v1.ListDatasourcesResponse, error) {
	items, total, page, pageSize, err := s.gateway.ListDatasources(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetType())
	if err != nil {
		return nil, err
	}
	return &v1.ListDatasourcesResponse{Items: datasourceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *DatasourceService) SelectDatasources(ctx context.Context, req *v1.SelectDatasourcesRequest) (*v1.SelectDatasourcesResponse, error) {
	items, err := s.gateway.SelectDatasources(ctx, req.GetLimit(), req.GetLastUid(), req.GetType())
	if err != nil {
		return nil, err
	}
	return &v1.SelectDatasourcesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *DatasourceService) GetDatasourceStatus(ctx context.Context, req *v1.GetDatasourceStatusRequest) (*v1.GetDatasourceStatusResponse, error) {
	status, err := s.gateway.Status(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetDatasourceStatusResponse{Status: status}, nil
}

func (s *DatasourceService) GetDatasourceMetadata(ctx context.Context, req *v1.GetDatasourceMetadataRequest) (*v1.GetDatasourceMetadataResponse, error) {
	metadata, err := s.gateway.Metadata(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DatasourceMetadataItem, 0, len(metadata.Items))
	for _, item := range metadata.Items {
		items = append(items, &v1.DatasourceMetadataItem{Key: item.Key, Value: item.Value})
	}
	nodes := make([]*v1.DatasourceMetadataNode, 0, len(metadata.Nodes))
	for _, node := range metadata.Nodes {
		nodes = append(nodes, &v1.DatasourceMetadataNode{
			Name: node.Name, Address: node.Address, Role: node.Role, Shards: node.Shards,
			DiskUsed: node.DiskUsed, DiskAvail: node.DiskAvail, DiskPercent: node.DiskPercent,
		})
	}
	return &v1.GetDatasourceMetadataResponse{Driver: metadata.Driver, Items: items, Nodes: nodes}, nil
}

func (s *DatasourceService) ListDatasourceMetrics(ctx context.Context, req *v1.ListDatasourceMetricsRequest) (*v1.ListDatasourceMetricsResponse, error) {
	metrics, err := s.gateway.Metrics(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(metrics))
	items := make([]*v1.DatasourceMetricItem, 0, len(metrics))
	for _, metric := range metrics {
		names = append(names, metric.Name)
		items = append(items, &v1.DatasourceMetricItem{
			Name: metric.Name,
			Help: metric.Help,
			Type: metric.Type,
			Unit: metric.Unit,
		})
	}
	return &v1.ListDatasourceMetricsResponse{Names: names, Metrics: items}, nil
}

func (s *DatasourceService) GetDatasourceMetric(ctx context.Context, req *v1.GetDatasourceMetricRequest) (*v1.GetDatasourceMetricResponse, error) {
	body, err := s.gateway.Metric(ctx, req.GetUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetDatasourceMetricResponse{Body: body}, nil
}

// datasourceRequest 是创建和更新请求的公共字段。两个生成的请求类型都实现它。
type datasourceRequest interface {
	GetName() string
	GetType() string
	GetDriver() string
	GetEndpoint() string
	GetIndexPattern() string
	GetTimeField() string
	GetUsername() string
	GetPassword() string
	GetSecret() string
	GetLevelUid() string
	GetRemark() string
	GetLabels() string
}

func datasourceInput(req datasourceRequest) bo.Datasource {
	return bo.Datasource{
		Name: req.GetName(), Type: req.GetType(), Driver: req.GetDriver(), Endpoint: req.GetEndpoint(),
		IndexPattern: req.GetIndexPattern(), TimeField: req.GetTimeField(), Username: req.GetUsername(),
		Password: req.GetPassword(), Secret: req.GetSecret(), LevelUID: req.GetLevelUid(),
		Remark: req.GetRemark(), Labels: req.GetLabels(),
	}
}
