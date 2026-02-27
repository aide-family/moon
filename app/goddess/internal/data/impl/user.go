package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

func NewUserRepository(d *data.Data) repository.User {
	return &userRepository{Data: d}
}

type userRepository struct {
	*data.Data
}

func (u *userRepository) GetUser(ctx context.Context, uid snowflake.ID) (*bo.UserItemBo, error) {
	mutation := query.User
	user, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("user %s not found", uid)
		}
		return nil, err
	}
	return convert.UserToBo(user), nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*bo.UserItemBo, error) {
	mutation := query.User
	user, err := mutation.WithContext(ctx).Where(mutation.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("user %s not found", email)
		}
		return nil, err
	}
	return convert.UserToBo(user), nil
}

func (u *userRepository) ListUser(ctx context.Context, req *bo.ListUserBo) (*bo.PageResponseBo[*bo.UserItemBo], error) {
	mutation := query.User
	wrappers := mutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Email) {
		wrappers = wrappers.Where(mutation.Email.Like("%" + req.Email + "%"))
	}
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.Email.Like(keyword)).
			Or(mutation.Name.Like(keyword)).
			Or(mutation.Nickname.Like(keyword))
	}
	if req.Status > enum.UserStatus_UserStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(uint8(req.Status)))
	}
	if req.Page > 0 && req.PageSize > 0 {
		countWrappers := wrappers
		total, err := countWrappers.Count()
		if err != nil {
			return nil, merr.ErrorInternalServer("list user failed: %v", err)
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))
	}
	users, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("list user failed: %v", err)
	}
	items := make([]*bo.UserItemBo, 0, len(users))
	for _, user := range users {
		items = append(items, convert.UserToBo(user))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (u *userRepository) SelectUser(ctx context.Context, req *bo.SelectUserBo) (*bo.SelectUserBoResult, error) {
	mutation := query.User
	wrappers := mutation.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		keyword := "%" + req.Keyword + "%"
		wrappers = wrappers.Where(mutation.Email.Like(keyword)).
			Or(mutation.Name.Like(keyword)).
			Or(mutation.Nickname.Like(keyword))
	}
	if req.Status > enum.UserStatus_UserStatus_UNKNOWN {
		wrappers = wrappers.Where(mutation.Status.Eq(uint8(req.Status)))
	}
	if req.LastUID > 0 {
		wrappers = wrappers.Where(mutation.UID.Lt(req.LastUID.Int64()))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, merr.ErrorInternalServer("select user failed: %v", err)
	}
	wrappers = wrappers.Limit(int(req.Limit)).Order(mutation.UID.Desc())
	users, err := wrappers.Find()
	if err != nil {
		return nil, merr.ErrorInternalServer("select user failed: %v", err)
	}
	items := make([]*bo.SelectUserItemBo, 0, len(users))
	for _, user := range users {
		items = append(items, &bo.SelectUserItemBo{
			Value:    user.UID,
			Label:    user.Nickname,
			Disabled: enum.GlobalStatus(user.Status) != enum.GlobalStatus_ENABLED,
			Tooltip:  user.Email,
		})
	}
	var lastUID snowflake.ID
	if len(users) > 0 {
		lastUID = users[len(users)-1].UID
	}
	return &bo.SelectUserBoResult{
		Items:   items,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(items) >= int(req.Limit),
	}, nil
}

func (u *userRepository) UpdateUserStatus(ctx context.Context, uid snowflake.ID, status int32) error {
	mutation := query.User
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Update(mutation.Status, uint8(status))
	return err
}

func (u *userRepository) UpdateUserEmail(ctx context.Context, uid snowflake.ID, email string) error {
	mutation := query.User
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Update(mutation.Email, email)
	return err
}

func (u *userRepository) UpdateUserAvatar(ctx context.Context, uid snowflake.ID, avatar string) error {
	mutation := query.User
	_, err := mutation.WithContext(ctx).Where(mutation.UID.Eq(uid.Int64())).Update(mutation.Avatar, avatar)
	return err
}
