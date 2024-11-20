package system

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/system"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
)

// Service 系统操作服务
type Service struct {
	pb.UnimplementedSystemServer

	systemBiz *biz.SystemBiz
}

// NewSystemService 创建系统操作服务
func NewSystemService(systemBiz *biz.SystemBiz) *Service {
	return &Service{systemBiz: systemBiz}
}

// ResetTeam 重置团队
func (s *Service) ResetTeam(ctx context.Context, req *pb.ResetTeamRequest) (*pb.ResetTeamReply, error) {
	if err := s.systemBiz.ResetTeam(ctx, req.GetTeamID()); err != nil {
		return nil, err
	}
	return &pb.ResetTeamReply{}, nil
}
