package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/strategy/group"
)

type GroupService struct {
	pb.UnimplementedGroupServer

	log *log.Helper
}

func NewGroupService(logger log.Logger) *GroupService {
	return &GroupService{
		log: log.NewHelper(log.With(logger, "module", "service.prom.strategy.group")),
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	return &pb.CreateGroupReply{}, nil
}
func (s *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	return &pb.UpdateGroupReply{}, nil
}
func (s *GroupService) BatchUpdateGroupStatus(ctx context.Context, req *pb.BatchUpdateGroupStatusRequest) (*pb.BatchUpdateGroupStatusReply, error) {
	return &pb.BatchUpdateGroupStatusReply{}, nil
}
func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	return &pb.DeleteGroupReply{}, nil
}
func (s *GroupService) BatchDeleteGroup(ctx context.Context, req *pb.BatchDeleteGroupRequest) (*pb.BatchDeleteGroupReply, error) {
	return &pb.BatchDeleteGroupReply{}, nil
}
func (s *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	return &pb.GetGroupReply{}, nil
}
func (s *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	return &pb.ListGroupReply{}, nil
}
func (s *GroupService) SelectGroup(ctx context.Context, req *pb.SelectGroupRequest) (*pb.SelectGroupReply, error) {
	return &pb.SelectGroupReply{}, nil
}
func (s *GroupService) ImportGroup(ctx context.Context, req *pb.ImportGroupRequest) (*pb.ImportGroupReply, error) {
	return &pb.ImportGroupReply{}, nil
}
func (s *GroupService) ExportGroup(ctx context.Context, req *pb.ExportGroupRequest) (*pb.ExportGroupReply, error) {
	return &pb.ExportGroupReply{}, nil
}
