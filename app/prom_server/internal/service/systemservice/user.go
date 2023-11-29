package systemservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/helper/model/system"
	"prometheus-manager/pkg/util/password"
	"prometheus-manager/pkg/util/slices"
)

type UserService struct {
	pb.UnimplementedUserServer
	log *log.Helper

	userBiz *biz.UserBiz
}

func NewUserService(userBiz *biz.UserBiz, logger log.Logger) *UserService {
	return &UserService{
		log:     log.NewHelper(log.With(logger, "module", "service.user")),
		userBiz: userBiz,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	newPassword, err := password.DecryptPassword(req.GetPassword(), password.DefaultIv)
	if err != nil {
		return nil, err
	}

	userBo := &dobo.UserBO{
		Username: req.GetUsername(),
		Password: newPassword,
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
	}
	userBo, err = s.userBiz.CreateUser(ctx, userBo)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserReply{Id: uint32(userBo.Id)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	userBo := &dobo.UserBO{
		Id:       uint(req.GetId()),
		Username: req.GetUsername(),
		Avatar:   req.GetAvatar(),
		Status:   valueobj.Status(req.GetStatus()),
		Remark:   req.GetRemark(),
	}
	userBo, err := s.userBiz.UpdateUserById(ctx, req.GetId(), userBo)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateUserReply{Id: req.GetId()}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	if err := s.userBiz.DeleteUserByIds(ctx, []uint32{req.GetId()}); err != nil {
		return nil, err
	}
	return &pb.DeleteUserReply{Id: req.GetId()}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	userBo, err := s.userBiz.GetUserInfoById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetUserReply{
		Detail: userBo.ApiUserV1(),
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		system.UserLike(req.GetKeyword()),
	}
	userBos, err := s.userBiz.GetUserList(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(userBos, func(userBo *dobo.UserBO) *api.UserV1 {
		return userBo.ApiUserV1()
	})
	return &pb.ListUserReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *UserService) SelectUser(ctx context.Context, req *pb.SelectUserRequest) (*pb.SelectUserReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		system.UserLike(req.GetKeyword()),
	}
	userBos, err := s.userBiz.GetUserList(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(userBos, func(userBo *dobo.UserBO) *api.UserSelectV1 {
		return userBo.ApiUserSelectV1()
	})
	return &pb.SelectUserReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}

func (s *UserService) EditUserPassword(ctx context.Context, req *pb.EditUserPasswordRequest) (*pb.EditUserPasswordReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}

	oldPassword, err := password.DecryptPassword(req.GetOldPassword(), password.DefaultIv)
	if err != nil {
		return nil, err
	}
	newPassword, err := password.DecryptPassword(req.GetNewPassword(), password.DefaultIv)
	if err != nil {
		return nil, err
	}

	userBo, err := s.userBiz.EditUserPassword(ctx, authClaims, oldPassword, newPassword)
	if err != nil {
		return nil, err
	}
	return &pb.EditUserPasswordReply{
		Id: uint32(userBo.Id),
	}, nil
}

func (s *UserService) EditUserStatus(ctx context.Context, req *pb.EditUserStatusRequest) (*pb.EditUserStatusReply, error) {
	if err := s.userBiz.UpdateUserStatusById(ctx, valueobj.Status(req.GetStatus()), req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.EditUserStatusReply{Ids: req.GetIds()}, nil
}

func (s *UserService) RelateRoles(ctx context.Context, req *pb.RelateRolesRequest) (*pb.RelateRolesReply, error) {
	if err := s.userBiz.RelateRoles(ctx, req.GetId(), req.GetRoleIds()); err != nil {
		return nil, err
	}
	return &pb.RelateRolesReply{Id: req.GetId()}, nil
}
