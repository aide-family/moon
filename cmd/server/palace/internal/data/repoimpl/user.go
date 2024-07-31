package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
)

// NewUserRepository 创建用户仓库
func NewUserRepository(data *data.Data) repository.User {
	return &userRepositoryImpl{
		data: data,
	}
}

type userRepositoryImpl struct {
	data *data.Data
}

func (l *userRepositoryImpl) UpdateBaseByID(ctx context.Context, user *bo.UpdateUserBaseParams) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(user.ID)).UpdateSimple(
		userQuery.Nickname.Value(user.Nickname),
		userQuery.Gender.Value(user.Gender.GetValue()),
		userQuery.Remark.Value(user.Remark),
	)
	return err
}

func (l *userRepositoryImpl) UpdateByID(ctx context.Context, user *bo.UpdateUserParams) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(user.ID)).UpdateSimple(
		userQuery.Nickname.Value(user.Nickname),
		userQuery.Avatar.Value(user.Avatar),
		userQuery.Gender.Value(user.Gender.GetValue()),
		userQuery.Remark.Value(user.Remark),
	)
	return err
}

func (l *userRepositoryImpl) DeleteByID(ctx context.Context, id uint32) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(id)).Delete()
	return err
}

func (l *userRepositoryImpl) UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.WithContext(ctx).Where(userQuery.ID.In(ids...)).Update(userQuery.Status, status)
	return err
}

func (l *userRepositoryImpl) UpdatePassword(ctx context.Context, id uint32, password types.Password) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.Where(userQuery.ID.Eq(id)).
		UpdateSimple(
			userQuery.Password.Value(password.String()),
			userQuery.Salt.Value(password.GetSalt()),
		)
	return err
}

func (l *userRepositoryImpl) Create(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error) {
	userModel := createUserParamsToModel(ctx, user)
	userModel.WithContext(ctx)
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	if err := userQuery.WithContext(ctx).Create(userModel); !types.IsNil(err) {
		return nil, err
	}
	runtimecache.GetRuntimeCache().AppendUser(ctx, userModel)
	return userModel, nil
}

func (l *userRepositoryImpl) BatchCreate(ctx context.Context, users []*bo.CreateUserParams) error {
	userModels := types.SliceToWithFilter(users, func(item *bo.CreateUserParams) (*model.SysUser, bool) {
		if types.IsNil(item) || types.TextIsNull(item.Name) {
			return nil, false
		}
		return createUserParamsToModel(ctx, item), true
	})
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).CreateInBatches(userModels, 10)
}

func (l *userRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysUser, error) {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.ID.Eq(id)).First()
}

func (l *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.Username.Eq(username)).First()
}

func (l *userRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryUserListParams) ([]*model.SysUser, error) {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	q := userQuery.WithContext(ctx)

	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, userQuery.Status.Eq(params.Status.GetValue()))
	}
	if !params.Gender.IsUnknown() {
		wheres = append(wheres, userQuery.Gender.Eq(params.Gender.GetValue()))
	}
	if !params.Role.IsAll() {
		wheres = append(wheres, userQuery.Role.Eq(params.Role.GetValue()))
	}
	if !types.TextIsNull(params.Keyword) {
		q = q.Or(userQuery.Username.Like(params.Keyword))
		q = q.Or(userQuery.Nickname.Like(params.Keyword))
		q = q.Or(userQuery.Email.Like(params.Keyword))
		q = q.Or(userQuery.Phone.Like(params.Keyword))
		q = q.Or(userQuery.Remark.Like(params.Keyword))
	}

	q = q.Where(wheres...)
	if err := types.WithPageQuery[query.ISysUserDo](q, params.Page); err != nil {
		return nil, err
	}
	return q.Order(userQuery.ID.Desc()).Find()
}

func (l *userRepositoryImpl) UpdateUser(ctx context.Context, user *model.SysUser) error {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	_, err := userQuery.WithContext(ctx).Where(userQuery.ID.Eq(user.ID)).Updates(user)
	return err
}

func (l *userRepositoryImpl) FindByIds(ctx context.Context, ids ...uint32) ([]*model.SysUser, error) {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.ID.In(ids...)).Find()
}

// createUserParamsToModel create user params to model
func createUserParamsToModel(ctx context.Context, user *bo.CreateUserParams) *model.SysUser {
	if types.IsNil(user) {
		return nil
	}
	userItem := &model.SysUser{
		Username: user.Name,
		Nickname: user.Nickname,
		Password: user.Password.String(),
		Email:    user.Email,
		Phone:    user.Phone,
		Status:   user.Status,
		Remark:   user.Remark,
		Avatar:   user.Avatar,
		Salt:     user.Password.GetSalt(),
		Gender:   user.Gender,
	}
	userItem.WithContext(ctx)
	return userItem
}
