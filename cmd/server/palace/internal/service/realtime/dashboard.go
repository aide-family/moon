package realtime

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/realtime"
)

// DashboardService 监控大盘服务
type DashboardService struct {
	pb.UnimplementedDashboardServer
}

// NewDashboardService 创建监控大盘服务
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// CreateDashboard 创建监控大盘
func (s *DashboardService) CreateDashboard(ctx context.Context, req *pb.CreateDashboardRequest) (*pb.CreateDashboardReply, error) {
	return &pb.CreateDashboardReply{}, nil
}

// UpdateDashboard 更新监控大盘
func (s *DashboardService) UpdateDashboard(ctx context.Context, req *pb.UpdateDashboardRequest) (*pb.UpdateDashboardReply, error) {
	return &pb.UpdateDashboardReply{}, nil
}

// DeleteDashboard 删除监控大盘
func (s *DashboardService) DeleteDashboard(ctx context.Context, req *pb.DeleteDashboardRequest) (*pb.DeleteDashboardReply, error) {
	return &pb.DeleteDashboardReply{}, nil
}

// GetDashboard 获取监控大盘
func (s *DashboardService) GetDashboard(ctx context.Context, req *pb.GetDashboardRequest) (*pb.GetDashboardReply, error) {
	return &pb.GetDashboardReply{}, nil
}

// ListDashboard 获取监控大盘列表
func (s *DashboardService) ListDashboard(ctx context.Context, req *pb.ListDashboardRequest) (*pb.ListDashboardReply, error) {
	return &pb.ListDashboardReply{}, nil
}

// ListDashboardSelect 获取监控大盘下拉列表
func (s *DashboardService) ListDashboardSelect(ctx context.Context, req *pb.ListDashboardSelectRequest) (*pb.ListDashboardSelectReply, error) {
	return &pb.ListDashboardSelectReply{}, nil
}
