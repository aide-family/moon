package systemservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/api"
	pb "prometheus-manager/api/server/system"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
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
		Id: b.Id,
	}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleReply, error) {
	b := &bo.RoleBO{
		Id:     req.GetId(),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Status: vo.Status(req.GetStatus()),
	}
	b, err := s.roleBiz.UpdateRoleById(ctx, b)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateRoleReply{Id: b.Id}, nil
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
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.UpdateAtDesc(),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
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
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *RoleService) SelectRole(ctx context.Context, req *pb.SelectRoleRequest) (*pb.SelectRoleReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.UpdateAtDesc(),
		basescopes.StatusEQ(vo.StatusEnabled),
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
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
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
	if err := s.roleBiz.UpdateRoleStatusById(ctx, vo.Status(req.GetStatus()), req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.EditRoleStatusReply{
		Ids: req.GetIds(),
	}, nil
}
