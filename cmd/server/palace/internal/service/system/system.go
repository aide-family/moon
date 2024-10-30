package system

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/system"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
)

type Service struct {
	pb.UnimplementedSystemServer

	systemBiz *biz.SystemBiz
}

func NewSystemService(systemBiz *biz.SystemBiz) *Service {
	return &Service{systemBiz: systemBiz}
}

func (s *Service) ResetTeam(ctx context.Context, req *pb.ResetTeamRequest) (*pb.ResetTeamReply, error) {
	if err := s.systemBiz.ResetTeam(ctx, req.GetTeamID()); err != nil {
		return nil, err
	}
	return &pb.ResetTeamReply{}, nil
}
