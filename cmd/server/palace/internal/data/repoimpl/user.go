package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/helper"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"gorm.io/gen"
)

// NewUserRepository 创建用户仓库
func NewUserRepository(bc *palaceconf.Bootstrap, data *data.Data, cacheRepo repository.Cache) repository.User {
	return &userRepositoryImpl{
		bc:        bc,
		data:      data,
		cacheRepo: cacheRepo,
	}
}

type userRepositoryImpl struct {
	bc        *palaceconf.Bootstrap
	data      *data.Data
	cacheRepo repository.Cache
}

func (l *userRepositoryImpl) GetUserByEmailOrPhone(ctx context.Context, emailOrPhone string) (*model.SysUser, error) {
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	userCtxQuery := userQuery.WithContext(ctx).Or(userQuery.Phone.Eq(emailOrPhone)).Or(userQuery.Email.Eq(emailOrPhone))
	return userCtxQuery.First()
}

func (l *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.SysUser, error) {
	if err := helper.CheckEmail(email); err != nil {
		return nil, err
	}
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.Email.Eq(email)).First()
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
	// 查询用户信息
	user, err := l.GetByID(ctx, id)
	if !types.IsNil(err) {
		return err
	}

	return l.data.GetMainDB(ctx).Transaction(func(tx *gorm.DB) error {
		userQuery := query.Use(tx).SysUser
		_, err = userQuery.Where(userQuery.ID.Eq(id)).
			UpdateSimple(
				userQuery.Password.Value(password.GetValue()),
				userQuery.Salt.Value(password.GetSalt()),
			)
		if err != nil {
			return err
		}
		return sendUserPassword(l.data.GetEmail(), l.bc, user, password.GetPlaintext())
	})
}

func (l *userRepositoryImpl) Create(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error) {
	// 根据email查询用户，如果存在，则返回错误
	userQuery := query.Use(l.data.GetMainDB(ctx)).SysUser
	if sysUser, err := userQuery.WithContext(ctx).Where(userQuery.Email.Eq(user.Email)).First(); !types.IsNil(err) {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return sysUser, nil
	}
	userModel := createUserParamsToModel(ctx, user)
	userModel.WithContext(ctx)

	if err := userQuery.WithContext(ctx).Create(userModel); !types.IsNil(err) {
		return nil, err
	}
	l.cacheRepo.AppendUser(ctx, userModel)
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
	userCtxQuery := userQuery.WithContext(ctx)

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
		userCtxQuery = userCtxQuery.Or(userQuery.Username.Like(params.Keyword))
		userCtxQuery = userCtxQuery.Or(userQuery.Nickname.Like(params.Keyword))
		userCtxQuery = userCtxQuery.Or(userQuery.Email.Like(params.Keyword))
		userCtxQuery = userCtxQuery.Or(userQuery.Phone.Like(params.Keyword))
		userCtxQuery = userCtxQuery.Or(userQuery.Remark.Like(params.Keyword))
	}
	if len(params.IDs) > 0 {
		userCtxQuery = userCtxQuery.Or(userQuery.ID.In(params.IDs...))
	}

	userCtxQuery = userCtxQuery.Where(wheres...)
	var err error
	if userCtxQuery, err = types.WithPageQuery(userCtxQuery, params.Page); err != nil {
		return nil, err
	}
	return userCtxQuery.Order(userQuery.ID.Desc()).Find()
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
		Role:     user.Role,
	}
	userItem.WithContext(ctx)
	return userItem
}
