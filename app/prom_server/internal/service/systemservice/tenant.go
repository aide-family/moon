package systemservice

import (
	"context"

	pb "github.com/aide-family/moon/api/server/system"
	"github.com/go-kratos/kratos/v2/log"
)

type TenantService struct {
	pb.UnimplementedTenantServer

	log *log.Helper
}

func NewTenantService(logger log.Logger) *TenantService {
	return &TenantService{
		log: log.NewHelper(log.With(logger, "module", "system-service.tenant")),
	}
}

func (s *TenantService) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantReply, error) {
	return &pb.CreateTenantReply{}, nil
}
func (s *TenantService) UpdateTenant(ctx context.Context, req *pb.UpdateTenantRequest) (*pb.UpdateTenantReply, error) {
	return &pb.UpdateTenantReply{}, nil
}
func (s *TenantService) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest) (*pb.DeleteTenantReply, error) {
	return &pb.DeleteTenantReply{}, nil
}
func (s *TenantService) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantReply, error) {
	return &pb.GetTenantReply{}, nil
}
func (s *TenantService) ListTenant(ctx context.Context, req *pb.ListTenantRequest) (*pb.ListTenantReply, error) {
	return &pb.ListTenantReply{}, nil
}
func (s *TenantService) ManageAdmin(ctx context.Context, req *pb.ManageAdminRequest) (*pb.ManageAdminReply, error) {
	return &pb.ManageAdminReply{}, nil
}
func (s *TenantService) TransferAdmin(ctx context.Context, req *pb.TransferAdminRequest) (*pb.TransferAdminReply, error) {
	return &pb.TransferAdminReply{}, nil
}
func (s *TenantService) BatchRemoveMember(ctx context.Context, req *pb.BatchRemoveMemberRequest) (*pb.BatchRemoveMemberReply, error) {
	return &pb.BatchRemoveMemberReply{}, nil
}
func (s *TenantService) BatchAddMember(ctx context.Context, req *pb.BatchAddMemberRequest) (*pb.BatchAddMemberReply, error) {
	return &pb.BatchAddMemberReply{}, nil
}
