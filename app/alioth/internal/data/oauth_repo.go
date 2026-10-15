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

func oauthProvider(provider string) (useroauth.Provider, error) {
	switch provider {
	case "GITHUB":
		return useroauth.ProviderGITHUB, nil
	case "GITEE":
		return useroauth.ProviderGITEE, nil
	case "FEISHU":
		return useroauth.ProviderFEISHU, nil
	default:
		return "", merr.Invalid("invalid request")
	}
}
