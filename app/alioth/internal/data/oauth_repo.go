package data

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
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
