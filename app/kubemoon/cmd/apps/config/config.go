package config

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Config struct {
	Scheme     *runtime.Scheme
	KubeConfig *rest.Config
	// TODO: complete config
}

func (c *Config) ManagerOptions() ctrl.Options {
	// TODO: complete manager options
	return ctrl.Options{}
}
