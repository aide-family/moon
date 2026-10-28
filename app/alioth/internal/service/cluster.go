package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/cluster"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ClusterService 把集群管理请求转成用例入参。kubeconfig 只作为入参出现，
// 响应里永远没有它。
type ClusterService struct {
	v1.UnimplementedClusterServiceServer
	registry *cluster.Registry
}

func NewClusterService(registry *cluster.Registry) *ClusterService {
	return &ClusterService{registry: registry}
}

func (s *ClusterService) CreateCluster(ctx context.Context, req *v1.CreateClusterRequest) (*v1.CreateClusterResponse, error) {
	item, err := s.registry.Create(ctx, req.GetName(), req.GetKubeconfig(), req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.CreateClusterResponse{Cluster: clusterItem(*item)}, nil
}

func (s *ClusterService) UpdateCluster(ctx context.Context, req *v1.UpdateClusterRequest) (*v1.UpdateClusterResponse, error) {
	item, err := s.registry.Update(ctx, req.GetUid(), req.GetName(), req.GetKubeconfig(), req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateClusterResponse{Cluster: clusterItem(*item)}, nil
}

func (s *ClusterService) DeleteCluster(ctx context.Context, req *v1.DeleteClusterRequest) (*v1.DeleteClusterResponse, error) {
	if err := s.registry.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteClusterResponse{}, nil
}

func (s *ClusterService) GetCluster(ctx context.Context, req *v1.GetClusterRequest) (*v1.GetClusterResponse, error) {
	item, err := s.registry.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetClusterResponse{Cluster: clusterItem(*item)}, nil
}

func (s *ClusterService) ListClusters(ctx context.Context, req *v1.ListClustersRequest) (*v1.ListClustersResponse, error) {
	items, total, page, pageSize, err := s.registry.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ClusterItem, 0, len(items))
	for _, item := range items {
		out = append(out, clusterItem(item))
	}
	return &v1.ListClustersResponse{
		Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *ClusterService) TestCluster(ctx context.Context, req *v1.TestClusterRequest) (*v1.TestClusterResponse, error) {
	probe, item, err := s.registry.Test(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.TestClusterResponse{Probe: clusterProbe(probe), Cluster: clusterItem(*item)}, nil
}

func (s *ClusterService) ProbeKubeconfig(ctx context.Context, req *v1.ProbeKubeconfigRequest) (*v1.ProbeKubeconfigResponse, error) {
	probe, err := s.registry.Probe(ctx, req.GetKubeconfig())
	if err != nil {
		return nil, err
	}
	return &v1.ProbeKubeconfigResponse{Probe: clusterProbe(probe)}, nil
}

func clusterItem(in bo.Cluster) *v1.ClusterItem {
	item := &v1.ClusterItem{
		Uid:           in.UID,
		Name:          in.Name,
		Server:        in.Server,
		Version:       in.Version,
		ClusterName:   in.ClusterName,
		ContextName:   in.ContextName,
		Remark:        in.Remark,
		LastCheckedAt: in.LastCheckedAt,
		LastError:     in.LastError,
	}
	if !in.CreatedAt.IsZero() {
		item.CreatedAt = in.CreatedAt.Unix()
	}
	return item
}

func clusterProbe(in *bo.ClusterProbe) *v1.ClusterProbe {
	if in == nil {
		return nil
	}
	return &v1.ClusterProbe{
		Server:      in.Server,
		Version:     in.Version,
		ClusterName: in.ClusterName,
		ContextName: in.ContextName,
		NodeCount:   in.NodeCount,
	}
}
