package data

import (
	"context"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *oauthRepo) OAuthAccount(ctx context.Context, provider, openID string) (*bo.Account, error) {
	value, err := oauthProvider(provider)
	if err != nil {
		return nil, err
	}
	link, err := r.client.UserOAuth.Query().Where(useroauth.ProviderEQ(value), useroauth.OpenIDEQ(openID)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	row, err := r.client.User.Query().Where(user.UIDEQ(link.UserUID)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *oauthRepo) LinkOAuth(ctx context.Context, userUID, provider, openID, email string) error {
	value, err := oauthProvider(provider)
	if err != nil {
		return err
	}
	_, err = r.client.UserOAuth.Create().
		SetUID(id.New()).
		SetUserUID(userUID).
		SetProvider(value).
		SetOpenID(openID).
		SetEmail(email).
		Save(ctx)
	return r.fail(err)
}

// MoveOAuth 把一条第三方绑定改到另一个账号，同时把绑定邮箱同步成刚验证过的邮箱。
// 改绑只认原来的 user_uid：0 行说明绑定已经被并发请求改走或者根本不存在，
// 由用例层再查一次绑定归属来决定是幂等成功还是"已绑定到其他账号"。
func (r *oauthRepo) MoveOAuth(ctx context.Context, provider, openID, fromUserUID, toUserUID, email string) error {
	value, err := oauthProvider(provider)
	if err != nil {
		return err
	}
	n, err := r.client.UserOAuth.Update().
		Where(useroauth.ProviderEQ(value), useroauth.OpenIDEQ(openID), useroauth.UserUIDEQ(fromUserUID)).
		SetUserUID(toUserUID).
		SetEmail(email).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

// ListOAuthByUser 列出账号名下的全部第三方绑定。最近绑定的排在前面，
// 个人资料页按这个顺序渲染。
func (r *oauthRepo) ListOAuthByUser(ctx context.Context, userUID string) ([]bo.OAuthLink, error) {
	rows, err := r.client.UserOAuth.Query().
		Where(useroauth.UserUIDEQ(userUID)).
		Order(ent.Desc(useroauth.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	links := make([]bo.OAuthLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, toOAuthLink(row))
	}
	return links, nil
}

// OAuthByUserProvider 取账号在某个来源下最近绑定的一条身份。同一来源允许绑定多个
// 第三方身份，列表接口也按条展示，所以这里只返回最近的一行。
func (r *oauthRepo) OAuthByUserProvider(ctx context.Context, userUID, provider string) (*bo.OAuthLink, error) {
	value, err := oauthProvider(provider)
	if err != nil {
		return nil, err
	}
	row, err := r.client.UserOAuth.Query().
		Where(useroauth.UserUIDEQ(userUID), useroauth.ProviderEQ(value)).
		Order(ent.Desc(useroauth.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	link := toOAuthLink(row)
	return &link, nil
}

// UnlinkOAuth 删除账号名下的一条绑定。WHERE 同时带 uid 和 user_uid，
// 别人的绑定无论 uid 对不对都删不到；0 行由用例层映射成 NotFound。
func (r *oauthRepo) UnlinkOAuth(ctx context.Context, userUID, uid string) error {
	n, err := r.client.UserOAuth.Delete().
		Where(useroauth.UIDEQ(uid), useroauth.UserUIDEQ(userUID)).
		Exec(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func toOAuthLink(row *ent.UserOAuth) bo.OAuthLink {
	return bo.OAuthLink{
		UID:      row.UID,
		Provider: row.Provider.String(),
		OpenID:   row.OpenID,
		Email:    row.Email,
		BoundAt:  row.CreatedAt,
	}
}

func oauthProvider(provider string) (useroauth.Provider, error) {
	switch provider {
	case "GITHUB":
		return useroauth.ProviderGITHUB, nil
	case "GITEE":
		return useroauth.ProviderGITEE, nil
	case "FEISHU":
		return useroauth.ProviderFEISHU, nil
	case "DINGTALK":
		return useroauth.ProviderDINGTALK, nil
	default:
		return "", merr.Invalid("invalid request")
	}
}
