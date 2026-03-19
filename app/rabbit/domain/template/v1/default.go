package templatev1

import (
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl"
	"github.com/aide-family/rabbit/internal/service"
	templatedomain "github.com/aide-family/rabbit/domain/template"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	templatedomain.RegisterTemplateV1Factory(config.DomainConfig_DEFAULT, NewDefaultTemplate)
}

// NewDefaultTemplate creates an in-process template server (DEFAULT driver).
func NewDefaultTemplate(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
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

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.template.v1"))
	d, closeData, err := data.New(bootstrap, helper)
	if err != nil {
		return nil, nil, err
	}

	templateRepo := impl.NewTemplateRepository(d)
	templateBiz := biz.NewTemplate(templateRepo, helper)
	srv := &defaultTemplate{TemplateServer: service.NewTemplateService(templateBiz)}

	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultTemplate struct {
	apiv1.TemplateServer
}
