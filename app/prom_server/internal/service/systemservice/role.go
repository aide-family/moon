package systemservice

import (
	"context"

	pb "prometheus-manager/api/system"
)

type RoleService struct {
	pb.UnimplementedRoleServer
}

func NewRoleService() *RoleService {
	return &RoleService{}
}

func (s *RoleService) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleReply, error) {
	return &pb.CreateRoleReply{}, nil
}
func (s *RoleService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleReply, error) {
	return &pb.UpdateRoleReply{}, nil
}
func (s *RoleService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleReply, error) {
	return &pb.DeleteRoleReply{}, nil
}
func (s *RoleService) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.GetRoleReply, error) {
	return &pb.GetRoleReply{}, nil
}
func (s *RoleService) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	return &pb.ListRoleReply{}, nil
}
func (s *RoleService) SelectRole(ctx context.Context, req *pb.SelectRoleRequest) (*pb.SelectRoleReply, error) {
	return &pb.SelectRoleReply{}, nil
}
