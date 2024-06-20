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

func NewUserRepository(data *data.Data) repository.User {
	return &userRepositoryImpl{
		data: data,
	}
}

type userRepositoryImpl struct {
	data *data.Data
}

func (l *userRepositoryImpl) UpdateByID(ctx context.Context, user *bo.UpdateUserParams) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).SysUser.WithContext(ctx).Where(query.SysUser.ID.Eq(user.ID)).UpdateSimple(
		query.SysUser.Nickname.Value(user.Nickname),
		query.SysUser.Avatar.Value(user.Avatar),
		query.SysUser.Gender.Value(user.Gender.GetValue()),
		query.SysUser.Remark.Value(user.Remark),
	)
	return err
}

func (l *userRepositoryImpl) DeleteByID(ctx context.Context, id uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(query.SysUser.ID.Eq(id)).Delete()
	return err
}

func (l *userRepositoryImpl) UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(query.SysUser.ID.In(ids...)).Update(query.SysUser.Status, status)
	return err
}

func (l *userRepositoryImpl) UpdatePassword(ctx context.Context, id uint32, password types.Password) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(query.SysUser.ID.Eq(id)).
		UpdateSimple(
			query.SysUser.Password.Value(password.String()),
			query.SysUser.Salt.Value(password.GetSalt()),
		)
	return err
}

func (l *userRepositoryImpl) Create(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error) {
	userModel := createUserParamsToModel(user)
	if err := userModel.Create(ctx, l.data.GetMainDB(ctx)); !types.IsNil(err) {
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
		return createUserParamsToModel(item), true
	})
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.CreateInBatches(userModels, 10)
}

func (l *userRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysUser, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysUser.WithContext(ctx).Where(query.SysUser.ID.Eq(id)).First()
}

func (l *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysUser.WithContext(ctx).Where(query.SysUser.Username.Eq(username)).First()
}

func (l *userRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryUserListParams) ([]*model.SysUser, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysUser.WithContext(ctx)

	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, query.SysUser.Status.Eq(params.Status.GetValue()))
	}
	if !params.Gender.IsUnknown() {
		wheres = append(wheres, query.SysUser.Gender.Eq(params.Gender.GetValue()))
	}
	if !params.Role.IsAll() {
		wheres = append(wheres, query.SysUser.Role.Eq(params.Role.GetValue()))
	}
	if !types.TextIsNull(params.Keyword) {
		q = q.Or(
			query.SysUser.Username.Like(params.Keyword),
			query.SysUser.Nickname.Like(params.Keyword),
			query.SysUser.Email.Like(params.Keyword),
			query.SysUser.Phone.Like(params.Keyword),
			query.SysUser.Remark.Like(params.Keyword),
		)
	}

	q = q.Where(wheres...)
	if !types.IsNil(params) {
		page := params.Page
		total, err := q.Count()
		if !types.IsNil(err) {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Order(query.SysUser.ID.Desc()).Find()
}

func (l *userRepositoryImpl) UpdateUser(ctx context.Context, user *model.SysUser) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(query.SysUser.ID.Eq(user.ID)).Updates(user)
	return err
}

func (l *userRepositoryImpl) FindByIds(ctx context.Context, ids ...uint32) ([]*model.SysUser, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysUser.WithContext(ctx).Where(query.SysUser.ID.In(ids...)).Find()
}

// createUserParamsToModel create user params to model
func createUserParamsToModel(user *bo.CreateUserParams) *model.SysUser {
	if types.IsNil(user) {
		return nil
	}
	return &model.SysUser{
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
}
