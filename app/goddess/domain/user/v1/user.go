package userv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/data/impl"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	RegisterUserFactoryV1(config.DomainConfig_GORM, NewUserRepository)
}

func NewUserRepository(c *config.DomainConfig) (goddessv1.UserServer, func() error, error) {
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
	helper := klog.NewHelper(klog.With(klog.GetLogger(), "module", "user"))
	userRepo := impl.NewUserRepositoryWithDB(db)
	userBiz := biz.NewUser(userRepo, helper)
	return &userRepository{UserServer: service.NewUserService(userBiz)}, close, nil
}

type userRepository struct {
	goddessv1.UserServer
}
