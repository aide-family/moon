package gormimpl

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	authv1 "github.com/aide-family/magicbox/domain/auth/v1"
	"github.com/aide-family/magicbox/domain/auth/v1/gormimpl/query"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
)

func init() {
	authv1.RegisterUserFactoryV1(config.DomainConfig_GORM, NewUserGormRepository)
}

func NewUserGormRepository(c *config.DomainConfig) (apiv1.UserServer, func() error, error) {
	ormConfig := &config.ORMConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), ormConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal orm config failed: %v", err)
		}
	}
	db, close, err := connect.NewDB(ormConfig)
	if err != nil {
		return nil, nil, err
	}
	query.SetDefault(db)
	return &gormUserRepository{repoConfig: c, db: db}, close, nil
}

type gormUserRepository struct {
	apiv1.UnimplementedUserServer
	repoConfig *config.DomainConfig
	db         *gorm.DB
}

// BanUser implements [v1.UserServer].
func (g *gormUserRepository) BanUser(ctx context.Context, req *apiv1.BanUserRequest) (*apiv1.BanUserReply, error) {
	if err := g.updateUserStatus(ctx, req.GetUid(), enum.UserStatus_BANNED); err != nil {
		return nil, err
	}
	return &apiv1.BanUserReply{
		Message: "user banned successfully",
	}, nil
}

// PermitUser implements [v1.UserServer].
func (g *gormUserRepository) PermitUser(ctx context.Context, req *apiv1.PermitUserRequest) (*apiv1.PermitUserReply, error) {
	if err := g.updateUserStatus(ctx, req.GetUid(), enum.UserStatus_ACTIVE); err != nil {
		return nil, err
	}
	return &apiv1.PermitUserReply{
		Message: "user permitted successfully",
	}, nil
}

// UpdateUserStatus implements [v1.UserServer].
func (g *gormUserRepository) updateUserStatus(ctx context.Context, uid int64, status enum.UserStatus) error {
	_, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(uid)).UpdateColumnSimple(query.User.Status.Value(int32(status)))
	if err != nil {
		return merr.ErrorInternalServer("update user status failed: %v", err)
	}
	return nil
}

// GetUser implements [v1.UserServer].
func (g *gormUserRepository) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.UserItem, error) {
	u, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(req.GetUid())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("user not found")
		}
		return nil, merr.ErrorInternalServer("get user failed: %v", err)
	}
	return convertUserItem(u), nil
}

// ListUser implements [v1.UserServer].
func (g *gormUserRepository) ListUser(ctx context.Context, req *apiv1.ListUserRequest) (*apiv1.ListUserReply, error) {
	mut := query.User
	wrappers := mut.WithContext(ctx)
	if strutil.IsNotEmpty(req.GetKeyword()) {
		kw := "%" + req.GetKeyword() + "%"
		wrappers = wrappers.Or(mut.Name.Like(kw), mut.Nickname.Like(kw), mut.Email.Like(kw), mut.Remark.Like(kw))
	}
	if req.GetStatus() > enum.UserStatus_UserStatus_UNKNOWN {
		wrappers = wrappers.Where(mut.Status.Eq(int32(req.GetStatus())))
	}
	if strutil.IsNotEmpty(req.GetEmail()) {
		wrappers = wrappers.Where(mut.Email.Eq(req.GetEmail()))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count user failed: %v", err)
	}
	if req.GetPage() > 0 && req.GetPageSize() > 0 {
		wrappers = wrappers.Limit(int(req.GetPageSize())).Offset(int(req.GetPageSize() * (req.GetPage() - 1)))
	}
	wrappers = wrappers.Order(mut.ID.Desc())
	list, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list user failed: %v", err)
	}
	items := make([]*apiv1.UserItem, 0, len(list))
	for _, u := range list {
		items = append(items, convertUserItem(u))
	}
	return &apiv1.ListUserReply{
		Items:    items,
		Total:    total,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// SelectUser implements [v1.UserServer].
func (g *gormUserRepository) SelectUser(ctx context.Context, req *apiv1.SelectUserRequest) (*apiv1.SelectUserReply, error) {
	mut := query.User
	wrappers := mut.WithContext(ctx)
	if strutil.IsNotEmpty(req.GetKeyword()) {
		kw := "%" + req.GetKeyword() + "%"
		wrappers = wrappers.Or(mut.Name.Like(kw), mut.Nickname.Like(kw), mut.Email.Like(kw), mut.Remark.Like(kw))
	}
	if req.GetStatus() > enum.UserStatus_UserStatus_UNKNOWN {
		wrappers = wrappers.Where(mut.Status.Eq(int32(req.GetStatus())))
	}
	if req.GetLastUID() > 0 {
		wrappers = wrappers.Where(mut.ID.Lt(req.GetLastUID()))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("count user failed: %v", err)
	}
	wrappers = wrappers.Limit(int(req.GetLimit())).Order(mut.ID.Desc())
	list, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select user failed: %v", err)
	}
	items := make([]*apiv1.SelectUserItem, 0, len(list))
	var lastUID int64
	for _, u := range list {
		items = append(items, convertUserSelectItem(u))
		lastUID = u.ID.Int64()
	}
	return &apiv1.SelectUserReply{
		Items:   items,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(list) == int(req.GetLimit()),
	}, nil
}
