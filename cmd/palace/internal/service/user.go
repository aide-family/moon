package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
)

// UserService is a user service implementation.
type UserService struct {
	palace.UnimplementedUserServer

	userBiz    *biz.UserBiz
	teamBiz    *biz.Team
	messageBiz *biz.Message
	log        *log.Helper
}

// NewUserService creates a new user service.
func NewUserService(
	userBiz *biz.UserBiz,
	teamBiz *biz.Team,
	messageBiz *biz.Message,
	logger log.Logger,
) *UserService {
	return &UserService{
		userBiz:    userBiz,
		teamBiz:    teamBiz,
		messageBiz: messageBiz,
		log:        log.NewHelper(log.With(logger, "module", "service.user")),
	}
}

// SelfInfo retrieves the current user's information.
func (s *UserService) SelfInfo(ctx context.Context, _ *common.EmptyRequest) (*common.UserItem, error) {
	user, err := s.userBiz.GetSelfInfo(ctx)
	if err != nil {
		return nil, err
	}

	return build.ToUserItem(user), nil
}

// UpdateSelfInfo updates the current user's information.
func (s *UserService) UpdateSelfInfo(ctx context.Context, req *palace.UpdateSelfInfoRequest) (*common.EmptyReply, error) {
	// Call business logic
	if err := s.userBiz.UpdateSelfInfo(ctx, build.ToSelfUpdateInfo(req)); err != nil {
		return nil, err
	}

	return &common.EmptyReply{Message: "修改成功"}, nil
}

// UpdateSelfPassword updates the current user's password.
func (s *UserService) UpdateSelfPassword(ctx context.Context, req *palace.UpdateSelfPasswordRequest) (*common.EmptyReply, error) {
	// Call business logic to update password
	if err := s.userBiz.UpdateSelfPassword(ctx, build.ToPasswordUpdateInfo(req, s.messageBiz.SendEmail)); err != nil {
		return nil, err
	}

	return &common.EmptyReply{Message: "修改密码成功"}, nil
}

// LeaveTeam allows the current user to leave a team.
func (s *UserService) LeaveTeam(ctx context.Context, req *palace.LeaveTeamRequest) (*common.EmptyReply, error) {
	// TODO: implement the logic
	return &common.EmptyReply{}, nil
}

// JoinTeam allows the current user to join a team.
func (s *UserService) JoinTeam(ctx context.Context, req *palace.JoinTeamRequest) (*common.EmptyReply, error) {
	// TODO: implement the logic
	return &common.EmptyReply{}, nil
}

// CreateTeam allows the current user to create a new team.
func (s *UserService) CreateTeam(ctx context.Context, req *palace.CreateTeamRequest) (*common.EmptyReply, error) {
	if err := s.teamBiz.SaveTeam(ctx, build.ToSaveOneTeamRequestByCreate(req)); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "创建团队成功"}, nil
}

// SelfTeamList retrieves the list of teams the current user is a member of.
func (s *UserService) SelfTeamList(ctx context.Context, _ *common.EmptyRequest) (*palace.SelfTeamListReply, error) {
	// Call business logic to get user's teams
	teams, err := s.userBiz.GetUserTeams(ctx)
	if err != nil {
		return nil, err
	}

	// 使用转换方法将领域对象转换为proto对象
	return &palace.SelfTeamListReply{
		Items: build.ToTeamItems(teams),
	}, nil
}

// SelfSubscribeTeamStrategies retrieves the list of team strategies the current user is subscribed to.
func (s *UserService) SelfSubscribeTeamStrategies(ctx context.Context, req *palace.SelfSubscribeTeamStrategiesRequest) (*palace.SelfSubscribeTeamStrategiesReply, error) {
	// TODO: implement the logic
	return &palace.SelfSubscribeTeamStrategiesReply{}, nil
}
