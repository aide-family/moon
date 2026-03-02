// Package authv1 is the repository for the auth service.
package authv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/data/impl"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	RegisterAuthV1Factory(config.DomainConfig_GORM, NewAuthRepository)
}

func NewAuthRepository(c *config.DomainConfig, jwtConfig *config.JWT) (goddessv1.AuthServiceServer, func() error, error) {
	ormConfig := &config.ORMConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), ormConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal orm config failed: %v", err)
		}
	}
	db, close, err := connect.NewDB(ormConfig)
	if err != nil {
		return nil, nil, err
	}
	loginRepo := impl.NewLoginRepositoryWithDB(db, jwtConfig)
	loginBiz := biz.NewLoginBiz(loginRepo)
	return &authRepository{AuthService: service.NewAuthService(loginBiz)}, close, nil
}

type authRepository struct {
	*service.AuthService
}
