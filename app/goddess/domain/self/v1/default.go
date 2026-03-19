package selfv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	selfdomain "github.com/aide-family/goddess/domain/self"
	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/data/impl"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	selfdomain.RegisterSelfFactoryV1(config.DomainConfig_DEFAULT, NewDefaultSelf)
}

// NewDefaultSelf creates an in-process self server (DEFAULT driver).
func NewDefaultSelf(c *config.DomainConfig) (goddessv1.SelfServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
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
	helper := klog.NewHelper(klog.With(klog.GetLogger(), "module", "self"))
	userRepo := impl.NewUserRepositoryWithDB(db)
	memberRepo := impl.NewMemberRepositoryWithDB(db)
	namespaceRepo := impl.NewNamespaceRepositoryWithDB(db)
	loginRepo := impl.NewLoginRepositoryWithDB(db, bootstrap.GetJwt())
	emailRepo := impl.NewEmailRepository(bootstrap)
	userBiz := biz.NewUser(userRepo, helper)
	memberBiz := biz.NewMember(bootstrap, transaction, memberRepo, userRepo, namespaceRepo, emailRepo, helper)
	namespaceBiz := biz.NewNamespace(transaction, namespaceRepo, userBiz, memberBiz, helper)
	loginBiz := biz.NewLoginBiz(transaction, loginRepo)
	return &defaultSelf{SelfServer: service.NewSelfService(userBiz, memberBiz, namespaceBiz, loginBiz)}, close, nil
}

type defaultSelf struct {
	goddessv1.SelfServer
}
