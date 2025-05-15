package plugin

import (
	"fmt"
	"plugin"

	"github.com/go-kratos/kratos/v2/log"
)

type LoadConfig struct {
	Path    string
	Logger  log.Logger
	Configs []byte
}

// NewFunc is the function signature plugins must implement
type NewFunc[T any] func(config *LoadConfig) (T, error)

func Load[T any](config *LoadConfig) (res T, err error) {
	p, err := plugin.Open(config.Path)
	if err != nil {
		err = fmt.Errorf("could not load plugin: %v", err)
		return
	}

	newFuncSym, err := p.Lookup("New")
	if err != nil {
		err = fmt.Errorf("could not find New symbol: %v", err)
		return
	}

	newFunc, ok := newFuncSym.(func(config *LoadConfig) (T, error))
	if !ok {
		err = fmt.Errorf("plugin New has wrong signature, %T", newFuncSym)
		return
	}

	return newFunc(config)
}
