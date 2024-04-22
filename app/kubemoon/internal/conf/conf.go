package conf

import (
	clusterv1beta1 "github.com/aide-family/moon/api/cluster/v1beta1"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aide-family/moon/pkg/helper/plog"
)

// ProviderSetConf is conf providers.
var ProviderSetConf = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "Server"),
	wire.FieldsOf(new(*Bootstrap), "Data"),
	wire.FieldsOf(new(*Bootstrap), "Env"),
	wire.FieldsOf(new(*Bootstrap), "Log"),
	wire.FieldsOf(new(*Bootstrap), "ApiWhite"),
	wire.Bind(new(plog.Config), new(*Log)),
	LoadConfig,
)

type Before func(bc *Bootstrap) error

func LoadConfig(flagConf *string, before Before) (*Bootstrap, error) {
	if flagConf == nil || *flagConf == "" {
		return nil, errors.NotFound("FLAG_CONFIGS", "config path not found")
	}
	c := config.New(
		config.WithSource(
			file.NewSource(*flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, before(&bc)
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clusterv1beta1.AddToScheme(scheme))
}

type Config struct {
	Scheme          *runtime.Scheme
	KubeConfig      *rest.Config
	ComponentConfig *ComponentConfig
}

func (c *Config) ManagerOptions() ctrl.Options {
	if c != nil && c.ComponentConfig != nil && c.ComponentConfig.Manager != nil {
		return ctrl.Options{
			Scheme:                  c.Scheme,
			LeaderElection:          c.ComponentConfig.Manager.LeaderElection,
			LeaderElectionID:        c.ComponentConfig.Manager.LeaderElectionID,
			LeaderElectionNamespace: c.ComponentConfig.Manager.LeaderElectionNamespace,
			HealthProbeBindAddress:  c.ComponentConfig.Manager.HealthProbeBindAddress,
		}
	}
	return ctrl.Options{}
}

// Validate .
func (s *Bootstrap) Validate() []error {
	var errs []error
	return errs
}

func (s *Bootstrap) Complete() (*Config, error) {
	kubeConfig := ctrl.GetConfigOrDie()
	componentConfiguration := s.GetComponentConfig()

	return &Config{
		Scheme:          scheme,
		KubeConfig:      kubeConfig,
		ComponentConfig: componentConfiguration,
	}, nil
}
