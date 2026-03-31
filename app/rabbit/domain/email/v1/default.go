package emailv1

import (
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	emaildomain "github.com/aide-family/rabbit/domain/email"
	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl"
	"github.com/aide-family/rabbit/internal/service"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	emaildomain.RegisterEmailV1Factory(config.DomainConfig_DEFAULT, NewDefaultEmail)
}

// NewDefaultEmail creates an in-process email server (DEFAULT driver).
func NewDefaultEmail(c *config.DomainConfig, driver *anypb.Any) (apiv1.EmailServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal default config failed: %v", err)
		}
	}

	bootstrap := &conf.Bootstrap{
		Jwt:      defaultConfig.GetJwt(),
		Database: defaultConfig.GetDatabase(),
		JobCore: &conf.JobCore{
			WorkerTotal: 1,
			Timeout:     durationpb.New(10 * time.Second),
			BufferSize:  1,
			MaxRetry:    0,
		},
		JobClusters: &config.ClusterConfig{},
	}

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.email.v1"))
	d, closeData, err := data.New(bootstrap, helper)
	if err != nil {
		return nil, nil, err
	}

	emailConfigRepo := impl.NewEmailConfigRepository(d)
	emailConfigBiz := biz.NewEmailConfig(emailConfigRepo, helper)
	srv := &defaultEmail{EmailServer: service.NewEmailService(emailConfigBiz)}

	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultEmail struct {
	apiv1.EmailServer
}
