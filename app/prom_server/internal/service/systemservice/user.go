package systemservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/model/basescopes"

	"prometheus-manager/api"
	pb "prometheus-manager/api/system"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/helper/model/systemscopes"
	valueobj2 "prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/password"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
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

	userBo := &bo.UserBO{
		Username: req.GetUsername(),
		Password: newPassword,
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Nickname: req.GetNickname(),
		Gender:   valueobj2.Gender(req.GetGender()),
	}

	if err = s.userBiz.CheckNewUser(ctx, userBo); err != nil {
		return nil, err
	}

	userBo, err = s.userBiz.CreateUser(ctx, userBo)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserReply{Id: userBo.Id}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	userBo := &bo.UserBO{
		Id:       req.GetId(),
		Nickname: req.GetNickname(),
		Avatar:   req.GetAvatar(),
		Status:   valueobj2.Status(req.GetStatus()),
		Remark:   req.GetRemark(),
		Gender:   valueobj2.Gender(req.GetGender()),
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
		Detail: userBo.ToApiV1(),
	}, nil
}

func (s *UserService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []query.ScopeMethod{
		systemscopes.UserLike(req.GetKeyword()),
		basescopes.CreatedAtDesc(),
	}
	userBos, err := s.userBiz.GetUserList(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(userBos, func(userBo *bo.UserBO) *api.UserV1 {
		return userBo.ToApiV1()
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
	pgInfo := query.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []query.ScopeMethod{
		systemscopes.UserLike(req.GetKeyword()),
		basescopes.CreatedAtDesc(),
	}
	userBos, err := s.userBiz.GetUserList(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(userBos, func(userBo *bo.UserBO) *api.UserSelectV1 {
		return userBo.ToApiSelectV1()
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
		Id: userBo.Id,
	}, nil
}

func (s *UserService) EditUserStatus(ctx context.Context, req *pb.EditUserStatusRequest) (*pb.EditUserStatusReply, error) {
	if err := s.userBiz.UpdateUserStatusById(ctx, valueobj2.Status(req.GetStatus()), req.GetIds()); err != nil {
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
