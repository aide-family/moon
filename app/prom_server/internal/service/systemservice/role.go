package systemservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/model/systemscopes"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type RoleService struct {
	pb.UnimplementedRoleServer
	log *log.Helper

	roleBiz *biz.RoleBiz
}

func NewRoleService(roleBiz *biz.RoleBiz, logger log.Logger) *RoleService {
	return &RoleService{
		log:     log.NewHelper(log.With(logger, "module", "service.role")),
		roleBiz: roleBiz,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleReply, error) {
	b := &bo.RoleBO{
		Name:   req.GetName(),
		Remark: req.GetRemark(),
	}
	b, err := s.roleBiz.CreateRole(ctx, b)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRoleReply{
		Id: uint32(b.Id),
	}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleReply, error) {
	b := &bo.RoleBO{
		Id:     uint(req.GetId()),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Status: valueobj.Status(req.GetStatus()),
	}
	b, err := s.roleBiz.UpdateRoleById(ctx, b)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateRoleReply{Id: uint32(b.Id)}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleReply, error) {
	if err := s.roleBiz.DeleteRoleByIds(ctx, []uint32{req.GetId()}); err != nil {
		return nil, err
	}
	return &pb.DeleteRoleReply{
		Id: req.GetId(),
	}, nil
}

func (s *RoleService) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.GetRoleReply, error) {
	roleBO, err := s.roleBiz.GetRoleById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetRoleReply{
		Detail: roleBO.ToApiV1(),
	}, nil
}

func (s *RoleService) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		systemscopes.RoleLike(req.GetKeyword()),
	}

	boList, err := s.roleBiz.ListRole(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(boList, func(t *bo.RoleBO) *api.RoleV1 {
		return t.ToApiV1()
	})
	return &pb.ListRoleReply{
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *RoleService) SelectRole(ctx context.Context, req *pb.SelectRoleRequest) (*pb.SelectRoleReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		systemscopes.RoleLike(req.GetKeyword()),
	}

	boList, err := s.roleBiz.ListRole(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := slices.To(boList, func(t *bo.RoleBO) *api.RoleSelectV1 {
		return t.ApiRoleSelectV1()
	})
	return &pb.SelectRoleReply{
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *RoleService) RelateApi(ctx context.Context, req *pb.RelateApiRequest) (*pb.RelateApiReply, error) {
	if err := s.roleBiz.RelateApiById(ctx, req.GetId(), req.GetApiIds()); err != nil {
		return nil, err
	}
	return &pb.RelateApiReply{
		Id: req.GetId(),
	}, nil
}

// EditRoleStatus 编辑角色状态
func (s *RoleService) EditRoleStatus(ctx context.Context, req *pb.EditRoleStatusRequest) (*pb.EditRoleStatusReply, error) {
	if err := s.roleBiz.UpdateRoleStatusById(ctx, valueobj.Status(req.GetStatus()), req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.EditRoleStatusReply{
		Ids: req.GetIds(),
	}, nil
}
