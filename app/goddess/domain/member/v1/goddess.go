package memberv1

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
	RegisterMemberV1Factory(config.DomainConfig_GORM, NewDefaultMember)
}

func NewDefaultMember(c *config.DomainConfig) (goddessv1.MemberServer, func() error, error) {
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
	memberRepo := impl.NewMemberRepositoryWithDB(db)
	userRepo := impl.NewUserRepositoryWithDB(db)
	namespaceRepo := impl.NewNamespaceRepositoryWithDB(db)
	helper := klog.NewHelper(klog.With(klog.GetLogger(), "module", "member"))
	memberBiz := biz.NewMember(memberRepo, userRepo, namespaceRepo, helper)
	return &defaultMember{
		MemberServer: service.NewMemberService(memberBiz),
	}, close, nil
}

type defaultMember struct {
	goddessv1.MemberServer
}
