package namespacev1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/data/impl"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	RegisterNamespaceV1Factory(config.DomainConfig_DEFAULT, NewDefaultNamespace)
}

func NewDefaultNamespace(c *config.DomainConfig) (goddessv1.NamespaceServer, func() error, error) {
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
	namespaceRepo := impl.NewNamespaceRepositoryWithDB(db)
	userRepo := impl.NewUserRepositoryWithDB(db)
	memberRepo := impl.NewMemberRepositoryWithDB(db)
	emailRepo := impl.NewEmailRepository(bootstrap)
	helper := klog.NewHelper(klog.With(klog.GetLogger(), "module", "namespace"))
	userBiz := biz.NewUser(userRepo, helper)
	memberBiz := biz.NewMember(bootstrap, transaction, memberRepo, userRepo, namespaceRepo, emailRepo, helper)
	namespaceBiz := biz.NewNamespace(transaction, namespaceRepo, userBiz, memberBiz, helper)
	return &defaultNamespace{
		NamespaceServer: service.NewNamespaceService(namespaceBiz),
	}, close, nil
}

type defaultNamespace struct {
	goddessv1.NamespaceServer
}
