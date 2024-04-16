package options

import (
	"flag"
	"os"
	"strings"

	clusterv1beta1 "github.com/aide-family/moon/api/cluster/v1beta1"
	"github.com/aide-family/moon/app/kubemoon/cmd/apps/config"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clusterv1beta1.AddToScheme(scheme))
}

// TODO: @sumengzs@gmail.com
type Options struct {
	Config   string
	ConfPath string
}

func NewOptions() *Options {
	return &Options{
		Config:   "./configs/config.yaml",
		ConfPath: "./configs",
	}
}

func (s *Options) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("moon")
	fs.StringVar(&s.Config, "config", s.Config, "service configuration file path")
	fs.StringVar(&s.ConfPath, "conf", s.ConfPath, "service configuration file path")
	logFs := fss.FlagSet("log")
	local := flag.NewFlagSet("log", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(f *flag.Flag) {
		f.Name = strings.Replace(f.Name, "_", "-", -1)
		logFs.AddGoFlag(f)
	})
	return fss
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) Complete() (*config.Config, error) {
	kubeConfig := ctrl.GetConfigOrDie()
	componentConfiguration, err := LoadComponentConfiguration(s.Config)
	if err != nil {
		return nil, err
	}
	return &config.Config{
		Scheme:          scheme,
		KubeConfig:      kubeConfig,
		ComponentConfig: componentConfiguration,
	}, nil
}

func LoadComponentConfiguration(path string) (*config.ComponentConfiguration, error) {
	componentConfiguration := new(config.ComponentConfiguration)
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, componentConfiguration)
	if err != nil {
		return nil, err
	}
	return componentConfiguration, nil
}
