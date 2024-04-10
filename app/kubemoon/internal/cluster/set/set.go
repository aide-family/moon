/*
Copyright 2023 The Multi Cluster Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterset

import (
	"context"
	"fmt"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

const (
	UserAgentName = "cluster-set"
)

type Set struct {
	mu     sync.RWMutex
	client client.Client
	cm     map[string]clu.Client
}

func New(config *rest.Config, scheme *runtime.Scheme) (clu.Set, error) {
	cm := make(map[string]clu.Client)
	config = rest.AddUserAgent(config, UserAgentName)

	cli, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return &Set{
		client: cli,
		cm:     cm,
	}, nil
}

func (p *Set) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, cli := range p.cm {
		if err := cli.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Set) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, cli := range p.cm {
		cli.Stop()
	}
}

func (p *Set) Add(cli clu.Client) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if oldC, ok := p.cm[cli.Name()]; ok {
		switch oldC.Status() {
		case clu.Started, clu.Ready, clu.Waiting:
			return fmt.Errorf("%s, can not replace", cli)
		}
	}
	p.cm[cli.Name()] = cli
	return nil
}

func (p *Set) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.cm[name]; ok {
		p.cm[name].Stop()
		delete(p.cm, name)
	}
}

func (p *Set) Cluster(name string) clu.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cm[name]
}

func (p *Set) Clusters() map[string]clu.Client {
	cm := make(map[string]clu.Client, len(p.cm))
	p.mu.RLock()
	defer p.mu.RUnlock()
	for name, cli := range p.cm {
		cm[name] = cli
	}
	return cm
}
