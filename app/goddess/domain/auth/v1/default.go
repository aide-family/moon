package authv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	authdomain "github.com/aide-family/goddess/domain/auth"
	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/data/impl"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	authdomain.RegisterAuthV1Factory(config.DomainConfig_DEFAULT, NewDefaultAuth)
}

// NewDefaultAuth creates an in-process auth server (DEFAULT driver).
func NewDefaultAuth(c *config.DomainConfig, driver *anypb.Any) (goddessv1.AuthServiceServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal default config failed: %v", err)
		}
	}
	db, close, err := connect.NewDB(defaultConfig.GetDatabase())
	if err != nil {
		return nil, nil, err
	}
	bootstrap := &conf.Bootstrap{
		GlobalEmail: defaultConfig.GetGlobalEmail(),
		SiteDomain:  defaultConfig.GetSiteDomain(),
		Jwt:         defaultConfig.GetJwt(),
	}
	transaction := impl.NewTransactionWithDB(db)
	loginRepo := impl.NewLoginRepositoryWithDB(db, bootstrap.GetJwt())
	emailRepo := impl.NewEmailRepository(bootstrap)
	emailBiz := biz.NewEmail(emailRepo)
	captchaBiz := biz.NewCaptcha()
	loginBiz := biz.NewLoginBiz(transaction, loginRepo)
	return &defaultAuth{AuthService: service.NewAuthService(loginBiz, emailBiz, captchaBiz)}, close, nil
}

type defaultAuth struct {
	*service.AuthService
}
