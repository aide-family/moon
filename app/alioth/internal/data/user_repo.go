package data

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *userRepo) CreateUser(ctx context.Context, in bo.Account) (*bo.Account, error) {
	status := user.StatusACTIVE
	if in.Status == "BANNED" {
		status = user.StatusBANNED
	}
	count, err := r.client.User.Query().Count(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	systemRole := user.SystemRoleUSER
	if count == 0 {
		systemRole = user.SystemRoleSUPER_ADMIN
	}
	create := func(role user.SystemRole) (*ent.User, error) {
		return r.client.User.Create().
			SetUID(in.UID).
			SetName(in.Name).
			SetNickname(in.Nickname).
			SetEmail(in.Email).
			SetAvatar(in.Avatar).
			SetPhone(in.Phone).
			SetRemark(in.Remark).
			SetStatus(status).
			SetSystemRole(role).
			Save(ctx)
	}
	row, err := create(systemRole)
	if err != nil && systemRole == user.SystemRoleSUPER_ADMIN {
		// 并发注册时另一个请求已经拿到唯一超级管理员，当前用户按普通用户创建。
		if ent.IsConstraintError(err) {
			row, err = create(user.SystemRoleUSER)
		}
	}
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UserByEmail(ctx context.Context, email string) (*bo.Account, error) {
	row, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UserByUID(ctx context.Context, uid string) (*bo.Account, error) {
	row, err := r.client.User.Query().Where(user.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UpdateUserProfile(ctx context.Context, uid string, email, phone, avatar, remark *string) (*bo.Account, error) {
	update := r.client.User.Update().Where(user.UIDEQ(uid))
	if email != nil {
		update.SetEmail(*email)
	}
	if phone != nil {
		update.SetPhone(*phone)
	}
	if avatar != nil {
		update.SetAvatar(*avatar)
	}
	if remark != nil {
		update.SetRemark(*remark)
	}
	n, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) UpdateUserStatus(ctx context.Context, uid, status string) (*bo.Account, error) {
	value := user.StatusACTIVE
	if status == "BANNED" {
		value = user.StatusBANNED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.User.Update().Where(user.UIDEQ(uid)).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) UpdateUserSystemRole(ctx context.Context, uid, role, authorizedBy string) (*bo.Account, error) {
	value := user.SystemRole(role)
	switch value {
	case user.SystemRoleSYSTEM_ADMIN, user.SystemRoleUSER:
	default:
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.User.Update().
		Where(user.UIDEQ(uid)).
		SetSystemRole(value).
		SetAuthorizedBy(authorizedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) ListUsers(ctx context.Context, page, pageSize int, name string) ([]bo.Account, int, error) {
	query := r.client.User.Query()
	if name != "" {
		query = query.Where(user.Or(user.NameContainsFold(name), user.EmailContainsFold(name)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(user.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Account, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAccount(row))
	}
	return items, total, nil
}

func (r *userRepo) SelectUsers(ctx context.Context, limit int, lastUID string) ([]bo.SelectItem, error) {
	query := r.client.User.Query()
	if lastUID != "" {
		query = query.Where(user.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(user.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		label := row.Name
		if label == "" {
			label = row.Email
		}
		items = append(items, bo.SelectItem{UID: row.UID, Name: label})
	}
	return items, nil
}

func toAccount(row *ent.User) bo.Account {
	return bo.Account{
		UID:          row.UID,
		Name:         row.Name,
		Nickname:     row.Nickname,
		Email:        row.Email,
		Phone:        row.Phone,
		Avatar:       row.Avatar,
		Remark:       row.Remark,
		Status:       string(row.Status),
		SystemRole:   string(row.SystemRole),
		AuthorizedBy: row.AuthorizedBy,
	}
}
