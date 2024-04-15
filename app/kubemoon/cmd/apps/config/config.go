package config

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Config struct {
	Scheme          *runtime.Scheme
	KubeConfig      *rest.Config
	ComponentConfig *ComponentConfiguration
}

type ComponentConfiguration struct {
	Manager *ManagerOptions
}

type ManagerOptions struct {
	// LeaderElection determines whether or not to use leader election when
	// starting the manager.
	LeaderElection bool `yaml:"leaderElection"`
	// LeaderElectionID determines the name of the resource that leader election
	// will use for holding the leader lock.
	LeaderElectionID string `yaml:"leaderElectionID"`
	// LeaderElectionNamespace determines the namespace in which the leader
	// election resource will be created.
	LeaderElectionNamespace string `yaml:"leaderElectionNamespace"`
	// HealthProbeBindAddress is the TCP address that the controller should bind to
	// for serving health probes
	// It can be set to "0" or "" to disable serving the health probe.
	HealthProbeBindAddress string `yaml:"healthProbeBindAddress"`
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
