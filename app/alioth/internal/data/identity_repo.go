package data

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/biz"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/captcha"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/emaillogincode"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/useroauth"
	"github.com/go-kratos/kratos/v2/log"
)

type identityRepo struct {
	client *ent.Client
	log    *log.Helper
}

func NewIdentityRepo(d *Data, logger log.Logger) biz.IdentityRepo {
	return &identityRepo{client: d.client, log: log.NewHelper(logger)}
}

func (r *identityRepo) fail(err error) error {
	if err == nil {
		return nil
	}
	if ent.IsNotFound(err) {
		return biz.ErrNotFound
	}
	if ent.IsConstraintError(err) {
		return biz.ErrConflict
	}
	r.log.Errorf("database: %v", err)
	return err
}

func (r *identityRepo) CreateUser(ctx context.Context, in biz.Account) (*biz.Account, error) {
	status := user.StatusACTIVE
	if in.Status == "BANNED" {
		status = user.StatusBANNED
	}
	row, err := r.client.User.Create().
		SetUID(in.UID).
		SetName(in.Name).
		SetNickname(in.Nickname).
		SetEmail(in.Email).
		SetAvatar(in.Avatar).
		SetRemark(in.Remark).
		SetStatus(status).
		SetPasswordHash(in.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *identityRepo) UserByEmail(ctx context.Context, email string) (*biz.Account, error) {
	row, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *identityRepo) UserByUID(ctx context.Context, uid string) (*biz.Account, error) {
	row, err := r.client.User.Query().Where(user.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *identityRepo) UpdateUserProfile(ctx context.Context, uid string, email, avatar, remark *string) (*biz.Account, error) {
	update := r.client.User.Update().Where(user.UIDEQ(uid))
	if email != nil {
		update.SetEmail(*email)
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
		return nil, biz.ErrNotFound
	}
	return r.UserByUID(ctx, uid)
}

func (r *identityRepo) UpdateUserStatus(ctx context.Context, uid, status string) (*biz.Account, error) {
	value := user.StatusACTIVE
	if status == "BANNED" {
		value = user.StatusBANNED
	} else if status != "ACTIVE" {
		return nil, biz.ErrInvalid
	}
	n, err := r.client.User.Update().Where(user.UIDEQ(uid)).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, biz.ErrNotFound
	}
	return r.UserByUID(ctx, uid)
}

func (r *identityRepo) ListUsers(ctx context.Context, page, pageSize int, name string) ([]biz.Account, int, error) {
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
	items := make([]biz.Account, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAccount(row))
	}
	return items, total, nil
}

func (r *identityRepo) SelectUsers(ctx context.Context, limit int, lastUID string) ([]biz.SelectItem, error) {
	query := r.client.User.Query()
	if lastUID != "" {
		query = query.Where(user.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(user.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]biz.SelectItem, 0, len(rows))
	for _, row := range rows {
		label := row.Name
		if label == "" {
			label = row.Email
		}
		items = append(items, biz.SelectItem{UID: row.UID, Name: label})
	}
	return items, nil
}

func (r *identityRepo) SaveCaptcha(ctx context.Context, id, answerHash string, expiresAt time.Time) error {
	_, err := r.client.Captcha.Create().
		SetUID(id).
		SetAnswerHash(answerHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return r.fail(err)
}

func (r *identityRepo) ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error) {
	row, err := r.client.Captcha.Query().Where(captcha.UIDEQ(id), captcha.UsedAtIsNil()).Only(ctx)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, r.fail(err)
	}
	if now.After(row.ExpiresAt) || subtle.ConstantTimeCompare([]byte(row.AnswerHash), []byte(answerHash)) != 1 {
		return false, nil
	}
	if err := r.client.Captcha.UpdateOne(row).SetUsedAt(now).Exec(ctx); err != nil {
		return false, r.fail(err)
	}
	return true, nil
}

func (r *identityRepo) SaveEmailCode(ctx context.Context, email, codeHash string, expiresAt time.Time) error {
	now := time.Now()
	if _, err := r.client.EmailLoginCode.Update().
		Where(emaillogincode.EmailEQ(email), emaillogincode.UsedAtIsNil()).
		SetUsedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	_, err := r.client.EmailLoginCode.Create().
		SetUID(newID()).
		SetEmail(email).
		SetCodeHash(codeHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return r.fail(err)
}

func (r *identityRepo) ConsumeEmailCode(ctx context.Context, email, codeHash string, now time.Time) (bool, error) {
	rows, err := r.client.EmailLoginCode.Query().Where(
		emaillogincode.EmailEQ(email),
		emaillogincode.UsedAtIsNil(),
		emaillogincode.ExpiresAtGT(now),
	).All(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	for _, row := range rows {
		if subtle.ConstantTimeCompare([]byte(row.CodeHash), []byte(codeHash)) != 1 {
			continue
		}
		if err := r.client.EmailLoginCode.UpdateOne(row).SetUsedAt(now).Exec(ctx); err != nil {
			return false, r.fail(err)
		}
		return true, nil
	}
	return false, nil
}

func (r *identityRepo) OAuthAccount(ctx context.Context, provider, openID string) (*biz.Account, error) {
	value, err := oauthProvider(provider)
	if err != nil {
		return nil, err
	}
	link, err := r.client.UserOAuth.Query().Where(useroauth.ProviderEQ(value), useroauth.OpenIDEQ(openID)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.UserByUID(ctx, link.UserUID)
}

func (r *identityRepo) LinkOAuth(ctx context.Context, userUID, provider, openID, email string) error {
	value, err := oauthProvider(provider)
	if err != nil {
		return err
	}
	_, err = r.client.UserOAuth.Create().
		SetUID(newID()).
		SetUserUID(userUID).
		SetProvider(value).
		SetOpenID(openID).
		SetEmail(email).
		Save(ctx)
	return r.fail(err)
}

func oauthProvider(provider string) (useroauth.Provider, error) {
	switch provider {
	case "GITHUB":
		return useroauth.ProviderGITHUB, nil
	case "GITEE":
		return useroauth.ProviderGITEE, nil
	case "FEISHU":
		return useroauth.ProviderFEISHU, nil
	default:
		return "", biz.ErrInvalid
	}
}

func toAccount(row *ent.User) biz.Account {
	return biz.Account{
		UID:          row.UID,
		Name:         row.Name,
		Nickname:     row.Nickname,
		Email:        row.Email,
		Avatar:       row.Avatar,
		Remark:       row.Remark,
		Status:       string(row.Status),
		PasswordHash: row.PasswordHash,
	}
}
